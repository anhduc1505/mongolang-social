package model

// User represents user collection from the database
type User struct {
	BaseModel
	FirstName    string `bson:"first_name" json:"first_name"`
	LastName     string `bson:"last_name" json:"last_name"`
	Email        string `bson:"email" json:"email"`
	Password     string `bson:"password" json:"password"`
	Pseudonym    string `bson:"pseudonym" json:"pseudonym"`
	ProfileImage string `bson:"profile_image" json:"profile_image"`
	Biography    string `bson:"biography" json:"biography"`
	IsVerified   bool   `bson:"is_verified" json:"is_verified"`
}
