package routes

import (
	"authentication/controllers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(app *fiber.App) {
	app.Post("auth/register", controllers.Register)
	app.Post("auth/login", controllers.Login)

}
