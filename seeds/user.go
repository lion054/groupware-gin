package seeds

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	netHttp "net/http"
	"os"
	"time"

	"groupware-gin/models"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/google/uuid"
	"syreclabs.com/go/faker"
)

func InstallUsers() error {
	// remove the existing user avatars from local disk
	err := os.RemoveAll("storage/users")
	if err != nil {
		return err
	}

	// create db connection
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
	found, err := db.CollectionExists(ctx, "users")
	if err != nil {
		return err
	}
	if found {
		usersCollection, err := db.Collection(ctx, "users")
		if err != nil {
			return err
		}
		usersCollection.Remove(ctx)
	}
	found, err = db.CollectionExists(ctx, "work_at")
	if err != nil {
		return err
	}
	if found {
		workAtCollection, err := db.Collection(ctx, "work_at")
		if err != nil {
			return err
		}
		workAtCollection.Remove(ctx)
	}
	found, err = db.GraphExists(ctx, "employment")
	if err != nil {
		return err
	}
	if found {
		graph, err := db.Graph(ctx, "employment")
		if err != nil {
			return err
		}
		graph.Remove(ctx)
	}

	// create new collections
	options := &driver.CreateCollectionOptions{
		Type: driver.CollectionTypeDocument,
	}
	usersCollection, err := db.CreateCollection(ctx, "users", options)
	if err != nil {
		return err
	}
	options = &driver.CreateCollectionOptions{
		Type: driver.CollectionTypeEdge,
	}
	workAtCollection, err := db.CreateCollection(ctx, "work_at", options)
	if err != nil {
		return err
	}

	// create new graph
	edgeDef := driver.EdgeDefinition{
		Collection: "work_at",
		From:       []string{"users"},
		To:         []string{"companies"},
	}
	graphGptions := &driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{edgeDef},
	}
	_, err = db.CreateGraph(ctx, "employment", graphGptions)
	if err != nil {
		return err
	}

	// create a few users about every company
	query := "FOR x IN companies RETURN x"
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	hasher := md5.New()
	pswd := []byte("123456")
	for {
		var company models.Company
		companyMeta, err := cursor.ReadDocument(ctx, &company)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		count := faker.Number().NumberInt(1)
		for i := 0; i < count; i++ {
			doc := models.User{
				Name:     faker.Name().Name(),
				Email:    faker.Internet().Email(),
				Password: hex.EncodeToString(hasher.Sum(pswd)),
			}
			userMeta, err := usersCollection.CreateDocument(ctx, doc)
			if err != nil {
				return err
			}
			// create the avatar
			err = os.MkdirAll("storage/users/"+userMeta.Key, os.ModePerm)
			if err != nil {
				return err
			}
			fileName := uuid.New().String() + ".jpg"
			filePath := "users/" + userMeta.Key + "/" + fileName
			DownloadFile("https://thispersondoesnotexist.com/image", "storage/"+filePath)
			patch := map[string]interface{}{
				"Avatar": filePath,
			}
			userMeta, err = usersCollection.UpdateDocument(ctx, userMeta.Key, patch)
			if err != nil {
				return err
			}
			// register user to company
			workAtData := models.WorkAt{
				From:  "users/" + userMeta.Key,
				To:    "companies/" + companyMeta.Key,
				Since: faker.Date().Backward(time.Duration(math.Pow10(9) * 3600 * 24 * 365 * 10)),
			}
			workAtCollection.CreateDocument(ctx, workAtData)
		}
	}

	return nil
}

func DownloadFile(srcURL string, destPath string) error {
	// create blink file
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	client := netHttp.Client{
		CheckRedirect: func(req *netHttp.Request, via []*netHttp.Request) error {
			req.URL.Opaque = req.URL.Path
			return nil
		},
	}
	// put content on file
	resp, err := client.Get(srcURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Printf("Downloaded a file %s with size %d\n", destPath, size)
	return nil
}
