package scoring

import (
	"mariners/player"
	"mariners/team"
)

type Score struct {
	Player player.Player
	TeamID team.Team
	Scores [9]int
}

type MPAverage struct {
	Player  player.Player
	Average float64
	Last20  float64
}

func (mp *MPAverage) GetAverage() error {

	return nil
}
