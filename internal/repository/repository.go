package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"golang-project/internal/contract"
	"golang-project/internal/model"
)

// User represents the repository actions to the user collection
type User interface {
	Read(primitive.ObjectID) (*model.User, error)
	Insert(*model.User) (*model.User, error)
	Update(*model.User, map[string]interface{}) (*model.User, error)
	ReadByEmail(string) (*model.User, error)
	ReadOwnPosts(id primitive.ObjectID, isPublishedFilter *bool) ([]*model.Post, error)
}

type Tag interface {
	Insert(*model.Tag) error
	Read(primitive.ObjectID) (*model.Tag, error)
	Delete(primitive.ObjectID) error
	HasPosts(primitive.ObjectID) (bool, error)
	Select([]primitive.ObjectID) ([]*model.Tag, error)
	SelectPost(primitive.ObjectID) ([]*model.Post, error)
	SelectPostTag([]primitive.ObjectID) ([]*model.PostTag, error)
	SelectUser([]primitive.ObjectID) ([]*model.User, error)
}

type Comment interface {
	Select(*contract.ListCommentRequest) ([]*model.Comment, int64, error)
	Insert(*model.Comment) (*model.Comment, error)
	Read(primitive.ObjectID) (*model.Comment, error)
	UpdateCommentByID(primitive.ObjectID, map[string]interface{}) error
	Delete(primitive.ObjectID) error
}

type Post interface {
	Read(primitive.ObjectID) (*model.Post, error)
	Insert(*model.Post) (*model.Post, error)
	AddPostTags(primitive.ObjectID, []primitive.ObjectID) error
	FindSlugsLike(string) ([]string, error)
	GetTags(primitive.ObjectID) ([]*model.Tag, error)
	ReadByCondition(map[string]interface{}, ...string) (*model.Post, error)
	Select(*contract.ListPostRequest) ([]*model.Post, error)
	UpdatePost(*model.Post, map[string]interface{}) error
	UpdatePostTag(*model.Post, []*model.Tag) error
	Delete(primitive.ObjectID) error
}

// Favourite represents the repository actions for managing user follows and post favorites
type Favourite interface {
	// User following operations
	IsFollowing(userID, followUserID primitive.ObjectID) (bool, error)
	SelectFollowing(userID primitive.ObjectID) ([]*model.User, error)
	Follow(*model.FollowUser) error
	Unfollow(userID, followUserID primitive.ObjectID) error
	SelectFollowingUsersPosts(userID primitive.ObjectID) ([]*model.Post, error)

	// Post favourite operations
	SelectFavouritePosts(userID primitive.ObjectID) ([]*model.Post, error)
	IsFavourite(userID, postID primitive.ObjectID) (bool, error)
	Favourite(*model.FavoritePost) error
	Unfavourite(userID, postID primitive.ObjectID) error
}
