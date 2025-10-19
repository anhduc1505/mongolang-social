package authentication

import (
	"golang-project/internal/handler"
	hdl "golang-project/internal/handler/authentication"
	repo "golang-project/internal/repository/user"
	svc "golang-project/internal/service/authentication"
	"golang-project/util/hashing"
)

// NewRegistry returns new resource handler for authentication API
func NewRegistry(route string) handler.ResourceHandler {
	return hdl.NewHandler(route, svc.NewService(repo.NewRepository(), hashing.NewBcrypt()))
}
