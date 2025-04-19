// utils/error_handler.go
package utils

import "github.com/gofiber/fiber/v2"

func RespondWithError(c *fiber.Ctx, status int, message string, err error) error {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = "Unknown Error"
	}
	return c.Status(status).JSON(fiber.Map{
		"status":  status,
		"message": message,
		"error":   errMsg,
	})
}
