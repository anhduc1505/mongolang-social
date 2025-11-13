package favourite

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"golang-project/internal/model"
	repo "golang-project/internal/repository"
	"golang-project/static"
)

// repository represents the implementation of repository.Favourite
type repository struct {
	followCollection   *mongo.Collection
	favoriteCollection *mongo.Collection
	userCollection     *mongo.Collection
	postCollection     *mongo.Collection
	postTagCollection  *mongo.Collection
	tagCollection      *mongo.Collection
}

// NewRepository returns a new implementation of repository.Favourite
func NewRepository(db *mongo.Database) repo.Favourite {
	return &repository{
		followCollection:   db.Collection(static.CollectionFollows),
		favoriteCollection: db.Collection(static.CollectionFavorites),
		userCollection:     db.Collection(static.CollectionUsers),
		postCollection:     db.Collection(static.CollectionPosts),
		postTagCollection:  db.Collection(static.CollectionPostTags),
		tagCollection:      db.Collection(static.CollectionTags),
	}
}

// IsFollowing checks if user is following followUser
func (r *repository) IsFollowing(userID, followUserID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.followCollection.CountDocuments(ctx, bson.M{
		"user_id":        userID,
		"follow_user_id": followUserID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SelectFollowing returns all users that the given user is following
func (r *repository) SelectFollowing(userID primitive.ObjectID) ([]*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use aggregation to join follows with users
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{
			"$lookup": bson.M{
				"from":         static.CollectionUsers,
				"localField":   "follow_user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{"$unwind": "$user"},
		{"$replaceRoot": bson.M{"newRoot": "$user"}},
	}

	cursor, err := r.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Follow adds a follow relationship between user and followUser
func (r *repository) Follow(follow *model.FollowUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if already following
	count, err := r.followCollection.CountDocuments(ctx, bson.M{
		"user_id":        follow.UserID,
		"follow_user_id": follow.FollowUserID,
	})
	if err != nil {
		return err
	}
	if count > 0 {
		// Already following, return nil (idempotent)
		return nil
	}

	_, err = r.followCollection.InsertOne(ctx, follow)
	return err
}

// Unfollow removes a follow relationship between user and followUser
func (r *repository) Unfollow(userID, followUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.followCollection.DeleteOne(ctx, bson.M{
		"user_id":        userID,
		"follow_user_id": followUserID,
	})
	return err
}

// SelectFollowingUsersPosts returns all published posts from users that the given user is following
func (r *repository) SelectFollowingUsersPosts(userID primitive.ObjectID) ([]*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use aggregation to join follows with posts
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{
			"$lookup": bson.M{
				"from":         static.CollectionPosts,
				"localField":   "follow_user_id",
				"foreignField": "user_id",
				"as":           "posts",
			},
		},
		{"$unwind": "$posts"},
		{"$match": bson.M{"posts.is_published": true}},
		{"$replaceRoot": bson.M{"newRoot": "$posts"}},
		{"$sort": bson.M{"created_at": -1}},
	}

	cursor, err := r.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*model.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Load User and Tags for each post
	for i := range posts {
		// Load User
		if !posts[i].UserID.IsZero() {
			var user model.User
			if err := r.userCollection.FindOne(ctx, bson.M{"_id": posts[i].UserID}).Decode(&user); err == nil {
				posts[i].User = &user
			}
		}

		// Load Tags using aggregation
		tagPipeline := []bson.M{
			{"$match": bson.M{"post_id": posts[i].ID}},
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

		tagCursor, err := r.postTagCollection.Aggregate(ctx, tagPipeline)
		if err == nil {
			defer tagCursor.Close(ctx)
			var tags []*model.Tag
			if err := tagCursor.All(ctx, &tags); err == nil {
				posts[i].Tags = tags
			}
		}
	}

	return posts, nil
}

// SelectFavouritePosts returns all posts that the given user has marked as favourite
func (r *repository) SelectFavouritePosts(userID primitive.ObjectID) ([]*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use aggregation to join favorites with posts
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{
			"$lookup": bson.M{
				"from":         static.CollectionPosts,
				"localField":   "post_id",
				"foreignField": "_id",
				"as":           "post",
			},
		},
		{"$unwind": "$post"},
		{"$match": bson.M{"post.is_published": true}},
		{"$replaceRoot": bson.M{"newRoot": "$post"}},
		{"$sort": bson.M{"created_at": -1}},
	}

	cursor, err := r.favoriteCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*model.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Load User and Tags for each post
	for i := range posts {
		// Load User
		if !posts[i].UserID.IsZero() {
			var user model.User
			if err := r.userCollection.FindOne(ctx, bson.M{"_id": posts[i].UserID}).Decode(&user); err == nil {
				posts[i].User = &user
			}
		}

		// Load Tags using aggregation
		tagPipeline := []bson.M{
			{"$match": bson.M{"post_id": posts[i].ID}},
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

		tagCursor, err := r.postTagCollection.Aggregate(ctx, tagPipeline)
		if err == nil {
			defer tagCursor.Close(ctx)
			var tags []*model.Tag
			if err := tagCursor.All(ctx, &tags); err == nil {
				posts[i].Tags = tags
			}
		}
	}

	return posts, nil
}

// IsFavourite checks if a post is marked as favourite by the user
func (r *repository) IsFavourite(userID, postID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.favoriteCollection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"post_id": postID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Favourite marks a post as favourite for a user
func (r *repository) Favourite(favourite *model.FavoritePost) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if already favorited
	count, err := r.favoriteCollection.CountDocuments(ctx, bson.M{
		"user_id": favourite.UserID,
		"post_id": favourite.PostID,
	})
	if err != nil {
		return err
	}
	if count > 0 {
		// Already favorited, return nil (idempotent)
		return nil
	}

	_, err = r.favoriteCollection.InsertOne(ctx, favourite)
	return err
}

// Unfavourite removes a post from a user's favourites
func (r *repository) Unfavourite(userID, postID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.favoriteCollection.DeleteOne(ctx, bson.M{
		"user_id": userID,
		"post_id": postID,
	})
	return err
}
