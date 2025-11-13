package comment

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	ct "golang-project/internal/contract"
	"golang-project/internal/model"
)

// prepareCommentModel creates a new comment model from request data
func prepareCommentModel(request *ct.CreateCommentRequest, userID primitive.ObjectID, post *model.Post) *model.Comment {
	var parentCommentID *primitive.ObjectID
	if request.ParentCommentID.ObjectID != nil {
		parentCommentID = request.ParentCommentID.ObjectID
	}
	return &model.Comment{
		Content:         request.Content,
		PostID:          request.PostID,
		UserID:          userID,
		ParentCommentID: parentCommentID,
	}
}

// prepareUpdateComment updates fields of a Comment model and prepares the update map
func prepareUpdateComment(o *model.Comment, req *ct.UpdateCommentRequest) map[string]interface{} {
	if req.Content != "" {
		o.Content = req.Content
	}

	return map[string]interface{}{
		"content": o.Content,
	}
}

// prepareProfileResponse converts a model.User to ct.ProfileResponse
func prepareProfileResponse(user *model.User) *ct.ProfileResponse {
	if user == nil {
		return nil
	}
	return &ct.ProfileResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Pseudonym:    user.Pseudonym,
		ProfileImage: user.ProfileImage,
		Biography:    user.Biography,
		CreatedAt:    formatTime(user.CreatedAt),
		UpdatedAt:    formatTime(user.UpdatedAt),
	}
}

// prepareTagResponse converts a model.Tag to ct.TagResponse
func prepareTagResponse(tag *model.Tag) *ct.TagResponse {
	if tag == nil {
		return nil
	}
	return &ct.TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}
}

// prepareListTagResponse transforms the data and returns the List Tag Response
func prepareListTagResponse(o []*model.Tag) []*ct.TagResponse {
	data := make([]*ct.TagResponse, 0, len(o))
	for _, tag := range o {
		data = append(data, prepareTagResponse(tag))
	}
	return data
}

// preparePostResponse converts a model.Post to ct.PostResponse
func preparePostResponse(post *model.Post) *ct.PostResponse {
	if post == nil {
		return nil
	}
	postResp := &ct.PostResponse{
		ID:          post.ID,
		Title:       post.Title,
		Body:        post.Body,
		Slug:        post.Slug,
		IsPublished: post.IsPublished,
		CreatedAt:   formatTime(post.CreatedAt),
		UpdatedAt:   formatTime(post.UpdatedAt),
	}

	if post.User != nil {
		postResp.User = prepareProfileResponse(post.User)
	}

	if post.Tags != nil {
		postResp.Tags = prepareListTagResponse(post.Tags)
	}

	return postResp
}

// prepareChildCommentResponse converts a model.Comment to ct.ChildCommentResponse
func prepareChildCommentResponse(o *model.Comment, u *model.User) *ct.ChildCommentResponse {
	if o == nil {
		return nil
	}
	data := &ct.ChildCommentResponse{
		ID:              o.ID,
		Content:         o.Content,
		ParentCommentID: o.ParentCommentID,
		User:            prepareProfileResponse(u),
	}

	if o.CreatedAt != nil {
		data.CreatedAt = formatTime(o.CreatedAt)
	}

	if o.UpdatedAt != nil {
		data.UpdatedAt = formatTime(o.UpdatedAt)
	}

	return data
}

// prepareCommentResponse converts a model.Comment to ct.CommentResponse
func prepareCommentResponse(comment *model.Comment) *ct.CommentResponse {
	if comment == nil {
		return nil
	}
	data := &ct.CommentResponse{
		ID:      comment.ID,
		Content: comment.Content,
	}

	if comment.ParentCommentID != nil {
		data.ParentCommentID = comment.ParentCommentID
	}

	if comment.CreatedAt != nil {
		data.CreatedAt = formatTime(comment.CreatedAt)
	}

	if comment.UpdatedAt != nil {
		data.UpdatedAt = formatTime(comment.UpdatedAt)
	}

	// Add user info if available
	if comment.User != nil {
		data.User = prepareProfileResponse(comment.User)
	}

	// Add post info if available
	if comment.Post != nil {
		data.Post = preparePostResponse(comment.Post)
	}

	// Add child comments if available
	if comment.ChildComments != nil {
		data.ChildComments = make([]*ct.ChildCommentResponse, 0, len(comment.ChildComments))
		for _, child := range comment.ChildComments {
			var childUser *model.User
			if child.User != nil {
				childUser = child.User
			}
			childResp := prepareChildCommentResponse(child, childUser)
			if childResp != nil {
				data.ChildComments = append(data.ChildComments, childResp)
			}
		}
	}

	return data
}

// prepareListCommentResponse converts a slice of model.Comment to ct.ListCommentResponse
func prepareListCommentResponse(comments []*model.Comment, paging ct.Paging) *ct.ListCommentResponse {
	response := &ct.ListCommentResponse{
		Comments: make([]*ct.CommentResponse, 0, len(comments)),
		Paging:   paging,
	}

	for _, comment := range comments {
		response.Comments = append(response.Comments, prepareCommentResponse(comment))
	}

	return response
}

// formatTime formats time to RFC3339 string
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
