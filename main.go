package main

import (
	"fmt"
	md "manage-me/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"log"
	_ "manage-me/docs"
	"os"
)

// Load environment variables
func loadEnv() string {
	// Get the environment type (default to "local" if not set)
	env := getEnv()

	// Load corresponding .env file
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("⚠️  No %s file found, using system environment variables", envFile)
	}

	return env
}

func getEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local" // Default environment
	}
	return env
}

func main() {
	// Load environment variables
	env := loadEnv()

	// Get server address and port
	host := os.Getenv("SERVER_HOST") // e.g., "0.0.0.0"
	port := os.Getenv("SERVER_PORT") // e.g., "8080"
	address := host + ":" + port

	// Fiber app with optimized settings
	app := fiber.New(fiber.Config{
		Prefork:      false, // Enables multi-process mode for better performance
		ServerHeader: "ManageMe",
		AppName:      "ManageMe API v1.0",
	})

	// Middleware (Logging, Recovery, Security)
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		return c.Next()
	})

	// API Routes
	api := app.Group("/api/v1")

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(fiber.Map{"msg": "Welcome to ManageMe API"})
	})

	var users []md.User

	app.Post("/register", func(ctx *fiber.Ctx) error {
		user := &md.User{}
		if err := ctx.BodyParser(user); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		if user.Email == "" || user.Password == "" || user.FirstName == "" || user.LastName == "" {
			return ctx.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
		}

		user.ID = len(users) + 1
		users = append(users, *user)

		return ctx.Status(200).JSON(fiber.Map{"message": "User created successfully", "data": user})
	})

	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	// Swagger Docs Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Start Server
	log.Printf("🚀 %s Server running on port %s\n", strings.ToUpper(env), port)
	if err := app.Listen(address); err != nil {
		log.Fatalf("❌ Failed to start %s server: %v", env, err)
	}
}
