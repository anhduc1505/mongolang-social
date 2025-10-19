package user

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"golang-project/database"
	"golang-project/internal/model"
	repo "golang-project/internal/repository"
	"golang-project/static"
)

// repository represents the implementation of repository.User
type repository struct {
	collection *mongo.Collection
}

// NewRepository returns a new implementation of repository.User
func NewRepository() repo.User {
	conn, err := database.NewConnectionFromEnv()
	if err != nil {
		panic("Failed to create database connection: " + err.Error())
	}

	_, err = conn.Connect()
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	db := conn.GetDatabase()
	collection := db.Collection(static.CollectionUsers)

	return &repository{collection: collection}
}

// Read finds and returns the user model by ID
func (r *repository) Read(id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &result, nil
}

// ReadByEmail finds and returns the user model by email
func (r *repository) ReadByEmail(email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &result, nil
}

// Insert performs insert action into user collection
func (r *repository) Insert(o *model.User) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate ObjectID if not set
	if o.ID.IsZero() {
		o.ID = primitive.NewObjectID()
	}

	// Set timestamps
	now := time.Now()
	o.CreatedAt = &now
	o.UpdatedAt = &now

	_, err := r.collection.InsertOne(ctx, o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

// Update performs update action into user collection
func (r *repository) Update(o *model.User, updates map[string]interface{}) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add updated timestamp
	updates["updated_at"] = time.Now()

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": o.ID}, bson.M{"$set": updates})
	if err != nil {
		return nil, err
	}

	// Return updated user
	return r.Read(o.ID)
}

func (r *repository) ReadOwnPosts(id primitive.ObjectID, isPublishedFilter *bool) ([]*model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{"user_id": id}
	if isPublishedFilter != nil {
		filter["is_published"] = *isPublishedFilter
	}

	cursor, err := r.collection.Database().Collection(static.CollectionPosts).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*model.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}
