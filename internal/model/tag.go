package model

// Tag represents tag collection from the database
type Tag struct {
	BaseModel
	Name string `bson:"name" json:"name"`
}
