package models

import (
	"time"
)

type User struct {
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	Avatar    string     `json:"avatar"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
