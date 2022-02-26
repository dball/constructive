package ids

import (
	. "github.com/dball/constructive/pkg/types"
)

type Seq struct {
	Length int
	Close  chan<- Void
	Values <-chan ID
}

type Constraint interface {
	Seq() Seq
}

type Scalar ID
type Set map[ID]Void
type Range struct {
	min ID
	max ID
}

func (scalar Scalar) Seq() Seq {
	values := make(chan ID, 1)
	values <- ID(scalar)
	close(values)
	// TODO what is the empty value for channels?
	// TODO should this have a buffer of 1, or a drain?
	return Seq{Close: make(chan Void), Length: 1, Values: values}
}

func (set Set) Seq() Seq {
	l := len(set)
	values := make(chan ID, l)
	for id := range set {
		values <- id
	}
	close(values)
	return Seq{Close: make(chan Void), Length: l, Values: values}
}

func (r Range) Seq() Seq {
	values := make(chan ID, 16)
	cls := make(chan Void)
	go func() {
	OUTER:
		for id := r.min; id <= r.max; id++ {
			select {
			case values <- id:
			case <-cls:
				break OUTER
			}
		}
		close(values)
	}()
	l := r.max - r.min + 1
	return Seq{Close: cls, Length: int(l), Values: values}
}
