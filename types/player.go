package types

import "time"

type Player struct {
	Name             string
	Score            int
	CorrectQuestions []string
	LastUpdate       time.Time
	BuzzIn           time.Time
}
