package server

type Timer struct {
	StartTime             string  `json:"startTime"` //???
	IsCompleted           bool    `json:"isCompleted"`
	Time                  uint16  `json:"time"`
	IsRunning             bool    `json:"isRunning"`
	IsBreak               bool    `json:"isBreak"`
	IsPaused              bool    `json:"isPaused"`
	InitialPomodoroTime   uint16  `json:"initialPomodoroTime"`
	InitialShortBreakTime uint16  `json:"initialShortBreakTime"`
	InitialLongBreakTime  uint16  `json:"initialLongBreakTime"`
	Sets                  uint8   `json:"sets"`
	CompletedSets         uint8   `json:"completedSets"`
	Pace                  float32 `json:"pace"`
	AutoContinue          bool    `json:"autoContinue"`
}
