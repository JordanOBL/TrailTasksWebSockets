package server

type Timer struct {
	IsStarted     bool
	IsRunning     bool
	IsPaused      bool
	Time          int //seconds
	FocusTime     int //seconds
	BreakTime     int //seconds
	LongBreakTime int //seconds
}
