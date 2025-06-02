package routes

import (
	"backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupCategoryRoutes(e *echo.Echo, ctrl *controllers.CategoryController) {
	categoriesGroup := e.Group("/categories")
	categoriesGroup.GET("", ctrl.GetCategories)
	categoriesGroup.GET("/:id", ctrl.GetCategory)
	categoriesGroup.POST("", ctrl.CreateCategory)
	categoriesGroup.PUT("/:id", ctrl.UpdateCategory)
	categoriesGroup.DELETE("/:id", ctrl.DeleteCategory)
}
