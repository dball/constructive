package database

import (
	"time"

	. "github.com/dball/constructive/pkg/types"
)

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

var SystemClock Clock = systemClock{}

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

func BuildFixedClock(now time.Time) fixedClock {
	return fixedClock{now}
}
