package seeds

import (
	"context"
	"math"
	"time"

	"groupware-gin/helpers"
	"groupware-gin/models"

	driver "github.com/arangodb/go-driver"
	"syreclabs.com/go/faker"
)

func InstallCompanies() error {
	ctx := context.Background()

	// open database
	db, err := helpers.OpenDatabase()
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
