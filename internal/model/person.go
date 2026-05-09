package model

import "errors"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type Person struct {
	Name        string `json:"name"`
	Height      int    `json:"height"`
	Gender      Gender `json:"gender"`
	WantedDates int    `json:"wanted_dates"`
}

func (p Person) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.Height <= 0 {
		return errors.New("height must be positive")
	}
	if p.Gender != GenderMale && p.Gender != GenderFemale {
		return errors.New("gender must be male or female")
	}
	if p.WantedDates <= 0 {
		return errors.New("wanted_dates must be positive")
	}
	return nil
}
