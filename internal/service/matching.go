package service

import (
	"errors"
	"sort"

	"example.com/tinder/internal/model"
	"example.com/tinder/internal/repository"
)

type matchingService struct {
	repo repository.PersonRepository
}

func NewMatchingService(repo repository.PersonRepository) MatchingService {
	return &matchingService{repo: repo}
}

func (s *matchingService) AddSinglePersonAndMatch(person model.Person) ([]model.MatchResult, error) {
	if err := person.Validate(); err != nil {
		return nil, err
	}
	if _, ok := s.repo.GetByName(person.Name); ok {
		return nil, errors.New("user already exists")
	}
	oppositeGender := model.GenderFemale
	if person.Gender == model.GenderFemale {
		oppositeGender = model.GenderMale
	}

	matches := make([]model.MatchResult, 0)
	current := &person

	opposites := s.repo.ListByGender(oppositeGender)

	for _, otherSnapshot := range opposites {
		if current.WantedDates <= 0 {
			break
		}

		if !canMatch(*current, otherSnapshot) {
			continue
		}

		updatedOther, success := s.repo.RedeemWantedDates(otherSnapshot.Name, oppositeGender)
		if !success {
			continue
		}

		current.WantedDates--

		res := model.MatchResult{BoyName: current.Name, GirlName: updatedOther.Name}
		if current.Gender == model.GenderFemale {
			res.BoyName, res.GirlName = updatedOther.Name, current.Name
		}
		matches = append(matches, res)
	}

	if current.WantedDates > 0 {
		s.repo.Upsert(*current)
	}

	return matches, nil
}

func (s *matchingService) RemoveSinglePerson(name string) bool {
	return s.repo.Remove(name)
}

func (s *matchingService) QuerySinglePeople() []model.CurrentMember {
	all := s.repo.ListAll()
	result := make([]model.CurrentMember, 0, len(all))
	for _, p := range all {
		result = append(result, model.CurrentMember{
			Name:           p.Name,
			Height:         p.Height,
			Gender:         p.Gender,
			RemainingDates: p.WantedDates,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (s *matchingService) QuerySinglePerson(name string) (*model.CurrentMember, bool) {
	p, ok := s.repo.GetByName(name)
	if !ok {
		return nil, false
	}
	return &model.CurrentMember{
		Name:           p.Name,
		Height:         p.Height,
		Gender:         p.Gender,
		RemainingDates: p.WantedDates,
	}, true
}

func (s *matchingService) QueryPersonMatches(name string, top int) ([]model.PersonMatchCandidate, bool) {
	if top <= 0 {
		return []model.PersonMatchCandidate{}, true
	}

	person, ok := s.repo.GetByName(name)
	if !ok {
		return nil, false
	}

	oppositeGender := model.GenderFemale
	if person.Gender == model.GenderFemale {
		oppositeGender = model.GenderMale
	}

	opposites := s.repo.ListByGender(oppositeGender)
	result := make([]model.PersonMatchCandidate, 0, len(opposites))

	for _, other := range opposites {
		if canMatch(person, other) {
			result = append(result, model.PersonMatchCandidate{
				Name:           other.Name,
				Height:         other.Height,
				Gender:         other.Gender,
				RemainingDates: other.WantedDates,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	if top > len(result) {
		top = len(result)
	}
	return result[:top], true
}

func canMatch(a, b model.Person) bool {
	if a.Gender == model.GenderMale {
		return a.Height > b.Height
	}
	return b.Height > a.Height
}
