package sequence

type Sequence[T any] struct {
	Payload  *T
	Next     *Sequence[T]
	Previous *Sequence[T]
}

func New[T any](payload *T) *Sequence[T] {
	return &Sequence[T]{
		Payload:  payload,
		Next:     nil,
		Previous: nil,
	}
}

func (s *Sequence[T]) InsertAfter(payload *T) *Sequence[T] {
	element := &Sequence[T]{
		Payload:  payload,
		Previous: s,
		Next:     nil,
	}
	s.Next = element
	return element
}

// Removes the current element and returns the previous element of the sequence, if such an element exits.
// Otherwise returns the next element if that exits. If this is the last element of the sequence returns a nil
func (s *Sequence[T]) Remove() *Sequence[T] {
	prev := s.Previous
	next := s.Next
	s.Next = nil
	s.Previous = nil
	if prev != nil {
		prev.Next = next
	}
	if next != nil {
		next.Previous = next
	}
	if prev != nil {
		return prev
	} else if next != nil {
		return next
	}
	return nil
}
