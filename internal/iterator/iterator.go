package iterator

type void struct{}

type Value interface{}

type Accept func(Value) bool

type Collection interface {
	Each(Accept)
}

type Iterator struct {
	stop    chan void
	values  chan Value
	current Value
}

func BuildIterator(coll Collection) *Iterator {
	values := make(chan Value)
	stop := make(chan void)
	go func() {
		defer close(values)
		coll.Each(func(value Value) bool {
			select {
			case values <- value:
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

func (iter *Iterator) Value() Value {
	return iter.current
}

func (iter *Iterator) Stop() {
	close(iter.stop)
}

type Iterators []Iterator

func (iters Iterators) Each(accept Accept) {
	for _, iterator := range iters {
		for iterator.Next() {
			if !accept(iterator.Value()) {
				// TODO close the remaining iterators, yeah?
				return
			}
		}
	}
}
