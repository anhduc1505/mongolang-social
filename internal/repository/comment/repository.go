package comment

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	ct "golang-project/internal/contract"
	"golang-project/internal/model"
	repo "golang-project/internal/repository"
	"golang-project/static"
	"golang-project/util/pagination"
)

// repository represents the implementation of repository.Comment
type repository struct {
	commentCollection *mongo.Collection
	userCollection    *mongo.Collection
	postCollection    *mongo.Collection
}

// NewRepository returns a new implementation of repository.Comment
func NewRepository(db *mongo.Database) repo.Comment {
	return &repository{
		commentCollection: db.Collection(static.CollectionComments),
		userCollection:    db.Collection(static.CollectionUsers),
		postCollection:    db.Collection(static.CollectionPosts),
	}
}

// Select retrieves parent comments for a post with pagination
func (r *repository) Select(ctx context.Context, request *ct.ListCommentRequest) ([]*model.Comment, int64, error) {
	// Count total parent comments
	filter := bson.M{
		"post_id":           request.PostID,
		"parent_comment_id": nil,
		"deleted_at":        nil,
	}

	total, err := r.commentCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	skip := int64(pagination.CalculateOffset(request.Page, request.PageSize))
	limit := int64(request.PageSize)

	// Find parent comments with pagination
	cursor, err := r.commentCollection.Find(ctx, filter,
		options.Find().
			SetSort(bson.M{"created_at": -1}).
			SetSkip(skip).
			SetLimit(limit))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, err
	}

	// Load User, Post, and ChildComments for each comment
	for _, comment := range comments {
		// Load User
		if !comment.UserID.IsZero() {
			var user model.User
			if err := r.userCollection.FindOne(ctx, bson.M{"_id": comment.UserID}).Decode(&user); err == nil {
				comment.User = &user
			}
		}

		// Load Post with User and Tags
		if !comment.PostID.IsZero() {
			post, err := r.loadPostWithRelations(ctx, comment.PostID)
			if err == nil {
				comment.Post = post
			}
		}

		// Load ChildComments
		childComments, err := r.loadChildComments(ctx, comment.ID)
		if err == nil {
			comment.ChildComments = childComments
		}
	}

	return comments, total, nil
}

// Read finds and returns the comment model by id with related data
func (r *repository) Read(ctx context.Context, id primitive.ObjectID) (*model.Comment, error) {
	var comment model.Comment
	err := r.commentCollection.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&comment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	// Load User
	if !comment.UserID.IsZero() {
		var user model.User
		if err := r.userCollection.FindOne(ctx, bson.M{"_id": comment.UserID}).Decode(&user); err == nil {
			comment.User = &user
		}
	}

	// Load Post with User and Tags
	if !comment.PostID.IsZero() {
		post, err := r.loadPostWithRelations(ctx, comment.PostID)
		if err == nil {
			comment.Post = post
		}
	}

	// Load ChildComments
	childComments, err := r.loadChildComments(ctx, comment.ID)
	if err == nil {
		comment.ChildComments = childComments
	}

	return &comment, nil
}

// Insert creates a new comment in the database
func (r *repository) Insert(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	now := time.Now()
	comment.CreatedAt = &now
	comment.UpdatedAt = &now

	result, err := r.commentCollection.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		comment.ID = oid
	}

	return comment, nil
}

// UpdateCommentByID updates an existing comment in the database
func (r *repository) UpdateCommentByID(ctx context.Context, commentID primitive.ObjectID, updates map[string]interface{}) error {
	now := time.Now()
	updates["updated_at"] = now

	_, err := r.commentCollection.UpdateOne(
		ctx,
		bson.M{"_id": commentID, "deleted_at": nil},
		bson.M{"$set": updates},
	)
	return err
}

// Delete soft deletes the comment and all its child comments
func (r *repository) Delete(ctx context.Context, commentID primitive.ObjectID) error {
	now := time.Now()

	// Use a session for transaction-like behavior
	session, err := r.commentCollection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		// Soft delete all child comments
		_, err := r.commentCollection.UpdateMany(
			sc,
			bson.M{"parent_comment_id": commentID, "deleted_at": nil},
			bson.M{"$set": bson.M{"deleted_at": now}},
		)
		if err != nil {
			return nil, err
		}

		// Soft delete the parent comment
		_, err = r.commentCollection.UpdateOne(
			sc,
			bson.M{"_id": commentID, "deleted_at": nil},
			bson.M{"$set": bson.M{"deleted_at": now}},
		)
		return nil, err
	})

	return err
}

// loadPostWithRelations loads a post with its User and Tags
func (r *repository) loadPostWithRelations(ctx context.Context, postID primitive.ObjectID) (*model.Post, error) {
	var post model.Post
	err := r.postCollection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
	if err != nil {
		return nil, err
	}

	// Load User
	if !post.UserID.IsZero() {
		var user model.User
		if err := r.userCollection.FindOne(ctx, bson.M{"_id": post.UserID}).Decode(&user); err == nil {
			post.User = &user
		}
	}

	// Load Tags using aggregation pipeline
	pipeline := []bson.M{
		{"$match": bson.M{"post_id": postID}},
		{
			"$lookup": bson.M{
				"from":         static.CollectionTags,
				"localField":   "tag_id",
				"foreignField": "_id",
				"as":           "tag",
			},
		},
		{"$unwind": "$tag"},
		{"$replaceRoot": bson.M{"newRoot": "$tag"}},
	}

	cursor, err := r.commentCollection.Database().Collection(static.CollectionPostTags).Aggregate(ctx, pipeline)
	if err == nil {
		defer cursor.Close(ctx)
		var tags []*model.Tag
		if err := cursor.All(ctx, &tags); err == nil {
			post.Tags = tags
		}
	}

	return &post, nil
}

// loadChildComments loads all child comments for a parent comment
func (r *repository) loadChildComments(ctx context.Context, parentID primitive.ObjectID) ([]*model.Comment, error) {
	cursor, err := r.commentCollection.Find(ctx, bson.M{
		"parent_comment_id": parentID,
		"deleted_at":        nil,
	}, options.Find().SetSort(bson.M{"created_at": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var childComments []*model.Comment
	if err := cursor.All(ctx, &childComments); err != nil {
		return nil, err
	}

	// Load User for each child comment
	for _, child := range childComments {
		if !child.UserID.IsZero() {
			var user model.User
			if err := r.userCollection.FindOne(ctx, bson.M{"_id": child.UserID}).Decode(&user); err == nil {
				child.User = &user
			}
		}
	}

	return childComments, nil
}
