package seeds

import (
	"context"
	"math"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
	"syreclabs.com/go/faker"

	"groupware-gin/helpers"
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
	if err != nil {
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
		now := time.Now().UTC()
		_, err := col.CreateDocument(ctx, gin.H{
			"name":        faker.Company().Name(),
			"since":       faker.Date().Backward(time.Duration(math.Pow10(9) * 3600 * 24 * 365 * 10)).UTC(),
			"created_at":  now,
			"modified_at": now,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
