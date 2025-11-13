package tag

import (
	"context"
	"errors"
	"golang-project/internal/model"
	repo "golang-project/internal/repository"
	"golang-project/static"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type repository struct {
	collection        *mongo.Collection
	postTagCollection *mongo.Collection
}

// NewRepository returns a new implementation of repository.Tag
func NewRepository(db *mongo.Database) repo.Tag {
	return &repository{
		collection:        db.Collection(static.CollectionTags),
		postTagCollection: db.Collection(static.CollectionPostTags),
	}
}

func (r *repository) Insert(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	now := time.Now()
	tag.CreatedAt = &now
	tag.UpdatedAt = &now

	result, err := r.collection.InsertOne(ctx, tag)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		tag.ID = oid
	}

	return tag, nil
}

// Read finds and returns the tag model by ID
func (r *repository) Read(ctx context.Context, id primitive.ObjectID) (*model.Tag, error) {
	var result model.Tag
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}
	return &result, nil
}

// Select finds and returns all tags
func (r *repository) Select(ctx context.Context, tagIDs []primitive.ObjectID) ([]*model.Tag, error) {
	filter := bson.M{}
	if len(tagIDs) > 0 {
		filter["_id"] = bson.M{"$in": tagIDs}
	}

	cursor, err := r.collection.Find(ctx, filter)
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

// Delete removes a tag from the database
func (r *repository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("tag not found")
	}
	return nil
}

// HasPosts checks if a tag has associated posts
func (r *repository) HasPosts(ctx context.Context, tagID primitive.ObjectID) (bool, error) {
	count, err := r.postTagCollection.CountDocuments(ctx, bson.M{"tag_id": tagID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
