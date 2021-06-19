package models

import "time"

type WorkAt struct {
	From     string    `json:"_from"`
	To       string    `json:"_to"`
	Since    time.Time `json:"since"`
	Position string    `json:"position"`
}
