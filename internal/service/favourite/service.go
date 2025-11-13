package favourite

import (
	"errors"
	ct "golang-project/internal/contract"
	"golang-project/internal/model"
	repo "golang-project/internal/repository"
	svc "golang-project/internal/service"
	"golang-project/static"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// service represents the implementation of service.Favourite
type service struct {
	favouriteRepo repo.Favourite
	userRepo      repo.User
	postRepo      repo.Post
}

// NewService returns a new implementation of service.Favourite
func NewService(favouriteRepo repo.Favourite, userRepo repo.User, postRepo repo.Post) svc.Favourite {
	return &service{
		favouriteRepo: favouriteRepo,
		userRepo:      userRepo,
		postRepo:      postRepo,
	}
}

// ListFollowingUsers returns a list of all bloggers that a user is following
func (s *service) ListFollowingUsers(userID primitive.ObjectID) (*ct.ListProfileResponse, error) {
	// Check if user exists
	_, err := s.userRepo.Read(userID)
	if err != nil {
		return nil, static.ErrUserNotFound
	}

	users, err := s.favouriteRepo.SelectFollowing(userID)
	if err != nil {
		return nil, err
	}

	return prepareListProfileResponse(users), nil
}

// UpdateFollowStatus handles follow/unfollow operations
func (s *service) UpdateFollowStatus(userID primitive.ObjectID, req *ct.BloggerFollowRequest) (*ct.BloggerFollowStatusResponse, error) {
	targetUserID := req.UserID

	// Check if target user exists
	_, err := s.userRepo.Read(targetUserID)
	if err != nil {
		return nil, static.ErrUserNotFound
	}

	// Prevent self-following
	if userID == targetUserID {
		return nil, static.ErrSelfFollow
	}

	// Check current follow status
	isFollowing, err := s.favouriteRepo.IsFollowing(userID, targetUserID)
	if err != nil {
		return nil, err
	}

	// Handle different actions
	switch req.Action {
	case static.Follow:
		if isFollowing {
			return &ct.BloggerFollowStatusResponse{
				UserID:      targetUserID,
				IsFollowing: true,
			}, nil
		}

		// Add to following list if not already there
		follow := &model.FollowUser{
			UserID:       userID,
			FollowUserID: targetUserID,
		}
		if err = s.favouriteRepo.Follow(follow); err != nil {
			return nil, err
		}

		return &ct.BloggerFollowStatusResponse{
			UserID:      targetUserID,
			IsFollowing: true,
		}, nil

	case static.Unfollow:
		if !isFollowing {
			return &ct.BloggerFollowStatusResponse{
				UserID:      targetUserID,
				IsFollowing: false,
			}, nil
		}

		// Remove from following list if it's there
		if err = s.favouriteRepo.Unfollow(userID, targetUserID); err != nil {
			return nil, err
		}

		return &ct.BloggerFollowStatusResponse{
			UserID:      targetUserID,
			IsFollowing: false,
		}, nil

	default:
		return nil, static.ErrUnsupportedFollowAction
	}
}

// ListUserPosts returns all posts from bloggers that a user is following
func (s *service) ListUserPosts(userID primitive.ObjectID) (*ct.ListPostResponse, error) {
	posts, err := s.favouriteRepo.SelectFollowingUsersPosts(userID)
	if err != nil {
		return nil, err
	}

	return prepareListPostResponse(posts), nil
}

// UpdateFavouriteStatus handles favorite/unfavorite operations
func (s *service) UpdateFavouriteStatus(userID primitive.ObjectID, req *ct.PostFavouriteRequest) (*ct.PostFavouriteStatusResponse, error) {
	postID := req.PostID

	// Check if post exists
	_, err := s.postRepo.Read(postID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, static.ErrPostNotFound
		}
		return nil, err
	}

	// Check current favourite status
	isAlreadyFavourited, err := s.favouriteRepo.IsFavourite(userID, postID)
	if err != nil {
		return nil, err
	}

	// Handle different actions
	switch req.Action {
	case static.Favourite:
		if isAlreadyFavourited {
			return &ct.PostFavouriteStatusResponse{
				PostID:      postID,
				IsFavourite: true,
			}, nil
		}

		// Add to favourites if not already there
		favourite := &model.FavoritePost{
			UserID: userID,
			PostID: postID,
		}
		if err = s.favouriteRepo.Favourite(favourite); err != nil {
			return nil, err
		}

		return &ct.PostFavouriteStatusResponse{
			PostID:      postID,
			IsFavourite: true,
		}, nil

	case static.Unfavourite:
		if !isAlreadyFavourited {
			return &ct.PostFavouriteStatusResponse{
				PostID:      postID,
				IsFavourite: false,
			}, nil
		}

		// Remove from favourites if it's there
		if err = s.favouriteRepo.Unfavourite(userID, postID); err != nil {
			return nil, err
		}

		return &ct.PostFavouriteStatusResponse{
			PostID:      postID,
			IsFavourite: false,
		}, nil

	default:
		return nil, static.ErrUnsupportedFavouriteAction
	}
}

// ListFavouritePosts returns all posts that a user has favorited
func (s *service) ListFavouritePosts(userID primitive.ObjectID) (*ct.ListPostResponse, error) {
	posts, err := s.favouriteRepo.SelectFavouritePosts(userID)
	if err != nil {
		return nil, err
	}

	return prepareListPostResponse(posts), nil
}
