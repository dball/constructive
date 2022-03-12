package database

import (
	"time"

	"github.com/dball/constructive/pkg/types"
)

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

var SystemClock types.Clock = systemClock{}

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

func BuildFixedClock(now types.Inst) fixedClock {
	return fixedClock{time.Time(now)}
}
