package tag

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"golang-project/internal/model"
	"golang-project/static"
)

// MockTagRepository is a mock implementation of repository.Tag
type MockTagRepository struct {
	tags      map[primitive.ObjectID]*model.Tag
	hasPosts  map[primitive.ObjectID]bool
	insertErr error
	readErr   error
	deleteErr error
	selectErr error
}

func NewMockTagRepository() *MockTagRepository {
	return &MockTagRepository{
		tags:     make(map[primitive.ObjectID]*model.Tag),
		hasPosts: make(map[primitive.ObjectID]bool),
	}
}

func (m *MockTagRepository) Insert(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	if tag.ID.IsZero() {
		tag.ID = primitive.NewObjectID()
	}
	m.tags[tag.ID] = tag
	return tag, nil
}

func (m *MockTagRepository) Read(ctx context.Context, id primitive.ObjectID) (*model.Tag, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	tag, ok := m.tags[id]
	if !ok {
		return nil, errors.New("tag not found")
	}
	return tag, nil
}

func (m *MockTagRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.tags, id)
	return nil
}

func (m *MockTagRepository) HasPosts(ctx context.Context, tagID primitive.ObjectID) (bool, error) {
	return m.hasPosts[tagID], nil
}

func (m *MockTagRepository) Select(ctx context.Context, tagIDs []primitive.ObjectID) ([]*model.Tag, error) {
	if m.selectErr != nil {
		return nil, m.selectErr
	}
	if tagIDs == nil {
		// Return all tags
		result := make([]*model.Tag, 0, len(m.tags))
		for _, tag := range m.tags {
			result = append(result, tag)
		}
		return result, nil
	}
	result := make([]*model.Tag, 0, len(tagIDs))
	for _, id := range tagIDs {
		if tag, ok := m.tags[id]; ok {
			result = append(result, tag)
		}
	}
	return result, nil
}

func TestTagService_Create(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		wantErr bool
	}{
		{
			name:    "successful creation",
			tagName: "golang",
			wantErr: false,
		},
		{
			name:    "empty tag name",
			tagName: "",
			wantErr: false, // Service doesn't validate empty name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockTagRepository()
			svc := NewService(mockRepo)

			ctx := context.Background()
			result, err := svc.Create(ctx, tt.tagName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
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
				if result.Name != tt.tagName {
					t.Errorf("expected tag name %s, got %s", tt.tagName, result.Name)
				}
			}
		})
	}
}

func TestTagService_List(t *testing.T) {
	mockRepo := NewMockTagRepository()

	// Add some test tags
	tag1 := &model.Tag{ID: primitive.NewObjectID(), Name: "golang"}
	tag2 := &model.Tag{ID: primitive.NewObjectID(), Name: "python"}
	mockRepo.tags[tag1.ID] = tag1
	mockRepo.tags[tag2.ID] = tag2

	svc := NewService(mockRepo)
	ctx := context.Background()

	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if len(result.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result.Tags))
	}
}

func TestTagService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		setupTag bool
		hasPosts bool
		wantErr  bool
		errType  error
	}{
		{
			name:     "successful deletion",
			setupTag: true,
			hasPosts: false,
			wantErr:  false,
		},
		{
			name:     "tag not found",
			setupTag: false,
			wantErr:  true,
			errType:  static.ErrTagNotFound,
		},
		{
			name:     "tag has posts",
			setupTag: true,
			hasPosts: true,
			wantErr:  true,
			errType:  static.ErrHasPosts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockTagRepository()
			tagID := primitive.NewObjectID()

			if tt.setupTag {
				mockRepo.tags[tagID] = &model.Tag{ID: tagID, Name: "test"}
				mockRepo.hasPosts[tagID] = tt.hasPosts
			}

			svc := NewService(mockRepo)
			ctx := context.Background()

			err := svc.Delete(ctx, tagID)

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
				// Verify tag was deleted
				if _, ok := mockRepo.tags[tagID]; ok {
					t.Errorf("tag should be deleted but still exists")
				}
			}
		})
	}
}
