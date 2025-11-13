package comment

import (
	"golang-project/internal/handler"
	hdl "golang-project/internal/handler/comment"
	commentRepo "golang-project/internal/repository/comment"
	postRepo "golang-project/internal/repository/post"
	userRepo "golang-project/internal/repository/user"
	svc "golang-project/internal/service/comment"

	"go.mongodb.org/mongo-driver/mongo"
)

// NewRegistry returns a new resource handler for comment API
func NewRegistry(route string, db *mongo.Database) handler.ResourceHandler {
	return hdl.NewHandler(route, svc.NewService(
		commentRepo.NewRepository(db),
		userRepo.NewRepository(),
		postRepo.NewRepository(db),
	))
}
