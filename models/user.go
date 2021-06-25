package models

import (
	"time"

	driver "github.com/arangodb/go-driver"
)

// remove password field to prevent sensitive info to be exposed
// this struct is used only for json output

type User struct {
	ID         driver.DocumentID `json:"_id,omitempty"`  // empty on create
	Key        string            `json:"_key,omitempty"` // empty on create
	Rev        string            `json:"_rev,omitempty"` // empty on create
	Name       string            `json:"name"`
	Email      string            `json:"email"`
	Avatar     string            `json:"avatar"`
	CreatedAt  time.Time         `json:"created_at"`
	ModifiedAt time.Time         `json:"modified_at"`
	DeletedAt  *time.Time        `json:"deleted_at,omitempty"`
}
