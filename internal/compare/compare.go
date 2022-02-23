package compare

import (
	. "github.com/dball/constructive/pkg/types"
)

type Fn func(d1 Datum, d2 Datum) int

func E(d1 Datum, d2 Datum) int {
	switch {
	case d1.E < d2.E:
		return -1
	case d1.E > d2.E:
		return 1
	default:
		return 0
	}
}

func EA(d1 Datum, d2 Datum) int {
	e := E(d1, d2)
	if e != 0 {
		return e
	}
	return A(d1, d2)
}

func EAV(d1 Datum, d2 Datum) int {
	ea := EA(d1, d2)
	if ea != 0 {
		return ea
	}
	return Compare(d1.V, d2.V)
}

func A(d1 Datum, d2 Datum) int {
	switch {
	case d1.A < d2.A:
		return -1
	case d1.A > d2.A:
		return 1
	default:
		return 0
	}
}

func AE(d1 Datum, d2 Datum) int {
	a := A(d1, d2)
	if a != 0 {
		return a
	}
	return E(d1, d2)
}

func AEV(d1 Datum, d2 Datum) int {
	ae := AE(d1, d2)
	if ae != 0 {
		return ae
	}
	return Compare(d1.V, d2.V)
}

func AV(d1 Datum, d2 Datum) int {
	a := A(d1, d2)
	if a != 0 {
		return a
	}
	return Compare(d1.V, d2.V)
}

func AVE(d1 Datum, d2 Datum) int {
	av := AV(d1, d2)
	if av != 0 {
		return av
	}
	return E(d1, d2)
}
