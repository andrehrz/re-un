package main

import (
	"fmt"
	"net/http"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	fmt.Println("Starting backend")
	if err := godotenv.Load(); err != nil {
		log.Println("Could not load environment variables, no .env file found")
	}

	port := os.Getenv("PORT")
	if port == ""{
		port = "8080"
	}

	// this is my router
	router := gin.Default()

	// defining some simple GET endpoints
	router.GET("/ping", func(c *gin.Context) {
		// returns a JSON response
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "server is running",
		})
	})

	log.Printf("Server is running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server")
	}
}
