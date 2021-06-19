package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"groupware-gin/seeds"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Use /seed flag to install fake dtabase and download fake images")
	fmt.Println()
	for _, arg := range os.Args[1:] {
		// fmt.Printf("Argument %d is %s\n", i, arg)
		if arg == "/seed" {
			seeds.InstallCompanies()
			seeds.InstallUsers()
			os.Exit(1)
		}
	}

	router := gin.Default()
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	router.Use(
		cors.New(
			cors.Config{
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
				AllowHeaders:     []string{"Origin"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
				AllowOriginFunc: func(origin string) bool {
					return origin == "https://github.com"
				},
				MaxAge: 12 * time.Hour,
			},
		),
	)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ping",
		})
	})
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
