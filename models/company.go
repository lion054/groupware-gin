package models

import (
	"time"

	driver "github.com/arangodb/go-driver"
)

// if the deleted_at field is empty,
// remove it from all results of json encode (omitempty)

type Company struct {
	ID        driver.DocumentID `json:"_id,omitempty"`  // empty on create
	Key       string            `json:"_key,omitempty"` // empty on create
	Rev       string            `json:"_rev,omitempty"` // empty on create
	Name      string            `json:"name"`
	Since     time.Time         `json:"since"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}
