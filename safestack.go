package safestack

import (
	"fmt"
	"sync"
)

type SafeStack[T any] struct {
	Items   []T
	mutex   sync.RWMutex
	Maxsize int
}

// NewSafeStack - the factory function; return a *SafeStack[T]
func NewSafeStack[T any](items []T) *SafeStack[T] {
	return &SafeStack[T]{
		Items:   items,
		mutex:   sync.RWMutex{},
		Maxsize: 0,
	}
}

// NewMax - set a new max stack size; trim to that size if necessary
func (s *SafeStack[T]) NewMax(n int) {
	s.Trim(n)
	s.mutex.Lock()
	s.Maxsize = n
	s.mutex.Unlock()
}

// Trim - drop the stack size down to n
func (s *SafeStack[T]) Trim(n int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if n < len(s.Items) {
		s.Items = s.Items[len(s.Items)-n : len(s.Items)]
	}
}

// RePopulate - insert a new slice into the stack; drop down to maxsize if necessary
// note that here the item order is the inverse of Clear() + PushMany(): FILO vs LIFO
func (s *SafeStack[T]) RePopulate(items []T) {
	s.mutex.Lock()
	s.Items = items
	s.mutex.Unlock()
	s.Trim(s.Maxsize)
}

// Push - add an item to the top of the stack; drop an item from the bottom if necessary
func (s *SafeStack[T]) Push(item T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Items = append(s.Items, item)
	if s.Maxsize != 0 && len(s.Items) > s.Maxsize {
		s.Items = s.Items[1:]
	}
}

// PushMany - add multiple items to the top of the stack; first in last out
// push [1, 2, 3] onto [0] -> stack [0, 1, 2, 3]
func (s *SafeStack[T]) PushMany(items []T) {
	for _, item := range items {
		s.Push(item)
	}
}

// Len - return the # of items in the stacK
func (s *SafeStack[T]) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.Items)
}

// Peek - look at the top item in the stack; but do not pop it
func (s *SafeStack[T]) Peek() (T, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var i T
	if len(s.Items) == 0 {
		return i, fmt.Errorf("empty stack")
	}

	i = s.Items[len(s.Items)-1]
	return i, nil
}

// Pop - pop the top item from the stack leaving it smaller by one
func (s *SafeStack[T]) Pop() (T, error) {
	i, e := s.Peek()
	if len(s.Items) == 0 {
		return i, e
	} else {
		s.Items = s.Items[:len(s.Items)-1]
		return i, e
	}
}

// AssumeSafePop - Pop() but brazenly assume that the stack is not empty
func (s *SafeStack[T]) AssumeSafePop() T {
	i, _ := s.Pop()
	return i
}

// Clear - empty the stack
func (s *SafeStack[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Items = []T{}
}

// PeekAll - return all items in the stack but leave the stack unchanged; last in first out
// push 1, 2, 3 -> stack [1, 2, 3] -> peek [3, 2, 1]
func (s *SafeStack[T]) PeekAll() []T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	li := len(s.Items)
	all := make([]T, li)

	for i := 0; i < li; i++ {
		all[(li-1)-i] = s.Items[i]
	}
	return all
}

// PeekAtSlice - return all items in the stack but leave the stack unchanged; first in last out
// push 1, 2, 3 -> stack [1, 2, 3] -> peek [1, 2, 3]
func (s *SafeStack[T]) PeekAtSlice() []T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Items
}

// PopAll - return all items in the stack and empty the stack; last in first out
// push 1, 2, 3 -> stack [1, 2, 3] -> pop [3, 2, 1]
func (s *SafeStack[T]) PopAll() []T {
	all := s.PeekAll()
	s.Clear()
	return all
}

// PopSlice - return all items in the stack and empty the stack; first in last out
// push 1, 2, 3 -> stack [1, 2, 3] -> pop [1, 2, 3]
func (s *SafeStack[T]) PopSlice() []T {
	all := s.PeekAtSlice()
	s.Clear()
	return all
}

// Reverse - invert the item order in the stack
func (s *SafeStack[T]) Reverse() {
	// PeekAll() reverses... Just need to write after the read.
	rev := s.PeekAll()
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.Items = rev
}
