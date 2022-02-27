package ids

import (
	. "github.com/dball/constructive/pkg/types"
)

type Accept func(ID) bool

type Seq interface {
	Each(Accept)
}

type Iterator struct {
	stop    chan Void
	values  chan ID
	current ID
}

func BuildIterator(seq Seq) *Iterator {
	values := make(chan ID)
	stop := make(chan Void)
	go func() {
		defer close(values)
		seq.Each(func(id ID) bool {
			select {
			case values <- id:
				return true
			case <-stop:
				return false
			}
		})
	}()
	return &Iterator{stop: stop, values: values}
}

func (iter *Iterator) Next() (ok bool) {
	iter.current, ok = <-iter.values
	return
}

func (iter *Iterator) Value() ID {
	return iter.current
}

func (iter *Iterator) Stop() {
	close(iter.stop)
}

type Constraint interface {
	Size() int
	Iterator() *Iterator
}

type Scalar ID
type Set map[ID]Void
type Range struct {
	Min ID
	Max ID
}

func (scalar Scalar) Each(accept Accept) {
	accept(ID(scalar))
}

func (scalar Scalar) Size() int {
	return 1
}

func (scalar Scalar) Iterator() *Iterator {
	return BuildIterator(scalar)
}

func (set Set) Each(accept Accept) {
	for id := range set {
		if !accept(id) {
			return
		}
	}
}

func (set Set) Size() int {
	return len(set)
}

func (set Set) Iterator() *Iterator {
	return BuildIterator(set)
}

func (r Range) Each(accept Accept) {
	for id := r.Min; id <= r.Max; id++ {
		if !accept(id) {
			return
		}
	}
}

func (r Range) Size() int {
	return int(r.Max - r.Min + 1)
}

func (r Range) Iterator() *Iterator {
	return BuildIterator(r)
}
