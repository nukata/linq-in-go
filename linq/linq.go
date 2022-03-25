// H24.11.28/R4.3.25 by SUZUKI Hisao

// Package linq implements "LINQ to Objects" in Go.
//
// See https://docs.microsoft.com/dotnet/api/system.linq.enumerable
//
package linq

import (
	"bufio"
	"container/list"
	"io"
)

// Enumerator represents a sequence abstractly.
// In fact, it is a higher order function that applies its function argument
// to each element of the sequence that Enumerator represents abstractly.
type Enumerator[T any] func(yield func(element T))

// ToList creates a list from the sequence which Enumerator represents.
func (loop Enumerator[T]) ToList() *list.List {
	result := list.New()
	loop(func(element T) {
		result.PushBack(element)
	})
	return result
}

// ToSlice creates a slice from the sequence which Enumerator represents.
func (loop Enumerator[T]) ToSlice() []T {
	lst := loop.ToList()
	result := make([]T, lst.Len())
	i := 0
	for element := lst.Front(); element != nil; element = element.Next() {
		result[i] = element.Value.(T)
		i++
	}
	return result
}

// Aggregate applies the binary function f to seed with each of elements
// e1, e2, ..., eN from loop, resulting in
// f(f(...f(f(seed, e1), e2), ...), eN).
func Aggregate[S any, T any](f func(S, T) S, seed S, loop Enumerator[T]) S {
	loop(func(element T) {
		seed = f(seed, element)
	})
	return seed
}

// AggregateWithExit is a variant of Aggregate.
// It supplies an "exit" argument to the function f.
// If f calls exit(x), the enumeration will terminate and x will be returned.
func AggregateWithExit[S any, T any](f func(S, T, func(S)) S,
	seed S, loop Enumerator[T]) S {
	loop.LoopWithExit(func(element T, _exit func()) {
		exit := func(x S) {
			seed = x
			_exit()
		}
		seed = f(seed, element, exit)
	})
	return seed
}

// tokenType represents a token to break the enumeration.
type tokenType int

// LoopWithExit calls f(element, exit) for each element of Enumerator.
// If f calls exit(), the enumeration will terminate.
func (loop Enumerator[T]) LoopWithExit(f func(T, func())) {
	var token tokenType
	defer func() {
		// Recover from the panic if it had been raised with panic(&token).
		r := recover()
		if r != nil && r != &token {
			panic(r)
		}
	}()
	exit := func() {
		panic(&token)
	}
	loop(func(element T) {
		f(element, exit)
	})
}

// Select creates an Enumerator which applies f to each of elements.
func Select[T any, R any](f func(T) R, loop Enumerator[T]) Enumerator[R] {
	return func(yield func(R)) {
		loop(func(element T) {
			value := f(element)
			yield(value)
		})
	}
}

// SelectMany creates an Enumerator which applies f to each of subsequences
// and concatenates them to a single flat sequence.
func SelectMany[T any, R any](f func(T) Enumerator[R],
	loop Enumerator[T]) Enumerator[R] {
	return func(yield func(R)) {
		loop(func(element T) {
			eachLoop := f(element)
			eachLoop(func(eachElement R) {
				yield(eachElement)
			})
		})
	}
}

// Where creates an Enumerator which selects elements by appling
// predicate to each of them.
func (loop Enumerator[T]) Where(predicate func(T) bool) Enumerator[T] {
	return func(yield func(T)) {
		loop(func(element T) {
			if predicate(element) {
				yield(element)
			}
		})
	}
}

// Take creates an Enumerator which takes the first n elements from
// the sequence.
func (loop Enumerator[T]) Take(n int) Enumerator[T] {
	return func(yield func(T)) {
		if n > 0 {
			i := 0
			loop.LoopWithExit(func(element T, exit func()) {
				i++
				yield(element)
				if i >= n {
					exit()
				}
			})
		}
	}
}

// TakeWhile creates an Enumerator which takes elements from the sequence
// until predicate applied to the element results in false.
func (loop Enumerator[T]) TakeWhile(predicate func(T) bool) Enumerator[T] {
	return func(yield func(T)) {
		loop.LoopWithExit(func(element T, exit func()) {
			if predicate(element) {
				yield(element)
			} else {
				exit()
			}
		})
	}
}

