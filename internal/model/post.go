package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents post collection from the database
type Post struct {
	BaseModel
	Title       string               `bson:"title" json:"title"`
	Body        string               `bson:"body" json:"body"`
	Slug        string               `bson:"slug" json:"slug"`
	IsPublished bool                 `bson:"is_published" json:"is_published"`
	UserID      primitive.ObjectID   `bson:"user_id" json:"user_id"`
	TagIDs      []primitive.ObjectID `bson:"tag_ids" json:"tag_ids"`
}
