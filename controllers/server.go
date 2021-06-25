package controllers

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/asaskevich/govalidator"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
	s.SetUpCors()
	s.SetUpRoutes()
	// extend validator for password confirmation
	govalidator.CustomTypeTagMap.Set("store_confirmed", govalidator.CustomTypeValidator(func(i, o interface{}) bool {
		result := i.(string) == o.(StoreUserParams).PasswordConfirmation
		return result
	}))
	govalidator.CustomTypeTagMap.Set("update_confirmed", govalidator.CustomTypeValidator(func(i, o interface{}) bool {
		result := i.(string) == o.(UpdateUserParams).PasswordConfirmation
		return result
	}))
	return nil
}

func (s *Server) SetUpCors() {
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	s.Router.Use(
		cors.New(
			cors.Config{
				AllowOrigins:     []string{os.Getenv("ORIGIN_ALLOWED")},
				AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE"},
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

func (s *Server) SetUpRoutes() {
	s.Router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ping",
		})
	})

	// companies routes
	s.Router.GET("/companies", s.FindCompanies)
	s.Router.GET("/companies/:key", s.ShowCompany)
	s.Router.POST("/companies", s.StoreCompany)
	s.Router.PATCH("/companies/:key", s.UpdateCompany)
	s.Router.DELETE("/companies/:key", s.DeleteCompany)

	// users routes
	s.Router.GET("/users", s.FindUsers)
	s.Router.GET("/users/:key", s.ShowUser)
	s.Router.POST("/users", s.StoreUser)
	s.Router.PATCH("/users/:key", s.UpdateUser)
	s.Router.DELETE("/users/:key", s.DeleteUser)
}

func (s *Server) HasCollection(name string) (bool, error) {
	ctx := context.Background()
	collections, err := s.DB.Collections(ctx)
	if err != nil {
		return false, err
	}
	for _, c := range collections {
		if c.Name() == name {
			return true, nil
		}
	}
	return false, nil
}

func IsDir(dirPath string) bool {
	pathAbs, err := filepath.Abs(dirPath)
	if err != nil {
		return false
	}
	fileInfo, err := os.Stat(pathAbs)
	if os.IsNotExist(err) {
		return false
	}
	if !fileInfo.IsDir() {
		return false
	}
	return true
}

func AcceptFile(c *gin.Context, fieldName string, destDir string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", nil
	}
	if !IsDir(destDir) {
		err = os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	fileName := uuid.New().String() + filepath.Ext(file.Filename)
	err = c.SaveUploadedFile(file, path.Join(destDir, fileName))
	if err != nil {
		return fileName, err
	}
	return fileName, nil
}
