package comment

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	ct "golang-project/internal/contract"
	repo "golang-project/internal/repository"
	svc "golang-project/internal/service"
	"golang-project/static"
)

// service represents the implementation of service.Comment
type service struct {
	commentRepo repo.Comment
	userRepo    repo.User
	postRepo    repo.Post
}

// NewService returns a new implementation of service.Comment
func NewService(commentRepo repo.Comment, userRepo repo.User, postRepo repo.Post) svc.Comment {
	return &service{
		commentRepo: commentRepo,
		userRepo:    userRepo,
		postRepo:    postRepo,
	}
}

// List executes the Comment list retrieval logic
func (s *service) List(req *ct.ListCommentRequest) (*ct.ListCommentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get comments from repository
	comments, total, err := s.commentRepo.Select(ctx, req)
	if err != nil {
		return nil, err
	}

	pagingResponse := ct.Paging{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    int(total),
	}

	return prepareListCommentResponse(comments, pagingResponse), nil
}

// Create creates a new comment
func (s *service) Create(request *ct.CreateCommentRequest, userID primitive.ObjectID) (*ct.CommentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get post info to validate post exists
	// Note: Post repository Read doesn't use context yet
	post, err := s.postRepo.Read(request.PostID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, static.ErrPostNotFound
		}
		return nil, err
	}

	// If this is a reply to another comment, validate the parent comment
	if request.ParentCommentID.ObjectID != nil {
		parentComment, err := s.commentRepo.Read(ctx, *request.ParentCommentID.ObjectID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, static.ErrCommentNotFound
			}
			return nil, err
		}

		// Check if parent comment belongs to the same post
		if parentComment.PostID != request.PostID {
			return nil, errors.New("parent comment does not belong to the same post")
		}

		// Check if parent comment is a root comment
		if parentComment.ParentCommentID != nil {
			return nil, errors.New("can only reply to root comments")
		}
	}

	// Create new comment model
	comment := prepareCommentModel(request, userID, post)

	// Save to database
	createdComment, err := s.commentRepo.Insert(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Load User for the created comment
	// Note: User repository Read doesn't use context yet
	user, err := s.userRepo.Read(userID)
	if err == nil {
		createdComment.User = user
	}

	// Set Post and ChildComments
	createdComment.Post = post
	// New comment has no children yet

	return prepareCommentResponse(createdComment), nil
}

// Update updates an existing comment
func (s *service) Update(req *ct.UpdateCommentRequest, userID primitive.ObjectID) (*ct.CommentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get comment by ID with preloaded data
	comment, err := s.commentRepo.Read(ctx, req.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, static.ErrCommentNotFound
		}
		return nil, err
	}

	// Check if user is the owner of the comment
	if comment.UserID != userID {
		return nil, static.ErrUserPermission
	}

	// Update comment
	updateMap := prepareUpdateComment(comment, req)
	updateCommentErr := s.commentRepo.UpdateCommentByID(ctx, req.ID, updateMap)
	if updateCommentErr != nil {
		return nil, updateCommentErr
	}

	// Read updated comment
	updatedComment, err := s.commentRepo.Read(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return prepareCommentResponse(updatedComment), nil
}

// Delete deletes a comment
func (s *service) Delete(commentID, ctxUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Read comment
	comment, err := s.commentRepo.Read(ctx, commentID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return static.ErrCommentNotFound
		}
		return err
	}

	// Check if user is the owner of the comment
	if comment.UserID != ctxUserID {
		return static.ErrUserPermission
	}

	return s.commentRepo.Delete(ctx, commentID)
}
