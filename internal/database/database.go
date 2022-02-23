package database

/*
func (idx *BTreeIndex) Lookup(ref LookupRef) ID {
	var A ID
	switch a := ref.A.(type) {
	case ID:
		A = a
	case LookupRef:
		A = idx.Lookup(a)
	case Ident:
		c := Constraints{
			A: ,
		}
	}
}

func (idx *BTreeIndex) Resolve(sel Selection) (c Constraints) {
	switch e := sel.E.(type) {
	case ID:
		c.E = IDSet{e: void}
		return c
	case LookupRef:

	case Ident:
	case ESet:
	case ERange:
	default:
	}
	switch sel.A.(type) {
	case ID:
	case LookupRef:
	case Ident:
	case Attr:
	case ASet:
	case ARange:
	default:
	}
	panic("TODO")
}
*/
