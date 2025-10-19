package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FollowUser represents follow_user collection from the database
type FollowUser struct {
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	FollowUserID primitive.ObjectID `bson:"follow_user_id" json:"follow_user_id"`
}

// FavoritePost represents favorite_post collection from the database
type FavoritePost struct {
	PostID primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
}
