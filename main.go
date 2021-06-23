package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"groupware-gin/controllers"
	"groupware-gin/seeds"
)

var server = controllers.Server{}

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

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env %v\n", err)
	}

	server.Initialize()
	server.Router.Run(":" + os.Getenv("PORT"))
}
