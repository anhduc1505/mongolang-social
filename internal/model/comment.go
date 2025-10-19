package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents comment collection from the database
type Comment struct {
	BaseModel
	Content         string              `bson:"content" json:"content"`
	PostID          primitive.ObjectID  `bson:"post_id" json:"post_id"`
	UserID          primitive.ObjectID  `bson:"user_id" json:"user_id"`
	ParentCommentID *primitive.ObjectID `bson:"parent_comment_id,omitempty" json:"parent_comment_id,omitempty"`
}
