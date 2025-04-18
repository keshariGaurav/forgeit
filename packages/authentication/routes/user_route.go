package routes

import (
	"authentication/controllers"
	"authentication/middleware"

	"github.com/gofiber/fiber/v2"
)

func UserRoute(app *fiber.App) {
	app.Post("/user", middleware.Protected(), controllers.CreateUser)
	app.Get("/user/:userId", controllers.GetAUser)
	app.Put("/user/:userId", middleware.Protected(), controllers.EditAUser)
	app.Delete("/user/:userId", middleware.Protected(),  controllers.DeleteAUser)
	app.Get("/users", controllers.GetAllUsers)
}
