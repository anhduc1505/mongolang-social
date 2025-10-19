package service

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	ct "golang-project/internal/contract"
)

// Authentication represents the service logic of Authentication
type Authentication interface {
	SignIn(*ct.SignInRequest) (*ct.SignInResponse, error)
	SignUp(*ct.SignUpRequest) (*ct.SignUpResponse, error)
}

// Profile represents the service logic of Profile
type Profile interface {
	GetByID(primitive.ObjectID) (*ct.ProfileResponse, error)
	GetPost(primitive.ObjectID, primitive.ObjectID) (*ct.PostResponse, error)
	Update(primitive.ObjectID, *ct.UpdateProfileRequest) (*ct.ProfileResponse, error)
	ChangePassword(primitive.ObjectID, *ct.ChangePasswordRequest) (*ct.ChangePasswordResponse, error)
	ListBloggerPosts(id primitive.ObjectID, isPublishedFilter string) (*ct.ListPostResponse, error)
}

type Tag interface {
	Create(string) (*ct.TagResponse, error)
	Delete(primitive.ObjectID) error
	List() (*ct.ListTagResponse, error)
	ListPosts(primitive.ObjectID) (*ct.ListPostResponse, error)
}

type Comment interface {
	List(*ct.ListCommentRequest) (*ct.ListCommentResponse, error)
	Create(*ct.CreateCommentRequest, primitive.ObjectID) (*ct.CommentResponse, error)
	Update(*ct.UpdateCommentRequest, primitive.ObjectID) (*ct.CommentResponse, error)
	Delete(primitive.ObjectID, primitive.ObjectID) error
}

type Post interface {
	GetByID(primitive.ObjectID) (*ct.PostResponse, error)
	List(*ct.ListPostRequest) (*ct.ListPostResponse, error)
	Create(*ct.CreatePostRequest, primitive.ObjectID) (*ct.PostResponse, error)
	Update(primitive.ObjectID, *ct.UpdatePostRequest) (*ct.PostResponse, error)
	Delete(primitive.ObjectID, primitive.ObjectID) error
}

// Favourite represents the service logic of Favourite features
type Favourite interface {
	// User following operations
	UpdateFollowStatus(userID primitive.ObjectID, req *ct.BloggerFollowRequest) (*ct.BloggerFollowStatusResponse, error)
	ListFollowingUsers(userID primitive.ObjectID) (*ct.ListProfileResponse, error)
	ListUserPosts(userID primitive.ObjectID) (*ct.ListPostResponse, error)
	// Post favorite operations
	UpdateFavouriteStatus(userID primitive.ObjectID, req *ct.PostFavouriteRequest) (*ct.PostFavouriteStatusResponse, error)
	ListFavouritePosts(userID primitive.ObjectID) (*ct.ListPostResponse, error)
}
