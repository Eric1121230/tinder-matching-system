package repository

import (
	"sync"

	"example.com/tinder/internal/model"
)

type InMemoryPersonRepository struct {
	maleMu sync.RWMutex
	males  map[string]model.Person

	femaleMu sync.RWMutex
	females  map[string]model.Person
}

func NewInMemoryPersonRepository() *InMemoryPersonRepository {
	return &InMemoryPersonRepository{
		males:   make(map[string]model.Person),
		females: make(map[string]model.Person),
	}
}

func (r *InMemoryPersonRepository) Upsert(person model.Person) {
	if person.Gender == model.GenderMale {
		r.maleMu.Lock()
		defer r.maleMu.Unlock()
		r.males[person.Name] = person
	} else {
		r.femaleMu.Lock()
		defer r.femaleMu.Unlock()
		r.females[person.Name] = person
	}
}
func (r *InMemoryPersonRepository) RedeemWantedDates(name string, gender model.Gender) (model.Person, bool) {
	var mu *sync.RWMutex
	var pool map[string]model.Person

	if gender == model.GenderMale {
		mu, pool = &r.maleMu, r.males
	} else if gender == "female" {
		mu, pool = &r.femaleMu, r.females
	} else {
		return model.Person{}, false
	}

	mu.Lock()
	defer mu.Unlock()

	p, ok := pool[name]
	if !ok || p.WantedDates <= 0 {
		return model.Person{}, false
	}

	p.WantedDates--
	if p.WantedDates <= 0 {
		delete(pool, name)
	} else {
		pool[name] = p
	}

	return p, true
}

func (r *InMemoryPersonRepository) Remove(name string) bool {
	r.maleMu.Lock()
	if _, ok := r.males[name]; ok {
		delete(r.males, name)
		r.maleMu.Unlock()
		return true
	}
	r.maleMu.Unlock()

	r.femaleMu.Lock()
	if _, ok := r.females[name]; ok {
		delete(r.females, name)
		r.femaleMu.Unlock()
		return true
	}
	r.femaleMu.Unlock()

	return false
}

func (r *InMemoryPersonRepository) GetByName(name string) (model.Person, bool) {
	r.maleMu.RLock()
	if p, ok := r.males[name]; ok {
		r.maleMu.RUnlock()
		return p, true
	}
	r.maleMu.RUnlock()

	r.femaleMu.RLock()
	defer r.femaleMu.RUnlock()
	p, ok := r.females[name]
	return p, ok
}

func (r *InMemoryPersonRepository) ListAll() []model.Person {
	r.maleMu.RLock()
	r.femaleMu.RLock()
	defer r.femaleMu.RUnlock()
	defer r.maleMu.RUnlock()

	result := make([]model.Person, 0, len(r.males)+len(r.females))
	for _, p := range r.males {
		result = append(result, p)
	}
	for _, p := range r.females {
		result = append(result, p)
	}
	return result
}

func (r *InMemoryPersonRepository) ListByGender(gender model.Gender) []model.Person {
	if gender == model.GenderMale {
		r.maleMu.RLock()
		defer r.maleMu.RUnlock()
		return r.mapToSlice(r.males)
	}
	r.femaleMu.RLock()
	defer r.femaleMu.RUnlock()
	return r.mapToSlice(r.females)
}

func (r *InMemoryPersonRepository) mapToSlice(m map[string]model.Person) []model.Person {
	res := make([]model.Person, 0, len(m))
	for _, p := range m {
		res = append(res, p)
	}
	return res
}
