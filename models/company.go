package models

import (
	"time"

	driver "github.com/arangodb/go-driver"
)

// if the deleted_at field is empty,
// remove it from all results of json encode (omitempty)

type Company struct {
	ID        driver.DocumentID `json:"_id"`
	Key       string            `json:"_key"`
	Rev       string            `json:"_rev"`
	Name      string            `json:"name"`
	Since     time.Time         `json:"since"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}
