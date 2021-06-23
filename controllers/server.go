package controllers

import (
	"net/http"
	"os"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"groupware-gin/helpers"
)

type Server struct {
	DB     driver.Database
	Router *gin.Engine
}

func (s *Server) Initialize() error {
	db, err := helpers.OpenDatabase()
	if err != nil {
		return err
	}
	s.DB = db
	s.Router = gin.Default()
	s.setUpCORS()
	s.setUpRoutes()
	return nil
}

func (s *Server) setUpCORS() {
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	s.Router.Use(
		cors.New(
			cors.Config{
				AllowOrigins:     []string{os.Getenv("ORIGIN_ALLOWED")},
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
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ping",
	})
}

func (s *Server) setUpRoutes() {
	s.Router.GET("/ping", Ping)

	// companies routes
	s.Router.GET("/companies", s.FindCompanies)
	s.Router.GET("/companies/:key", s.ShowCompany)
	s.Router.POST("/companies", s.StoreCompany)
	s.Router.PUT("/companies/:key", s.UpdateCompany)
	s.Router.DELETE("/companies/:key", s.DeleteCompany)
	s.Router.PATCH("/companies/:key", s.RestoreCompany)

	// users routes
}
