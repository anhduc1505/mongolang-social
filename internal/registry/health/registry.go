package health

import (
	"golang-project/database"

	"golang-project/internal/handler"
	hdl "golang-project/internal/handler/health"
)

// NewRegistry returns new resource handler for health API
func NewRegistry(route string, db database.Connection) handler.ResourceHandler {
	return hdl.NewHandler(route, db)
}
