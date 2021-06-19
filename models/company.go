package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Company struct {
	Name  string    `json:"name"`
	Since time.Time `json:"since"`
}

func (c *Company) Validate(action string) error {
	if action == "store" {
		err := validation.ValidateStruct(c,
			// Name cannot be empty
			validation.Field(&c.Name, validation.Required, validation.Length(5, 50)),
			// Since cannot be empty
			validation.Field(&c.Since, validation.Required, validation.Date("")),
		)
		return err
	} else if action == "update" {
		err := validation.ValidateStruct(c,
			// Name cannot be empty
			validation.Field(&c.Name, validation.Length(5, 50)),
			// Since cannot be empty
			validation.Field(&c.Since, validation.Date("")),
		)
		return err
	}
	return nil
}
