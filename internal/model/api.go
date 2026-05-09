package model

type AddSinglePersonAndMatchRequest struct {
	Name        string `json:"name"`
	Height      int    `json:"height"`
	Gender      Gender `json:"gender"`
	WantedDates int    `json:"wanted_dates"`
}

type MatchResult struct {
	BoyName  string `json:"boy_name"`
	GirlName string `json:"girl_name"`
}

type AddSinglePersonAndMatchResponse struct {
	Matches []MatchResult `json:"matches"`
}

type QuerySinglePeopleResponse struct {
	People []CurrentMember `json:"people"`
}

type QueryPersonMatchesResponse struct {
	Name    string                 `json:"name"`
	Matches []PersonMatchCandidate `json:"matches"`
}

type PersonMatchCandidate struct {
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Gender         Gender `json:"gender"`
	RemainingDates int    `json:"remaining_dates"`
}

type CurrentMember struct {
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Gender         Gender `json:"gender"`
	RemainingDates int    `json:"remaining_dates"`
}
