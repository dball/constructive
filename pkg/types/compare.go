package types

func typeCompare(v Value) int {
	switch v.(type) {
	case ID:
		return 1
	case Bool:
		return 2
	case Int:
		return 3
	case String:
		return 4
	case Inst:
		return 5
	default:
		panic(ErrInvalidValue)
	}
}

func Compare(v1 Value, v2 Value) int {
	if v1 == nil {
		if v2 == nil {
			return 0
		} else {
			return -1
		}
	} else if v2 == nil {
		return 1
	}
	switch x1 := v1.(type) {
	case ID:
		x2, ok := v2.(ID)
		if !ok {
			return typeCompare(x1) - typeCompare(x2)
		}
		switch {
		case x1 < x2:
			return -1
		case x1 > x2:
			return 1
		default:
			return 0
		}
	case Int:
		x2, ok := v2.(Int)
		if !ok {
			return typeCompare(x1) - typeCompare(x2)
		}
		switch {
		case x1 < x2:
			return -1
		case x1 > x2:
			return 1
		default:
			return 0
		}
	case String:
		x2, ok := v2.(String)
		if !ok {
			return typeCompare(x1) - typeCompare(x2)
		}
		switch {
		case x1 < x2:
			return -1
		case x1 > x2:
			return 1
		default:
			return 0
		}
	case Bool:
		x2, ok := v2.(Bool)
		if !ok {
			return typeCompare(x1) - typeCompare(x2)
		}
		if x1 == x2 {
			return 0
		}
		if x1 {
			return 1
		}
		return -1
	}
	panic("welp")
}
