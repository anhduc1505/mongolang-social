package tag

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	ct "golang-project/internal/contract"
	hdl "golang-project/internal/handler"
	"golang-project/internal/middleware"
	svc "golang-project/internal/service"
	"golang-project/server"
	"golang-project/static"
)

// handler represents the implementation of hdl.Tag
type handler struct {
	route  string
	tagSvc svc.Tag
}

// NewHandler returns a new implementation of hdl.Tag
func NewHandler(route string, tagSvc svc.Tag) hdl.Tag {
	return &handler{
		route:  route,
		tagSvc: tagSvc,
	}
}

// RegisterRoutes registers the handler routes and returns the server.HandlerRegistry
func (h *handler) RegisterRoutes() server.HandlerRegistry {
	return server.HandlerRegistry{
		Route: h.route,
		Register: func(group *echo.Group) {
			group.POST("", h.Create, middleware.Authentication(nil))
			// TODO: Implement List, ListPosts, Delete methods in service
			group.GET("", h.List)
			group.GET("/:id/posts", h.ListPosts)
			group.DELETE("/:id", h.Delete, middleware.Authentication(nil))
		},
	}
}

// List handles the request to get all tags
//
//	@Summary		Get all tags
//	@Description	Readers/Bloggers can view all blog tags
//	@Tags			tag
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	contract.ListTagResponse
//	@Failure		400	{object}	error
//	@Router			/tags [get]
func (h *handler) List(e echo.Context) error {
	ctx := e.Request().Context()
	response, err := h.tagSvc.List(ctx)

	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, response)
}

// ListPosts handles the request to get all posts for a tag
//
//	@Summary		Get all posts for a tag
//	@Description	Readers/Bloggers can view all blog posts belong to a particular tag
//	@Tags			tag
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Tag ID"
//	@Success		200	{object}	contract.ListPostResponse
//	@Failure		400	{object}	error
//	@Router			/tags/{id}/posts [get]
//
// ListPosts handles the request to get all posts for a tag
// TODO: Implement ListPosts method in service
func (h *handler) ListPosts(e echo.Context) error {
	return e.JSON(http.StatusNotImplemented, "List posts by tag not yet implemented")
}

// Create handles the request to create a tag
//
//	@Summary		Create a new tag
//	@Description	Create a new tag with the provided name
//	@Tags			tag
//	@Accept			json
//	@Produce		json
//	@Security		BearerToken
//	@Param			request	body		contract.CreateTagRequest	true	"Create Tag Request"
//	@Success		200		{object}	contract.TagResponse		"Tag created successfully"
//	@Failure		400		{object}	string						"Invalid request"
//	@Failure		422		{object}	string						"Unprocessable entity"
//	@Router			/tags [post]
func (h *handler) Create(c echo.Context) error {
	var req ct.CreateTagRequest
	ctx := c.Request().Context()

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		return c.JSON(http.StatusUnprocessableEntity, "Name is required")
	}

	createdTag, err := h.tagSvc.Create(ctx, req.Name)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return c.JSON(http.StatusBadRequest, "Tag with this name already exists")
		}
		return c.JSON(http.StatusBadRequest, "Unable to create tag")
	}

	// Return created tag
	return c.JSON(http.StatusOK, createdTag)
}

// Delete handles the request to delete a tag
//
//	@Summary		Delete a tag
//	@Description	Blogger can delete a tag that does not contain any blog
//	@Tags			tag
//	@Accept			json
//	@Produce		json
//	@Security		BearerToken
//	@Param			id	path	string	true	"Tag ID"
//	@Success		204	"No Content"
//	@Failure		400	{object}	string	"Invalid request"
//	@Router			/tags/{id} [delete]
func (h *handler) Delete(e echo.Context) error {
	idParam := e.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid tag ID")
	}

	ctx := e.Request().Context()
	err = h.tagSvc.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, static.ErrTagNotFound) {
			return e.JSON(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, static.ErrHasPosts) {
			return e.JSON(http.StatusBadRequest, err.Error())
		}
		return e.JSON(http.StatusBadRequest, "Unable to delete tag")
	}

	return e.NoContent(http.StatusNoContent)
}
