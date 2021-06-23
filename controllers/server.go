package controllers

import (
	"fmt"
	"log"
	"net/http"

	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"

	"groupware-gin/helpers"
	"groupware-gin/middlewares"
	"groupware-gin/responses"
)

type Server struct {
	DB     driver.Database
	Router *mux.Router
}

func (s *Server) Initialize() error {
	db, err := helpers.OpenDatabase()
	if err != nil {
		return err
	}
	s.DB = db
	s.Router = mux.NewRouter()
	s.setUpRoutes()
	return nil
}

func (s *Server) Run(addr string) {
	fmt.Println("Listening to port 8080")
	log.Fatal(http.ListenAndServe(addr, s.Router))
}

func Ping(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, gin.H{
		"message": "ping",
	})
}

func (s *Server) setUpRoutes() {
	s.Router.HandleFunc(
		"/ping",
		middlewares.InstallJsonMiddleware(Ping),
	).Methods("GET")

	// companies routes
	s.Router.HandleFunc(
		"/companies",
		middlewares.InstallJsonMiddleware(s.FindCompanies),
	).Methods("GET")

	s.Router.HandleFunc(
		"/companies/{key}",
		middlewares.InstallJsonMiddleware(s.ShowCompany),
	).Methods("GET")

	s.Router.HandleFunc(
		"/companies",
		middlewares.InstallJsonMiddleware(s.StoreCompany),
	).Methods("POST")

	s.Router.HandleFunc(
		"/companies/{key}",
		middlewares.InstallJsonMiddleware(s.UpdateCompany),
	).Methods("PUT")

	s.Router.HandleFunc(
		"/companies/{key}",
		middlewares.InstallJsonMiddleware(s.DeleteCompany),
	).Methods("DELETE")

	s.Router.HandleFunc(
		"/companies/{key}",
		middlewares.InstallJsonMiddleware(s.RestoreCompany),
	).Methods("PATCH")

	// users routes
}
