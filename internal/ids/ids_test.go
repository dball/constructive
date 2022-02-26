package ids

import (
	"testing"

	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Seq(t *testing.T) {
	t.Run("scalar", func(t *testing.T) {
		scalar := Scalar(ID(5))
		actual := []ID{}
		for id := range scalar.Seq().Values {
			actual = append(actual, id)
		}
		assert.Equal(t, []ID{5}, actual)
	})
	t.Run("set", func(t *testing.T) {
		set := Set{ID(5): Void{}, ID(7): Void{}}
		actual := []ID{}
		for id := range set.Seq().Values {
			actual = append(actual, id)
		}
		assert.ElementsMatch(t, []ID{5, 7}, actual)
	})
	t.Run("range", func(t *testing.T) {
		r := Range{Min: ID(1), Max: ID(20)}
		actual := []ID{}
		for id := range r.Seq().Values {
			actual = append(actual, id)
		}
		assert.Equal(t, []ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, actual)
	})
}
