// H24.11/28 - H31.1/22 by SUZUKI Hisao

// Package linq implements "LINQ to Objects" in Go.
//
// cf. https://docs.microsoft.com/dotnet/api/system.linq.enumerable
//
package linq

import (
	"bufio"
	"container/list"
	"io"
	"reflect"
)

// Any represents an element of a sequence.
type Any = interface{}

// Enumerator represents a sequence.
// To be precise, it is a higher order function that applies the function
// argument to each element of the sequence which Enumerator represents.
type Enumerator func(yield func(element Any))

// ToList creates a list from the sequence which Enumerator represents.
func (loop Enumerator) ToList() *list.List {
	result := list.New()
	loop(func(element Any) {
		result.PushBack(element)
	})
	return result
}

// ToSlice creates a slice from the sequence which Enumerator represents.
func (loop Enumerator) ToSlice() []Any {
	lst := loop.ToList()
	result := make([]Any, lst.Len())
	i := 0
	for element := lst.Front(); element != nil; element = element.Next() {
		result[i] = element.Value
		i++
	}
	return result
}

// Aggregate applies the given binary function f to seed with each of
// elements e1, e2, ..., eN in the squence, resulting in
// f(f(...f(f(seed, e1), e2), ...), eN).
func (loop Enumerator) Aggregate(seed Any, f func(Any, Any) Any) Any {
	loop(func(element Any) {
		seed = f(seed, element)
	})
	return seed
}

// AggregateWithExit is a variant of Aggregate.
// It gives an "exit" argument to the given function f.
// If f calls exit(x), the enumeration will terminate and x will be returned.
func (loop Enumerator) AggregateWithExit(seed Any,
	f func(Any, Any, func(Any)) Any) Any {
	loop.LoopWithExit(func(element Any, rawExit func()) {
		exit := func(x Any) {
			seed = x
			rawExit()
		}
		seed = f(seed, element, exit)
	})
	return seed
}

// tokenT represents a token to break the enumeration.
type tokenT int

// recoverAsBreak makes the program recover from the panic which
// had been raised with panic(&token).
func recoverAsBreak(token *tokenT) {
	r := recover()
	if r != nil && r != token {
		panic(r)
	}
}

// LoopWithExit calls f(element, exit) for each element of Enumerator.
// If f calls exit(), the enumeration will terminate.
func (loop Enumerator) LoopWithExit(f func(Any, func())) {
	var token tokenT
	defer recoverAsBreak(&token)
	exit := func() {
		panic(&token)
	}
	loop(func(element Any) {
		f(element, exit)
	})
}

// Select creates an Enumerator which applies f to each of elements.
func (loop Enumerator) Select(f func(Any) Any) Enumerator {
	return func(yield func(Any)) {
		loop(func(element Any) {
			value := f(element)
			yield(value)
		})
	}
}

// SelectMany creates an Enumerator which applies f to each of subsequences
// and concatenates them to a single flat sequence.
func (loop Enumerator) SelectMany(f func(Any) Enumerator) Enumerator {
	return func(yield func(Any)) {
		loop(func(element Any) {
			loopOfLoop := f(element)
			loopOfLoop(func(elementOfElement Any) {
				yield(elementOfElement)
			})
		})
	}
}

// Where creates an Enumerator which selects elements by appling
// predicate to each of them.
func (loop Enumerator) Where(predicate func(Any) bool) Enumerator {
	return func(yield func(Any)) {
		loop(func(element Any) {
			if predicate(element) {
				yield(element)
			}
		})
	}
}

// Take creates an Enumerator which takes the first n elements from
// the sequence.
func (loop Enumerator) Take(n int) Enumerator {
	return func(yield func(Any)) {
		if n > 0 {
			i := 0
			loop.LoopWithExit(func(element Any, exit func()) {
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
func (loop Enumerator) TakeWhile(predicate func(Any) bool) Enumerator {
	return func(yield func(Any)) {
		loop.LoopWithExit(func(element Any, exit func()) {
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
func (loop Enumerator) Skip(n int) Enumerator {
	return func(yield func(Any)) {
		i := 0
		loop(func(element Any) {
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
func (loop Enumerator) SkipWhile(predicate func(Any) bool) Enumerator {
	return func(yield func(Any)) {
		atHead := true
		loop(func(element Any) {
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
func (loop Enumerator) Concat(loop2 Enumerator) Enumerator {
	return func(yield func(Any)) {
		body := func(element Any) {
			yield(element)
		}
		loop(body)
		loop2(body)
	}
}

// Zip creates an Enumerator which enumerates loop and loop2 in step,
// applying f to each element pair.
func (loop Enumerator) Zip(loop2 Enumerator, f func(Any, Any) Any) Enumerator {
	return func(yield func(Any)) {
		dataChan := make(chan Any)
		quitChan := make(chan Any, 1)
		defer close(quitChan)

		go sendForEach(loop2, quitChan, dataChan)
		loop.LoopWithExit(func(element Any, exit func()) {
			quitChan <- nil
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

func sendForEach(loop Enumerator, quitChan <-chan Any, dataChan chan<- Any) {
	defer close(dataChan)

	loop.LoopWithExit(func(element Any, exit func()) {
		_, ok := <-quitChan
		if ok {
			dataChan <- element
		} else {
			exit()
		}
	})
}

// Empty returns an empty sequence.
func Empty() Enumerator {
	return func(yield func(Any)) {}
}

// Range creates an Enumerator which counts from start
// up to start + count - 1.
func Range(start, count int) Enumerator {
	end := start + count
	return func(yield func(Any)) {
		for i := start; i < end; i++ {
			yield(i)
		}
	}
}

// Repeat creates an Enumerator which repeats element count times.
func Repeat(element Any, count int) Enumerator {
	return func(yield func(Any)) {
		for i := 0; i < count; i++ {
			yield(element)
		}
	}
}

// IntsFrom returns an infinite sequence of integers n, n+1, n+2, ...
func IntsFrom(n int) Enumerator {
	return func(yield func(Any)) {
		for i := n; ; i++ {
			yield(i)
		}
	}
}

// From creates an Enumerator from an argument.
// If the argument is one of io.Reader, *list.List, string, slice, array
// or chan, the Enumerator will yield each element of the argument.
// Otherwise it will yield the whole argument as its sole element.
// For io.Reader, it will yield each line as a string of (*Scanner) Text()
// and may panic with (*Scanner) Err().
func From(x interface{}) Enumerator {
	switch seq := x.(type) {
	case io.Reader:
		return func(yield func(Any)) {
			scanner := bufio.NewScanner(seq)
			for scanner.Scan() {
				yield(scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				panic(err)
			}
		}
	case *list.List:
		return func(yield func(Any)) {
			for e := seq.Front(); e != nil; e = e.Next() {
				yield(e.Value)
			}
		}
	case string:
		return func(yield func(Any)) {
			for _, value := range seq {
				yield(value)
			}
		}
	default:
		v := reflect.ValueOf(x)
		k := v.Kind()
		if k == reflect.Slice || k == reflect.Array {
			return func(yield func(Any)) {
				len := v.Len()
				for i := 0; i < len; i++ {
					e := v.Index(i)
					yield(e.Interface())
				}
			}
		} else if k == reflect.Chan {
			return func(yield func(Any)) {
				for {
					value, ok := v.Recv()
					if !ok {
						break
					}
					yield(value)
				}
			}
		}
	}
	return func(yield func(Any)) {
		yield(x)
	}
}
