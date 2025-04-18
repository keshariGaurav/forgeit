package main

import (
	"os"
	"log"

	"authentication/configs"
	"authentication/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("AWS_ACCESS_KEY_ID"))
	os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("AWS_SECRET_ACCESS_KEY"))
	os.Setenv("AWS_REGION", os.Getenv("AWS_REGION"))
	app := fiber.New()
	configs.ConnectDB()
	routes.UserRoute(app)
	routes.AuthRoute(app)
	

	app.Listen(":6000")
}
