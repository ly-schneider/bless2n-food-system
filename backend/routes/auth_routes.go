package routes

import (
	"backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupAuthRoutes(e *echo.Echo, ctrl *controllers.AuthController) {
	e.GET("/auth/login", ctrl.Login)
	e.GET("/auth/callback", ctrl.Callback)
	e.GET("/user", ctrl.User)
	e.GET("/auth/logout", ctrl.Logout)
}
