package server

type Session struct {
	Name         string  `json:"name"`
	Distance     float64 `json:"distance"`
	Level        uint8   `json:"level"`
	highestLevel uint8   `json:"-"`
	Strikes      uint8   `json:"strikes"`
	TokensEarned uint8   `json:"tokensEarned"`
	BonusTokens  uint8   `json:"bonusTokens"`
}

func (s *Session) Reset() {
	s.Distance = 0.0
	s.Level = 0
	s.highestLevel = 0
	s.Strikes = 0
	s.TokensEarned = 0
	s.BonusTokens = 0
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
