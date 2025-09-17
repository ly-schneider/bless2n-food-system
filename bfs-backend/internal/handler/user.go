package handler

import (
	_ "backend/internal/domain"
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService service.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}
