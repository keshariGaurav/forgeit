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
	username := c.Locals("username").(string)

	if err != nil {
	return utils.RespondWithError(c, fiber.StatusBadRequest, "Resume file is required", err)
	}
	
	file, err := fileHeader.Open()
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to open resume file", err)
	}
	defer file.Close()


	var user models.User
	err = userCollection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": username},
		}}).Decode(&user)
	if err == nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to save user - user already exists", err)
	}


	s3Client, bucketName := utils.InitS3()
	resumeURL, err := utils.UploadToS3(s3Client, bucketName, file, fileHeader.Filename)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to upload resume to S3", err)
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
		Username: username,
	}

	if validationErr := validate.Struct(&newUser); validationErr != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Validation failed", validationErr)
	}

	result, err := userCollection.InsertOne(ctx, newUser)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to save user", err)
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
		return utils.RespondWithError(c, http.StatusInternalServerError, "User does not exist.", err)
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
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid user ID", err)
	}

	// Parse request body (for non-file fields)
	if err := c.BodyParser(&user); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Failed to parse body.", err)
	}

	// Validate fields using validator
	if validationErr := validate.Struct(&user); validationErr != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Validation Error", validationErr)
	}

	// Handle optional resume file upload
	var resumeURL string
	fileHeader, err := c.FormFile("resume")
	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			return utils.RespondWithError(c, fiber.StatusBadRequest, "Failed to open resume file.", err)
		}
		defer file.Close()

		// Upload to S3
		s3Client, bucketName := utils.InitS3()
		uploadURL, err := utils.UploadToS3(s3Client, bucketName, file, fileHeader.Filename)
		if err != nil {
			return utils.RespondWithError(c, fiber.StatusInternalServerError, "Failed to upload to S3.", err)
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
			return utils.RespondWithError(c, fiber.StatusInternalServerError, "Failed to update user.", err)
	}

	// Fetch updated user
	var updatedUser models.User
	if result.MatchedCount == 1 {
		err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&updatedUser)
		if err != nil {
			return utils.RespondWithError(c, fiber.StatusInternalServerError, "Failed to fetch updated user.", err)
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
			return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete user.", err)
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
			return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to Fetch users.", err)
	}

	//reading from the db in an optimal way
	defer results.Close(ctx)
	for results.Next(ctx) {
		var singleUser models.User
		if err = results.Decode(&singleUser); err != nil {
			return utils.RespondWithError(c, http.StatusInternalServerError, "Error in fetching.", err)
		}

		users = append(users, singleUser)
	}

	return c.Status(http.StatusOK).JSON(
		responses.UserResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"data": users}},
	)
}
