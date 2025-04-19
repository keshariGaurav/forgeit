package controllers

import (
	"authentication/configs"
	"authentication/models"
	"authentication/responses"
	"authentication/utils"
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")
var validate = validator.New()

func CreateUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name := c.FormValue("name")
	email := c.FormValue("email")
	location := c.FormValue("location")
	title := c.FormValue("title")
	address := c.FormValue("address")
	linkedin := c.FormValue("linkedin")
	twitter := c.FormValue("twitter")
	dob := c.FormValue("dob")
	fileHeader, err := c.FormFile("resume")
	
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.UserResponse{
			Status:  http.StatusBadRequest,
			Message: "Resume file is required",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to open resume file",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}
	defer file.Close()

	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to save user - user already exists",
			Data:    &fiber.Map{"data": "User with this email already exists"},
		})
	}

	s3Client, bucketName := utils.InitS3()
	resumeURL, err := utils.UploadToS3(s3Client, bucketName, file, fileHeader.Filename)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to upload resume to S3",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}

	newUser := models.User{
		Id:       primitive.NewObjectID(),
		Name:     name,
		Email: 		email,
		Location: location,
		Title:    title,
		Address:  address,
		LinkedIn: linkedin,
		Twitter:  twitter,
		DOB:      dob,
		Resume:   resumeURL,
	}

	if validationErr := validate.Struct(&newUser); validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.UserResponse{
			Status:  http.StatusBadRequest,
			Message: "Validation failed",
			Data:    &fiber.Map{"data": validationErr.Error()},
		})
	}

	result, err := userCollection.InsertOne(ctx, newUser)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to save user",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}

	return c.Status(http.StatusCreated).JSON(responses.UserResponse{
		Status:  http.StatusCreated,
		Message: "User created successfully",
		Data:    &fiber.Map{"data": result},
	})
}


func GetAUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	userId := c.Params("userId")
	var user models.User
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(userId)

	err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusOK).JSON(responses.UserResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"data": user}})
}

func EditAUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userId := c.Params("userId")
	var user models.User

	// Convert userId string to ObjectID
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.UserResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid user ID",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}

	// Parse request body (for non-file fields)
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.UserResponse{
			Status:  fiber.StatusBadRequest,
			Message: "failed to parse body",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}

	// Validate fields using validator
	if validationErr := validate.Struct(&user); validationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.UserResponse{
			Status:  fiber.StatusBadRequest,
			Message: "validation error",
			Data:    &fiber.Map{"data": validationErr.Error()},
		})
	}

	// Handle optional resume file upload
	var resumeURL string
	fileHeader, err := c.FormFile("resume")
	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.UserResponse{
				Status:  fiber.StatusBadRequest,
				Message: "failed to open resume file",
				Data:    &fiber.Map{"data": err.Error()},
			})
		}
		defer file.Close()

		// Upload to S3
		s3Client, bucketName := utils.InitS3()
		uploadURL, err := utils.UploadToS3(s3Client, bucketName, file, fileHeader.Filename)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.UserResponse{
				Status:  fiber.StatusInternalServerError,
				Message: "failed to upload to S3",
				Data:    &fiber.Map{"data": err.Error()},
			})
		}
		resumeURL = uploadURL
	}

	// Build update payload
	update := bson.M{
		"name":     user.Name,
		"location": user.Location,
		"title":    user.Title,
		"address":  user.Address,
		"linkedin": user.LinkedIn,
		"twitter":  user.Twitter,
		"dob":      user.DOB,
	}
	if resumeURL != "" {
		update["resume_url"] = resumeURL
	}

	// Perform update in MongoDB
	result, err := userCollection.UpdateOne(ctx, bson.M{"id": objId}, bson.M{"$set": update})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.UserResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to update user",
			Data:    &fiber.Map{"data": err.Error()},
		})
	}

	// Fetch updated user
	var updatedUser models.User
	if result.MatchedCount == 1 {
		err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&updatedUser)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.UserResponse{
				Status:  fiber.StatusInternalServerError,
				Message: "failed to fetch updated user",
				Data:    &fiber.Map{"data": err.Error()},
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(responses.UserResponse{
		Status:  fiber.StatusOK,
		Message: "user updated successfully",
		Data:    &fiber.Map{"data": updatedUser},
	})
}



func DeleteAUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	userId := c.Params("userId")
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(userId)

	result, err := userCollection.DeleteOne(ctx, bson.M{"id": objId})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	if result.DeletedCount < 1 {
		return c.Status(http.StatusNotFound).JSON(
			responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: &fiber.Map{"data": "User with specified ID not found!"}},
		)
	}

	return c.Status(http.StatusOK).JSON(
		responses.UserResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"data": "User successfully deleted!"}},
	)
}

func GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var users []models.User
	defer cancel()

	results, err := userCollection.Find(ctx, bson.M{})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//reading from the db in an optimal way
	defer results.Close(ctx)
	for results.Next(ctx) {
		var singleUser models.User
		if err = results.Decode(&singleUser); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		users = append(users, singleUser)
	}

	return c.Status(http.StatusOK).JSON(
		responses.UserResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"data": users}},
	)
}
