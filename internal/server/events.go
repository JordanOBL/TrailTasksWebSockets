package server

type Event struct {
	Name            string
	Difficulty      int     // 2 bits: 01 for easy, 10 for medium, 11 for hard
	Duration        int     // Time in seconds or minutes required to pass the event
	CanPause        bool    // Can the user pause during this event?
	TokenPenalty    int     // Trail penalty (in distance lost or pace reduction)
	DistancePenalty float32 // Distance lost or pace reduction    // Distance penalty
	PacePenalty     int     // Pace penalty
	Reward          int     // Tokens or pace boost reward
}

var easyEvents = []Event{
	{
		Name:            "Deer Sighting",
		Difficulty:      1,
		Duration:        60, // 1 minute (60 seconds)
		CanPause:        true,
		TokenPenalty:    0, // No token penalty
		DistancePenalty: 0, // No distance penalty
		PacePenalty:     0, // No pace penalty
		Reward:          5, // 2 tokens
	},
	{
		Name:            "Scenic Rest Stop",
		Difficulty:      1,
		Duration:        60, // 1 minute (60 seconds)
		CanPause:        true,
		TokenPenalty:    0,
		DistancePenalty: 0,
		PacePenalty:     0,
		Reward:          2,
	},
	{
		Name:            "Light Rain",
		Difficulty:      1,
		Duration:        120, // 2 minutes (120 seconds)
		CanPause:        false,
		TokenPenalty:    1,   // 1 token lost if failed
		DistancePenalty: 0.1, // 0.1 miles lost if failed
		PacePenalty:     1,   // 1% pace reduction
		Reward:          3,   // 3 tokens
	},
	{
		Name:            "Rockslide",
		Difficulty:      1,
		Duration:        120, // 2 minutes (120 seconds)
		CanPause:        false,
		TokenPenalty:    1,
		DistancePenalty: 0.1,
		PacePenalty:     1,
		Reward:          3,
	},
	{
		Name:            "River Crossing",
		Difficulty:      1,
		Duration:        120, // 2 minutes (120 seconds)
		CanPause:        false,
		TokenPenalty:    1,
		DistancePenalty: 0.1,
		PacePenalty:     1,
		Reward:          2,
	},
}
var mediumEvents = []Event{
	{
		Name:            "Bear Encounter",
		Difficulty:      2,
		Duration:        300, // 5 minutes (300 seconds)
		CanPause:        false,
		TokenPenalty:    3,   // 3 tokens lost if failed
		DistancePenalty: 0.3, // 0.3 miles lost if failed
		PacePenalty:     2,   // 2% pace reduction
		Reward:          5,   // 5 tokens
	},
	{
		Name:            "Moderate Rain",
		Difficulty:      2,
		Duration:        240, // 4 minutes (240 seconds)
		CanPause:        false,
		TokenPenalty:    2,
		DistancePenalty: 0.2,
		PacePenalty:     2,
		Reward:          4,
	},
	{
		Name:            "Rockslide",
		Difficulty:      2,
		Duration:        240, // 4 minutes (240 seconds)
		CanPause:        false,
		TokenPenalty:    2,
		DistancePenalty: 0.2,
		PacePenalty:     2,
		Reward:          4,
	},
	{
		Name:            "River Crossing",
		Difficulty:      2,
		Duration:        180, // 3 minutes (180 seconds)
		CanPause:        false,
		TokenPenalty:    2,
		DistancePenalty: 0.2,
		PacePenalty:     2,
		Reward:          4,
	},
}
var hardEvents = []Event{
	{
		Name:            "Mountain Lion Standoff",
		Difficulty:      3,
		Duration:        600, // 10 minutes (600 seconds)
		CanPause:        false,
		TokenPenalty:    5,   // 5 tokens lost if failed
		DistancePenalty: 0.5, // 0.5 miles lost if failed
		PacePenalty:     5,   // 5% pace reduction
		Reward:          10,  // 10 tokens
	},
	{
		Name:            "Thunderstorm",
		Difficulty:      3,
		Duration:        420, // 7 minutes (420 seconds)
		CanPause:        false,
		TokenPenalty:    4,
		DistancePenalty: 0.3,
		PacePenalty:     5,
		Reward:          6,
	},
	{
		Name:            "Bear Encounter",
		Difficulty:      3,
		Duration:        480, // 8 minutes (480 seconds)
		CanPause:        false,
		TokenPenalty:    5,
		DistancePenalty: 0.5,
		PacePenalty:     4,
		Reward:          8,
	},
	{
		Name:            "River Crossing",
		Difficulty:      3,
		Duration:        300, // 5 minutes (300 seconds)
		CanPause:        false,
		TokenPenalty:    3,
		DistancePenalty: 0.3,
		PacePenalty:     3,
		Reward:          6,
	},
}
