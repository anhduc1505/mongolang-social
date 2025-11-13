package repository

import (
	"context"

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
	Insert(ctx context.Context, tag *model.Tag) (*model.Tag, error)
	Read(ctx context.Context, id primitive.ObjectID) (*model.Tag, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	HasPosts(ctx context.Context, tagID primitive.ObjectID) (bool, error)
	Select(ctx context.Context, tagIDs []primitive.ObjectID) ([]*model.Tag, error)
	//SelectPost(primitive.ObjectID) ([]*model.Post, error)
	//SelectPostTag([]primitive.ObjectID) ([]*model.PostTag, error)
	//SelectUser([]primitive.ObjectID) ([]*model.User, error)
}

type Comment interface {
	Select(ctx context.Context, req *contract.ListCommentRequest) ([]*model.Comment, int64, error)
	Insert(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	Read(ctx context.Context, id primitive.ObjectID) (*model.Comment, error)
	UpdateCommentByID(ctx context.Context, commentID primitive.ObjectID, updates map[string]interface{}) error
	Delete(ctx context.Context, commentID primitive.ObjectID) error
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
