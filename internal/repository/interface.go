package repository

import "example.com/tinder/internal/model"

type PersonRepository interface {
	Upsert(person model.Person)
	Remove(name string) bool
	GetByName(name string) (model.Person, bool)
	ListAll() []model.Person
	ListByGender(gender model.Gender) []model.Person
	RedeemWantedDates(name string) (model.Person, bool)
}
