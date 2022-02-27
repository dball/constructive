package ids

import (
	"github.com/dball/constructive/internal/iterator"
	. "github.com/dball/constructive/pkg/types"
)

type Constraint interface {
	Size() int
	Iterator() *iterator.Iterator
}

type Scalar ID
type Set map[ID]Void
type Range struct {
	Min ID
	Max ID
}

func (scalar Scalar) Each(accept iterator.Accept) {
	accept(ID(scalar))
}

func (scalar Scalar) Size() int {
	return 1
}

func (scalar Scalar) Iterator() *iterator.Iterator {
	return iterator.BuildIterator(scalar)
}

func (set Set) Each(accept iterator.Accept) {
	for id := range set {
		if !accept(id) {
			return
		}
	}
}

func (set Set) Size() int {
	return len(set)
}

func (set Set) Iterator() *iterator.Iterator {
	return iterator.BuildIterator(set)
}

func (r Range) Each(accept iterator.Accept) {
	for id := r.Min; id <= r.Max; id++ {
		if !accept(id) {
			return
		}
	}
}

func (r Range) Size() int {
	return int(r.Max - r.Min + 1)
}

func (r Range) Iterator() *iterator.Iterator {
	return iterator.BuildIterator(r)
}
