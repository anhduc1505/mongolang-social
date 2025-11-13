package profile

import (
	"go.mongodb.org/mongo-driver/mongo"

	"golang-project/internal/handler"
	hdl "golang-project/internal/handler/profile"
	postRepo "golang-project/internal/repository/post"
	tagRepo "golang-project/internal/repository/tag"
	userRepo "golang-project/internal/repository/user"
	svc "golang-project/internal/service/profile"
)

// NewRegistry returns new resource handler for profile API
func NewRegistry(route string, db *mongo.Database) handler.ResourceHandler {
	return hdl.NewHandler(route, svc.NewService(
		userRepo.NewRepository(),
		postRepo.NewRepository(db),
		tagRepo.NewRepository(db),
	))
}
