package controllers

import (
	"context"
	"errors"
	"fmt"

	"authentication/configs"
	"authentication/models"
	"authentication/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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

func UpdatePassword(c *fiber.Ctx) error {
	type Request struct {
	Username        string `json:"username" validate:"required"`
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required"`
}
var validate = validator.New()

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	if err := validate.Struct(req); err != nil {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errorMessages := make(map[string]string)
		for _, e := range ve {
			errorMessages[e.Field()] = fmt.Sprintf("failed on '%s' tag", e.Tag())
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"validationErrors": errorMessages})
	}

	// fallback in case it's not a ValidationError
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
}

	var user models.Auth
	err := userCol.FindOne(context.TODO(), bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Current password is incorrect"})
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash new password"})
	}

	// Update password in DB
	update := bson.M{"$set": bson.M{"password": string(hashedPassword)}}
	_, err = userCol.UpdateOne(context.TODO(), bson.M{"username": req.Username}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}


