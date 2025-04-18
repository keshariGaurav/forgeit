package controllers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"authentication/models"
	"authentication/utils"
	"authentication/configs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCol *mongo.Collection = configs.GetCollection(configs.DB, "auth")


func Register(c *fiber.Ctx) error {
	var user models.Auth
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Check if user already exists
	count, _ := userCol.CountDocuments(context.TODO(), bson.M{"username": user.Username})
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "User already exists"})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	user.Password = string(hash)
	_, err := userCol.InsertOne(context.TODO(), user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to register user"})
	}

	return c.JSON(fiber.Map{"message": "User registered"})
}

func Login(c *fiber.Ctx) error {
	var data models.Auth
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user models.Auth
	err := userCol.FindOne(context.TODO(), bson.M{"username": data.Username}).Decode(&user)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, _ := utils.GenerateJWT(user.Username)
	return c.JSON(fiber.Map{"token": token})
}

