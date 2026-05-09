package service

import "example.com/tinder/internal/model"

type MatchingService interface {
	AddSinglePersonAndMatch(person model.Person) ([]model.MatchResult, error)
	RemoveSinglePerson(name string) bool
	QuerySinglePeople() []model.CurrentMember
	QuerySinglePerson(name string) (*model.CurrentMember, bool)
	QueryPersonMatches(name string, top int) ([]model.PersonMatchCandidate, bool)
}
