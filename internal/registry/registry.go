package registry

import (
	"github.com/labstack/echo/v4"
	swagger "github.com/swaggo/echo-swagger"

	"golang-project/database"
	_ "golang-project/docs/swagger"
	"golang-project/internal/handler"
	"golang-project/internal/registry/authentication"
	"golang-project/internal/registry/health"

	// TODO: Migrate remaining services to MongoDB
	// "golang-project/internal/registry/comment"
	// "golang-project/internal/registry/favourite"
	// "golang-project/internal/registry/post"
	// "golang-project/internal/registry/profile"
	// "golang-project/internal/registry/tag"
	"golang-project/server"
)

// NewHandlerRegistries returns all server handler registries
func NewHandlerRegistries(db database.Connection) ([]server.HandlerRegistry, error) {
	registries := []server.HandlerRegistry{
		initSwaggerRegistry(),
		initHealthCheckHandler(db).RegisterRoutes(),
	}

	for _, hdl := range initResourceHandlers(db) {
		registries = append(registries, hdl.RegisterRoutes())
	}

	return registries, nil
}

// initSwaggerRegistry returns the swagger handler registry
func initSwaggerRegistry() server.HandlerRegistry {
	return server.HandlerRegistry{
		Route: "/swagger",
		Register: func(group *echo.Group) {
			group.GET("/*", swagger.WrapHandler)
		},
	}
}

// initHealthCheckHandler returns the health check handler for MongoDB
func initHealthCheckHandler(db database.Connection) handler.ResourceHandler {
	return health.NewRegistry("/health", db)
}

// initResourceHandlers returns the service resource handler registry
func initResourceHandlers(db database.Connection) []handler.ResourceHandler {
	return []handler.ResourceHandler{
		authentication.NewRegistry("/auth"),
		// TODO: Migrate remaining services to MongoDB
		// profile.NewRegistry("/profile", db),
		// tag.NewRegistry("/tags", db),
		// favourite.NewRegistry("/favorites", db),
		// comment.NewRegistry("/comments", db),
		// post.NewRegistry("/posts", db),
	}
}
