package helpers

import (
	"context"
	"os"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func OpenDatabase() (driver.Database, context.Context, error) {
	// create db connection
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT")},
	})
	if err != nil {
		return nil, nil, err
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")),
	})
	if err != nil {
		return nil, nil, err
	}
	ctx := context.Background()

	// open database
	db, err := c.Database(ctx, os.Getenv("DB_DATABASE"))
	if err != nil {
		return nil, nil, err
	}

	return db, ctx, nil
}
