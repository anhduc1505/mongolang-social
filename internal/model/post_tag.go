package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostTag represents post_tag collection from the database
type PostTag struct {
	TagID  primitive.ObjectID `bson:"tag_id" json:"tag_id"`
	PostID primitive.ObjectID `bson:"post_id" json:"post_id"`
}
