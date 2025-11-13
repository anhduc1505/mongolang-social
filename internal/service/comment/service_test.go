package comment

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	ct "golang-project/internal/contract"
	"golang-project/internal/model"
	"golang-project/static"
)

// MockCommentRepository is a mock implementation of repository.Comment
type MockCommentRepository struct {
	comments  map[primitive.ObjectID]*model.Comment
	insertErr error
	readErr   error
	selectErr error
	updateErr error
	deleteErr error
}

func NewMockCommentRepository() *MockCommentRepository {
	return &MockCommentRepository{
		comments: make(map[primitive.ObjectID]*model.Comment),
	}
}

func (m *MockCommentRepository) Select(ctx context.Context, req *ct.ListCommentRequest) ([]*model.Comment, int64, error) {
	if m.selectErr != nil {
		return nil, 0, m.selectErr
	}
	var result []*model.Comment
	for _, comment := range m.comments {
		if comment.PostID == req.PostID && comment.ParentCommentID == nil {
			result = append(result, comment)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockCommentRepository) Insert(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	if comment.ID.IsZero() {
		comment.ID = primitive.NewObjectID()
	}
	now := time.Now()
	comment.CreatedAt = &now
	comment.UpdatedAt = &now
	m.comments[comment.ID] = comment
	return comment, nil
}

func (m *MockCommentRepository) Read(ctx context.Context, id primitive.ObjectID) (*model.Comment, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	comment, ok := m.comments[id]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	return comment, nil
}

func (m *MockCommentRepository) UpdateCommentByID(ctx context.Context, commentID primitive.ObjectID, updates map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	comment, ok := m.comments[commentID]
	if !ok {
		return mongo.ErrNoDocuments
	}
	if content, ok := updates["content"].(string); ok {
		comment.Content = content
	}
	return nil
}

func (m *MockCommentRepository) Delete(ctx context.Context, commentID primitive.ObjectID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.comments, commentID)
	return nil
}

// MockUserRepository is a mock implementation of repository.User
type MockUserRepository struct {
	users map[primitive.ObjectID]*model.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[primitive.ObjectID]*model.User),
	}
}

func (m *MockUserRepository) Read(id primitive.ObjectID) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) Insert(*model.User) (*model.User, error) { return nil, nil }
func (m *MockUserRepository) Update(*model.User, map[string]interface{}) (*model.User, error) {
	return nil, nil
}
func (m *MockUserRepository) ReadByEmail(string) (*model.User, error) { return nil, nil }
func (m *MockUserRepository) ReadOwnPosts(id primitive.ObjectID, isPublishedFilter *bool) ([]*model.Post, error) {
	return nil, nil
}

// MockPostRepository is a mock implementation of repository.Post
type MockPostRepository struct {
	posts map[primitive.ObjectID]*model.Post
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		posts: make(map[primitive.ObjectID]*model.Post),
	}
}

func (m *MockPostRepository) Read(id primitive.ObjectID) (*model.Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	return post, nil
}

func (m *MockPostRepository) Insert(*model.Post) (*model.Post, error)                    { return nil, nil }
func (m *MockPostRepository) AddPostTags(primitive.ObjectID, []primitive.ObjectID) error { return nil }
func (m *MockPostRepository) FindSlugsLike(string) ([]string, error)                     { return nil, nil }
func (m *MockPostRepository) GetTags(primitive.ObjectID) ([]*model.Tag, error)           { return nil, nil }
func (m *MockPostRepository) ReadByCondition(map[string]interface{}, ...string) (*model.Post, error) {
	return nil, nil
}
func (m *MockPostRepository) Select(*ct.ListPostRequest) ([]*model.Post, error)    { return nil, nil }
func (m *MockPostRepository) UpdatePost(*model.Post, map[string]interface{}) error { return nil }
func (m *MockPostRepository) UpdatePostTag(*model.Post, []*model.Tag) error        { return nil }
func (m *MockPostRepository) Delete(primitive.ObjectID) error                      { return nil }

func TestCommentService_Create(t *testing.T) {
	tests := []struct {
		name      string
		req       *ct.CreateCommentRequest
		userID    primitive.ObjectID
		setupPost bool
		wantErr   bool
		errType   error
	}{
		{
			name: "successful creation",
			req: &ct.CreateCommentRequest{
				Content: "Great post!",
				PostID:  primitive.NewObjectID(),
			},
			userID:    primitive.NewObjectID(),
			setupPost: true,
			wantErr:   false,
		},
		{
			name: "post not found",
			req: &ct.CreateCommentRequest{
				Content: "Great post!",
				PostID:  primitive.NewObjectID(),
			},
			userID:    primitive.NewObjectID(),
			setupPost: false,
			wantErr:   true,
			errType:   static.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommentRepo := NewMockCommentRepository()
			mockUserRepo := NewMockUserRepository()
			mockPostRepo := NewMockPostRepository()

			if tt.setupPost {
				post := &model.Post{
					BaseModel:   model.BaseModel{ID: tt.req.PostID},
					Title:       "Test Post",
					IsPublished: true,
				}
				mockPostRepo.posts[tt.req.PostID] = post
			}

			user := &model.User{BaseModel: model.BaseModel{ID: tt.userID}, Email: "test@example.com"}
			mockUserRepo.users[tt.userID] = user

			svc := NewService(mockCommentRepo, mockUserRepo, mockPostRepo)

			result, err := svc.Create(tt.req, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("expected error %v, got %v", tt.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result == nil {
					t.Errorf("expected result but got nil")
					return
				}
				if result.Content != tt.req.Content {
					t.Errorf("expected content %s, got %s", tt.req.Content, result.Content)
				}
			}
		})
	}
}

func TestCommentService_List(t *testing.T) {
	mockCommentRepo := NewMockCommentRepository()
	mockUserRepo := NewMockUserRepository()
	mockPostRepo := NewMockPostRepository()

	postID := primitive.NewObjectID()
	comment1ID := primitive.NewObjectID()
	comment2ID := primitive.NewObjectID()
	comment1 := &model.Comment{
		BaseModel: model.BaseModel{ID: comment1ID},
		PostID:    postID,
		Content:   "First comment",
	}
	comment2 := &model.Comment{
		BaseModel: model.BaseModel{ID: comment2ID},
		PostID:    postID,
		Content:   "Second comment",
	}
	mockCommentRepo.comments[comment1ID] = comment1
	mockCommentRepo.comments[comment2ID] = comment2

	svc := NewService(mockCommentRepo, mockUserRepo, mockPostRepo)

	req := &ct.ListCommentRequest{
		PostID:   postID,
		Page:     1,
		PageSize: 10,
	}

	result, err := svc.List(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if len(result.Comments) != 2 {
		t.Errorf("expected 2 comments, got %d", len(result.Comments))
	}
}
