package tag

import (
	"golang-project/internal/handler"
	hdl "golang-project/internal/handler/tag"
	repo "golang-project/internal/repository/tag"
	svc "golang-project/internal/service/tag"

	"go.mongodb.org/mongo-driver/mongo"
)

// NewRegistry returns a new resource handler for tag API
func NewRegistry(route string, db *mongo.Database) handler.ResourceHandler {
	return hdl.NewHandler(route, svc.NewService(repo.NewRepository(db)))
}
