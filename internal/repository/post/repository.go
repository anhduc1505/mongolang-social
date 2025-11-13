package post

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang-project/internal/contract"
	"golang-project/internal/model"
	repo "golang-project/internal/repository"
	"golang-project/static"
	"golang-project/util/pagination"
)

// repository represents the implementation of repository.Post
type repository struct {
	collection        *mongo.Collection
	userCollection    *mongo.Collection
	tagCollection     *mongo.Collection
	postTagCollection *mongo.Collection
}

// NewRepository returns a new implementation of repository.Post
func NewRepository(db *mongo.Database) repo.Post {
	return &repository{
		collection:        db.Collection(static.CollectionPosts),
		userCollection:    db.Collection(static.CollectionUsers),
		tagCollection:     db.Collection(static.CollectionTags),
		postTagCollection: db.Collection(static.CollectionPostTags),
	}
}

// Read finds and returns the post model by id
func (r *repository) Read(id primitive.ObjectID) (*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var post model.Post
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "is_published": true}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// Load user
	if !post.UserID.IsZero() {
		var user model.User
		err = r.userCollection.FindOne(ctx, bson.M{"_id": post.UserID}).Decode(&user)
		if err == nil {
			post.User = &user
		}
	}

	// Load tags
	tags, err := r.GetTags(id)
	if err == nil {
		post.Tags = tags
	}

	return &post, nil
}

// Insert creates a new post in the database
func (r *repository) Insert(post *model.Post) (*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	post.CreatedAt = &now
	post.UpdatedAt = &now

	result, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		post.ID = oid
	}

	return post, nil
}

// AddPostTags adds multiple tag associations to a post
func (r *repository) AddPostTags(postID primitive.ObjectID, tagIDs []primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove duplicate tag IDs
	uniq := make(map[primitive.ObjectID]struct{}, len(tagIDs))
	pivots := make([]interface{}, 0, len(tagIDs))

	for _, id := range tagIDs {
		if _, dup := uniq[id]; dup {
			continue
		}
		uniq[id] = struct{}{}
		pivots = append(pivots, bson.M{
			"post_id": postID,
			"tag_id":  id,
		})
	}

	// If no associations to create, return early
	if len(pivots) == 0 {
		return nil
	}

	// Create post tag associations
	_, err := r.postTagCollection.InsertMany(ctx, pivots)
	return err
}

