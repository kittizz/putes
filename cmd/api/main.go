package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		fmt.Println("accepted request")
		return c.SendString("Hello, World!")
	})

	app.Listen(":8081")
}
