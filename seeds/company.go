package seeds

import (
	"context"
	"math"
	"time"

	"groupware-gin/models"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"syreclabs.com/go/faker"
)

func InstallCompanies() error {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		return err
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("root", ""),
	})
	if err != nil {
		return err
	}
	ctx := context.Background()

	// open database
	db, err := c.Database(ctx, "_system")
	if err != nil {
		return err
	}

	// at the first, clean up old collection
	found, err := db.CollectionExists(ctx, "companies")
	if err == nil {
		return err
	}
	if found {
		col, err := db.Collection(ctx, "companies")
		if err != nil {
			return err
		}
		col.Remove(ctx)
	}

	// create new collection
	options := &driver.CreateCollectionOptions{
		Type: driver.CollectionTypeDocument,
	}
	col, err := db.CreateCollection(ctx, "companies", options)
	if err != nil {
		return err
	}

	// create a few companies in this collection
	for i := 0; i < 10; i++ {
		doc := models.Company{
			Name:  faker.Company().Name(),
			Since: faker.Date().Backward(time.Duration(math.Pow10(9) * 3600 * 24 * 365 * 10)),
		}
		_, err := col.CreateDocument(ctx, doc)
		if err != nil {
			return err
		}
	}

	return nil
}
