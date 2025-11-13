package contract

import (
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OptionalObjectID is a custom type that can unmarshal from string, null, or empty string
// It treats empty strings, "string", and null as nil
type OptionalObjectID struct {
	*primitive.ObjectID
}

// UnmarshalJSON implements custom unmarshaling for OptionalObjectID
func (o *OptionalObjectID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// If it's null or invalid, set to nil
		o.ObjectID = nil
		return nil
	}

	// Treat empty string or "string" as nil
	s = strings.TrimSpace(s)
	if s == "" || s == "string" || len(s) != 24 {
		o.ObjectID = nil
		return nil
	}

	// Try to parse as ObjectID
	oid, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		// If invalid, set to nil instead of error
		o.ObjectID = nil
		return nil
	}

	o.ObjectID = &oid
	return nil
}

// MarshalJSON implements custom marshaling for OptionalObjectID
func (o OptionalObjectID) MarshalJSON() ([]byte, error) {
	if o.ObjectID == nil {
		return []byte("null"), nil
	}
	return json.Marshal(o.ObjectID.Hex())
}

// CommentResponse defines the structure of a single comment
// with all child comments returned in the API response.
type CommentResponse struct {
	ID              primitive.ObjectID      `json:"id,omitempty"`
	Content         string                  `json:"content,omitempty"`
	User            *ProfileResponse        `json:"user,omitempty"`
	Post            *PostResponse           `json:"post,omitempty"`
	ChildComments   []*ChildCommentResponse `json:"child_comments,omitempty" `
	ParentCommentID *primitive.ObjectID     `json:"parent_comment_id,omitempty"`
	CreatedAt       string                  `json:"created_at,omitempty"`
	UpdatedAt       string                  `json:"updated_at,omitempty"`
}

// ChildCommentResponse defines the structure of a single child comment returned .
type ChildCommentResponse struct {
	ID              primitive.ObjectID  `json:"id,omitempty"`
	Content         string              `json:"content,omitempty"`
	ParentCommentID *primitive.ObjectID `json:"parent_comment_id,omitempty"`
	User            *ProfileResponse    `json:"user,omitempty"`
	CreatedAt       string              `json:"created_at,omitempty"`
	UpdatedAt       string              `json:"updated_at,omitempty"`
}

// Paging response returned in the list parent comments API response
type Paging struct {
	Page     int `json:"page" `
	PageSize int `json:"page_size" `
	Total    int `json:"total" `
}

// ListCommentResponse wraps a list of comments that are
// returned when fetching all comments for a post.
type ListCommentResponse struct {
	Comments []*CommentResponse `json:"comments"`
	Paging   Paging             `json:"paging"`
}

// ListCommentRequest defines the query parameters for retrieving comments.
type ListCommentRequest struct {
	PostID   primitive.ObjectID `json:"post_id" query:"post_id" validate:"required"`
	Page     int                `json:"page" query:"page"`           // Page number
	PageSize int                `json:"page_size" query:"page_size"` // Number of posts per page
}

// CreateCommentRequest defines the expected payload when
// a user wants to create a new comment.
type CreateCommentRequest struct {
	Content         string             `json:"content" validate:"required"`
	PostID          primitive.ObjectID `json:"post_id" validate:"required"`
	ParentCommentID OptionalObjectID   `json:"parent_comment_id,omitempty"`
}

// UpdateCommentRequest defines the expected payload when
// a user wants to update an exist comment.
type UpdateCommentRequest struct {
	ID      primitive.ObjectID `param:"commentId" swaggerignore:"true"`
	Content string             `json:"content" validate:"required"`
}
