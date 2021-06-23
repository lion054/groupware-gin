package helpers

import (
	"context"
	"os"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func OpenDatabase() (driver.Database, error) {
	// create db connection
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://" + host + ":" + port},
	})
	if err != nil {
		return nil, err
	}
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(username, password),
	})
	if err != nil {
		return nil, err
	}

	// open database
	ctx := context.Background()
	dbName := os.Getenv("DB_DATABASE")
	db, err := c.Database(ctx, dbName)
	if err != nil {
		return nil, err
	}

	return db, nil
}
