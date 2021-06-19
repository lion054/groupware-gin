package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

type User struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	Avatar               string `json:"avatar"`
}

func (u *User) Validate(action string) error {
	if action == "store" {
		err := validation.ValidateStruct(u,
			// Name cannot be empty
			validation.Field(&u.Name, validation.Required, validation.Length(5, 64)),
			// Email cannot be empty
			validation.Field(&u.Email, validation.Required, is.Email),
			// Password cannot be empty
			validation.Field(&u.Password, validation.Required, validation.Length(6, 64)),
			// PasswordConfirmation cannot be empty
			validation.Field(&u.PasswordConfirmation, validation.Required, validation.Length(6, 64)),
		)
		return err
	} else if action == "update" {
		err := validation.ValidateStruct(u,
			// Name cannot be empty
			validation.Field(&u.Name, validation.Length(5, 64)),
			// Email cannot be empty
			validation.Field(&u.Email, is.Email),
			// Password cannot be empty
			validation.Field(&u.Password, validation.Length(6, 64), validation.When(
				u.PasswordConfirmation != "",
				validation.Required,
				validation.In(u.PasswordConfirmation).Error("Password and confirmation does not match"),
			)),
			// PasswordConfirmation cannot be empty
			validation.Field(&u.PasswordConfirmation, validation.Length(6, 64), validation.When(
				u.Password != "",
				validation.Required,
				validation.In(u.Password).Error("Password and confirmation does not match"),
			)),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
