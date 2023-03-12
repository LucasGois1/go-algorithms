package iterator

type Iterator[E any] interface {
	Iter() <-chan E
	Map(f func(E) interface{}) Collection[interface{}]
	Filter(f func(E) bool) Collection[E]
	ForEach(f func(E))
}

type Collection[E any] interface {
	Iterator[E]
	Append(element E)
	Remove(index int)
	IsEmpty() bool
	Size() uint16
}

type List[E any] struct {
	elements []E
}

func NewList[E any]() Collection[E] {
	return &List[E]{
		elements: make([]E, 0),
	}
}

func (l *List[E]) Iter() <-chan E {
	iterator := make(chan E)

	go func() {
		for _, element := range l.elements {
			iterator <- element
		}

		close(iterator)
	}()

	return iterator
}

func (l *List[E]) Map(f func(E) interface{}) Collection[interface{}] {
	collection := NewList[interface{}]()

	for entry := range l.Iter() {
		collection.Append(f(entry))
	}

	return collection
}

func (l *List[E]) Filter(f func(E) bool) Collection[E] {
	collection := NewList[E]()

	for entry := range l.Iter() {
		if f(entry) {
			collection.Append(entry)
		}
	}

	return collection
}

func (l *List[E]) ForEach(f func(E)) {
	for entry := range l.Iter() {
		f(entry)
	}
}

func (l *List[E]) Append(element E) {
	l.elements = append(l.elements, element)
}

func (l *List[E]) Remove(index int) {
	l.elements = append(l.elements[:index], l.elements[index+1:]...)
}

func (l *List[E]) IsEmpty() bool {
	return len(l.elements) == 0
}

func (l *List[E]) Size() uint16 {
	return uint16(len(l.elements))
}
