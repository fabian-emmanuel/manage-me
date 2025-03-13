package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	md "manage-me/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"log"
	_ "manage-me/docs"
	"os"
)

var collection *mongo.Collection

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
	HOST := os.Getenv("SERVER_HOST") // e.g., "0.0.0.0"
	PORT := os.Getenv("SERVER_PORT") // e.g., "8080"
	address := HOST + ":" + PORT
	mongoDBUri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoDBUri))

	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}

	defer func() {
		err := client.Disconnect(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("❌ Failed to ping MongoDB: %v", err)
	}

	log.Println("🚀 Connection to MongoDB established")

	collection = client.Database("manage-me").Collection("users")

	// Fiber app with optimized settings
	app := fiber.New(fiber.Config{
		Prefork:      false, // Enables multi-process mode for better performance
		ServerHeader: "ManageMe",
		AppName:      "ManageMe API v1.0",
	})

	// Middleware (Logging, Recovery, Security)
	//app.Use(func(c *fiber.Ctx) error {
	//	c.Set("X-Content-Type-Options", "nosniff")
	//	c.Set("X-Frame-Options", "DENY")
	//	c.Set("X-XSS-Protection", "1; mode=block")
	//	return c.Next()
	//})

	// API Routes
	registerApis(app)

	// Swagger Docs Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Start Server
	log.Printf("🚀 %s Server running on port %s\n", strings.ToUpper(env), PORT)
	if err := app.Listen(address); err != nil {
		log.Fatalf("❌ Failed to start %s server: %v", env, err)
	}
}

func registerApis(app *fiber.App) {
	api := app.Group("/api/v1")
	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})
	api.Post("/register", registerUser)
	api.Get("/all", getAllUsers)
}

func registerUser(c *fiber.Ctx) error {
	user := new(md.User)

	if err := c.BodyParser(user); err != nil {
		return serverError(err, c)
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return serverError(err, c)
	}
	user.ID = result.InsertedID.(bson.ObjectID)

	return c.Status(201).JSON(fiber.Map{"message": "User created successfully", "user": user})
}

func getAllUsers(c *fiber.Ctx) error {
	var users []md.User
	result, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return serverError(err, c)
	}

	defer func() {
		err := result.Close(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	for result.Next(context.Background()) {
		var user md.User
		err := result.Decode(&user)
		if err != nil {
			return serverError(err, c)
		}
		users = append(users, user)
	}
	return c.Status(200).JSON(fiber.Map{"message": "Users retrieved successfully", "users": users})
}

func serverError(err error, c *fiber.Ctx) error {
	log.Printf("❌ Internal Server Error: %v", err.Error())
	return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error. Please try again later."})
}