// Skip creates an Enumerator which skips the first n elements in the
// sequence.
func (loop Enumerator[T]) Skip(n int) Enumerator[T] {
	return func(yield func(T)) {
		i := 0
		loop(func(element T) {
			if i >= n {
				yield(element)
			} else {
				i++
			}
		})
	}
}

// SkipWhile creates an Enumerator which skip elements until predicate
// applied to the element results in false.
func (loop Enumerator[T]) SkipWhile(predicate func(T) bool) Enumerator[T] {
	return func(yield func(T)) {
		atHead := true
		loop(func(element T) {
			if atHead {
				if predicate(element) {
					return
				}
				atHead = false
			}
			yield(element)
		})
	}
}

// Concat concatenates two Enumerators loop and loop2.
func (loop Enumerator[T]) Concat(loop2 Enumerator[T]) Enumerator[T] {
	return func(yield func(T)) {
		body := func(element T) {
			yield(element)
		}
		loop(body)
		loop2(body)
	}
}

// Zip creates an Enumerator which enumerates loop1 and loop2 in step,
// applying f to each element pair.
func Zip[T any, U any, R any](f func(T, U) R,
	loop1 Enumerator[T], loop2 Enumerator[U]) Enumerator[R] {
	return func(yield func(R)) {
		dataChan := make(chan U)
		quitChan := make(chan bool, 1)
		defer close(quitChan)

		go sendForEach(loop2, quitChan, dataChan)
		loop1.LoopWithExit(func(element T, exit func()) {
			quitChan <- true
			element2, ok := <-dataChan
			if ok {
				value := f(element, element2)
				yield(value)
			} else { // run out of loop2
				exit()
			}
		})
	}
}

func sendForEach[U any](loop Enumerator[U],
	quitChan <-chan bool, dataChan chan<- U) {
	defer close(dataChan)

	loop.LoopWithExit(func(element U, exit func()) {
		_, ok := <-quitChan
		if ok {
			dataChan <- element
		} else {
			exit()
		}
	})
}

// Empty[T] returns an empty Enumerator[T].
func Empty[T any]() Enumerator[T] {
	return func(yield func(T)) {}
}

// Range creates an Enumerator which counts from start up to start+count-1.
func Range(start, count int) Enumerator[int] {
	end := start + count
	return func(yield func(int)) {
		for i := start; i < end; i++ {
			yield(i)
		}
	}
}

// Repeat creates an Enumerator which repeats element count times.
func Repeat[T any](element T, count int) Enumerator[T] {
	return func(yield func(T)) {
		for i := 0; i < count; i++ {
			yield(element)
		}
	}
}

// IntsFrom returns an infinite sequence of integers n, n+1, n+2, ...
func IntsFrom(n int) Enumerator[int] {
	return func(yield func(int)) {
		for i := n; ; i++ {
			yield(i)
		}
	}
}

// From creates an Enumerator from a slice.
func From[T ~[]E, E any](x T) Enumerator[E] {
	return func(yield func(E)) {
		for _, element := range x {
			yield(element)
		}
	}
}

// FromChan creates an Enumerator from a channel.
func FromChan[T ~(<-chan E), E any](x T) Enumerator[E] {
	return func(yield func(E)) {
		for e := range x {
			yield(e)
		}
	}
}

// FromString creates an Enumerator[rune] from a string.
func FromString[S ~string](x S) Enumerator[rune] {
	return func(yield func(rune)) {
		for _, c := range x {
			yield(c)
		}
	}
}

// FromList[T] creats an Enumerator[T] from a list.List.
func FromList[T any](x *list.List) Enumerator[T] {
	return func(yield func(T)) {
		for e := x.Front(); e != nil; e = e.Next() {
			yield(e.Value.(T))
		}
	}
}

// FromReader creats an Enumerator[string] from an io.Reader.
// The enumerator will yield each line of scanner.Text() and may panic with
// scanner.Err().
func FromReader(x io.Reader) Enumerator[string] {
	return func(yield func(string)) {
		scanner := bufio.NewScanner(x)
		for scanner.Scan() {
			yield(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}
}