// FindSlugsLike retrieves all slugs that start with the given baseSlug
func (r *repository) FindSlugsLike(baseSlug string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"slug": bson.M{
			"$regex":   fmt.Sprintf("^%s", baseSlug),
			"$options": "i",
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(bson.M{"slug": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var slugs []string
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		if slug, ok := doc["slug"].(string); ok {
			slugs = append(slugs, slug)
		}
	}

	return slugs, nil
}

// GetTags retrieves all tags associated with a post
func (r *repository) GetTags(postID primitive.ObjectID) ([]*model.Tag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get all tag IDs from post_tags collection
	cursor, err := r.postTagCollection.Find(ctx, bson.M{"post_id": postID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tagIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var postTag model.PostTag
		if err := cursor.Decode(&postTag); err != nil {
			continue
		}
		tagIDs = append(tagIDs, postTag.TagID)
	}

	if len(tagIDs) == 0 {
		return []*model.Tag{}, nil
	}

	// Then, get all tags by IDs
	cursor, err = r.tagCollection.Find(ctx, bson.M{"_id": bson.M{"$in": tagIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tags []*model.Tag
	if err = cursor.All(ctx, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// Select retrieves all posts from the database with optional filters
func (r *repository) Select(filter *contract.ListPostRequest) ([]*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	mongoFilter := bson.M{"is_published": true}

	if filter != nil {
		if filter.Title != "" {
			mongoFilter["title"] = bson.M{
				"$regex":   filter.Title,
				"$options": "i",
			}
		}
		if filter.Tag != "" {
			// Find tag by name first
			var tag model.Tag
			err := r.tagCollection.FindOne(ctx, bson.M{"name": filter.Tag}).Decode(&tag)
			if err == nil {
				// Find all post_tags with this tag_id
				cursor, err := r.postTagCollection.Find(ctx, bson.M{"tag_id": tag.ID})
				if err == nil {
					var postTags []model.PostTag
					cursor.All(ctx, &postTags)
					cursor.Close(ctx)

					// Extract post IDs
					postIDs := make([]primitive.ObjectID, 0, len(postTags))
					for _, pt := range postTags {
						postIDs = append(postIDs, pt.PostID)
					}
					if len(postIDs) > 0 {
						mongoFilter["_id"] = bson.M{"$in": postIDs}
					} else {
						// No posts with this tag
						return []*model.Post{}, nil
					}
				}
			}
		}
	}

	// Build find options
	findOptions := options.Find()
	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := pagination.CalculateOffset(filter.Page, filter.PageSize)
		findOptions.SetSkip(int64(offset))
		findOptions.SetLimit(int64(filter.PageSize))
	}

	// Find posts
	cursor, err := r.collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*model.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Load users and tags for each post
	for i := range posts {
		// Load user
		if !posts[i].UserID.IsZero() {
			var user model.User
			err = r.userCollection.FindOne(ctx, bson.M{"_id": posts[i].UserID}).Decode(&user)
			if err == nil {
				posts[i].User = &user
			}
		}

		// Load tags
		tags, err := r.GetTags(posts[i].ID)
		if err == nil {
			posts[i].Tags = tags
		}

		// Handle pseudonym filter
		if filter != nil && filter.Pseudonym != "" {
			if posts[i].User == nil || posts[i].User.Pseudonym != filter.Pseudonym {
				// Remove this post from results (we'll filter after)
				continue
			}
		}
	}

	// Filter by pseudonym if needed (after loading users)
	if filter != nil && filter.Pseudonym != "" {
		filteredPosts := make([]*model.Post, 0)
		for _, post := range posts {
			if post.User != nil && post.User.Pseudonym == filter.Pseudonym {
				filteredPosts = append(filteredPosts, post)
			}
		}
		return filteredPosts, nil
	}

	return posts, nil
}

// ReadByCondition finds a post based on provided conditions
func (r *repository) ReadByCondition(condition map[string]interface{}, preloads ...string) (*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert condition map to bson.M
	filter := bson.M{}
	for k, v := range condition {
		filter[k] = v
	}

	var post model.Post
	err := r.collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// Load user
	if !post.UserID.IsZero() {
		var user model.User
		err = r.userCollection.FindOne(ctx, bson.M{"_id": post.UserID}).Decode(&user)
		if err == nil {
			post.User = &user
		}
	}

	// Load tags
	tags, err := r.GetTags(post.ID)
	if err == nil {
		post.Tags = tags
	}

	return &post, nil
}

// UpdatePost updates the post model in the database
func (r *repository) UpdatePost(post *model.Post, updateMap map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateMap["updated_at"] = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": post.ID},
		bson.M{"$set": updateMap},
	)
	return err
}

// UpdatePostTag updates the post_tag associations
func (r *repository) UpdatePostTag(post *model.Post, tags []*model.Tag) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete existing associations
	_, err := r.postTagCollection.DeleteMany(ctx, bson.M{"post_id": post.ID})
	if err != nil {
		return err
	}

	// Create new associations
	if len(tags) > 0 {
		pivots := make([]interface{}, len(tags))
		for i, tag := range tags {
			pivots[i] = bson.M{
				"post_id": post.ID,
				"tag_id":  tag.ID,
			}
		}
		_, err = r.postTagCollection.InsertMany(ctx, pivots)
		if err != nil {
			return err
		}
	}

	// Update post tag_ids field
	tagIDs := make([]primitive.ObjectID, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": post.ID},
		bson.M{"$set": bson.M{"tag_ids": tagIDs}},
	)

	return err
}

// Delete deletes the post and related data
func (r *repository) Delete(postID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete post_tags
	_, err := r.postTagCollection.DeleteMany(ctx, bson.M{"post_id": postID})
	if err != nil {
		return err
	}

	// Delete comments (if comments collection exists)
	// Note: This assumes comments collection exists
	// commentsCollection := r.collection.Database().Collection(static.CollectionComments)
	// commentsCollection.DeleteMany(ctx, bson.M{"post_id": postID})

	// Delete post
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": postID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("post not found")
	}

	return nil
}
