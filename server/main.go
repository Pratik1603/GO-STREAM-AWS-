// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"server/myproject/services"
	"server/myproject/store"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

var ginLambda *ginadapter.GinLambda
var ginLambdaV2 *ginadapter.GinLambdaV2

func main() {
	godotenv.Load()

	db := InitDB()
	defer db.Close()

	// Initialize Store
	storeVal := store.NewStore(db)

	// Initialize Services
	authService := services.NewAuthService(storeVal)
	movieService := services.NewMovieService(storeVal)
	recService := services.NewRecommendationService(storeVal)

	// If in AWS Lambda, initialize S3 Client
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err == nil {
			s3Client := s3.NewFromConfig(cfg)
			movieService.SetS3Client(s3Client)
			log.Println("S3 Client initialized for Lambda environment")
		} else {
			log.Printf("Failed to load AWS config: %v", err)
		}
	}

	// Initialize Handlers
	handler := NewHandler(authService, movieService)
	recHandler := NewRecommendationHandler(recService, movieService)

	r := gin.Default()
	r.Use(CORSMiddleware())

	// Serve static files
	r.Static("/uploads", "./uploads")

	// Public routes

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "serving",
		})
	})

	r.POST("/api/register", handler.Register)
	r.POST("/api/login", handler.Login)

	// Protected routes
	api := r.Group("/api")

	api.Use(AuthMiddleware())
	{
		api.GET("/me", handler.GetProfile)

		// Movie routes
		api.GET("/movies", handler.ListMovies)
		api.GET("/movies/search", handler.SearchMovies)
		api.GET("/movies/:id", handler.GetMovie)
		api.GET("/movies/:id/episodes", handler.GetEpisodes)
		api.GET("/movies/:id/thumbnail", handler.GetThumbnail)
		api.PUT("/movies/:id", handler.UpdateMovie)
		api.DELETE("/movies/:id", handler.DeleteMovie)
		api.GET("/stream/:id", handler.StreamMovie)

		// AI Recommendations
		api.GET("/browse", recHandler.GetBrowseData)
		api.GET("/movies/:id/similar", recHandler.GetSimilarMovies)
		api.POST("/watch-history", recHandler.UpdateWatchProgress)
		api.GET("/watch-history", recHandler.GetWatchHistory)

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(AdminMiddleware())
		{
			admin.GET("/users", handler.ListUsers)
			admin.PUT("/users/:id/role", handler.UpdateUserRole)
			admin.DELETE("/users/:id", handler.DeleteUser)
			admin.POST("/movies", handler.UploadMovie)
			admin.DELETE("/movies/:id", handler.DeleteMovie)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)

	// AWS Lambda check
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		ginLambda = ginadapter.New(r)
		ginLambdaV2 = ginadapter.NewV2(r)

		lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {
			// Try to unmarshal as V2 first (HTTP API)
			var reqV2 events.APIGatewayV2HTTPRequest
			if err := json.Unmarshal(event, &reqV2); err == nil && reqV2.RawPath != "" {
				return ginLambdaV2.ProxyWithContext(ctx, reqV2)
			}

			// Fallback to V1 (Rest API or Lambda Console Test)
			var reqV1 events.APIGatewayProxyRequest
			if err := json.Unmarshal(event, &reqV1); err == nil {
				return ginLambda.ProxyWithContext(ctx, reqV1)
			}

			return nil, fmt.Errorf("unexpected event format")
		})
	} else {
		r.Run(":" + port)
	}
}
