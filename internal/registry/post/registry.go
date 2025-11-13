package post

import (
	"go.mongodb.org/mongo-driver/mongo"

	"golang-project/internal/handler"
	hdl "golang-project/internal/handler/post"
	postRepo "golang-project/internal/repository/post"
	tagRepo "golang-project/internal/repository/tag"
	userRepo "golang-project/internal/repository/user"
	svc "golang-project/internal/service/post"
)

// NewRegistry returns a new resource handler for post API
func NewRegistry(route string, db *mongo.Database) handler.ResourceHandler {
	return hdl.NewHandler(route, svc.NewService(
		postRepo.NewRepository(db),
		userRepo.NewRepository(),
		tagRepo.NewRepository(db),
	))
}
