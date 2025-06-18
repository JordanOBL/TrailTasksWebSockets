package server

import (
	"math"
	"sync"
)

type Session struct {
	SessionMux            sync.RWMutex `json:"-"`
	Name                  string       `json:"name"`
	Distance              float64      `json:"distance"`
	Level                 uint8        `json:"level"`
	HighestCompletedLevel uint8        `json:"highestCompletedLevel"`
	Strikes               uint8        `json:"strikes"`
	TokensEarned          uint8        `json:"tokensEarned"`
	BonusTokens           uint8        `json:"bonusTokens"`
}

func (s *Session) Reset() {
	s.SessionMux.Lock()
	defer s.SessionMux.Unlock()
	s.Distance = 0.0
	s.Level = 0
	s.HighestCompletedLevel = 0
	s.Strikes = 0
	s.TokensEarned = 0
	s.BonusTokens = 0
}

func (s *Session) calculateTokensEarned(r *Room) int {
	s.SessionMux.Lock()
	defer s.SessionMux.Unlock()

	// Base reward: 1 token for every half mile completed
	baseTokens := int(math.Floor(s.Distance / 0.5))

	// Add any bonus tokens earned from events or extras
	totalTokens := baseTokens + int(s.BonusTokens)

	// Apply strike penalty to reduce total tokens
	penalty := int(float64(totalTokens) * s.calculateStrikePenalty())
	totalTokens -= penalty

	if totalTokens < 0 {
		totalTokens = 0
	}

	s.TokensEarned = uint8(totalTokens)

	return totalTokens
}

func (s *Session) calculateStrikePenalty() float64 {

	if s.Strikes < 3 {
		return 0.10
	} else if s.Strikes < 5 {
		return 0.30
	} else if s.Strikes < 7 {
		return 0.50
	} else if s.Strikes < 9 {
		return 0.70
	}

	return 1.00
}
