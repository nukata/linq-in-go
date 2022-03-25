# LINQ in Go

This is an extensively revised version of "LINQ to Objects" in Go
which I had written and once presented at 
<http://www.oki-osk.jp/esc/golang/linq3.html> (now broken link) in 2014.
So it is not a newcomer, however, it seems still innovative:

- The sequences are **represented by functions** abstractly.

- LINQ methods are defined on functions directly.

- Thus, the sequences are inherently **lazy**.
  You can operate **infinite sequences** naturally.
  The code is concise and the space complexity is usually O(1).
  You can enjoy the best essence of C# LINQ in Go.

Now in 2022, Go 1.18 comes with generics.  You can get rid of almost every type
assersion such as `e.(int)`.  I have revised the LINQ in Go again.


## Let's try

```
$ go build
$ ./example
1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz 16 17 Fizz 19
$ 
```

Here you have run [`./fizzbuzz_example.go`](fizzbuzz_example.go):

```go
package main

import (
	"fmt"
	. "github.com/nukata/linq-in-go/linq"
)

func main() {
	fizzbuzz := Select(func(i int) any {
		if i%3 == 0 {
			if i%5 == 0 {
				return "FizzBuzz"
			}
			return "Fizz"
		} else if i%5 == 0 {
			return "Buzz"
		}
		return i
	}, IntsFrom(1))

	fizzbuzz.Take(19)(func(e any) {
		fmt.Print(e, " ")
	})
	fmt.Println()
	// Output:
	// 1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz 16 17 Fizz 19
}

```

You can run more examples and read the documentations:

```
$ cd linq
$ go test
PASS
ok  	github.com/nukata/linq-in-go/linq	0.326s
$ godoc -http=localhost:6060 &
```

Now open the web browser and see
[http://localhost:6060/pkg/github.com/nukata/linq-in-go/linq/](http://localhost:6060/pkg/github.com/nukata/linq-in-go/linq/).



## Basics

In the above examle, the function `IntsFrom` returns an `Enumerator[int]` value
and the variable `fizzbuzz` has an `Enumerator[any]` value.

The type `Enumerator[T]` is defined as follows:

```Go
// Enumerator represents a sequence abstractly.
// In fact, it is a higher order function that applies its function argument
// to each element of the sequence that Enumerator represents abstractly.
type Enumerator[T any] func(yield func(element T))
```

the function `IntsFrom` is defined as follows:

```Go
// IntsFrom returns an infinite sequence of integers n, n+1, n+2, ...
func IntsFrom(n int) Enumerator[int] {
	return func(yield func(int)) {
		for i := n; ; i++ {
			yield(i)
		}
	}
}
```

and the function `Select` is defined as follows:

```Go
// Select creates an Enumerator which applies f to each of elements.
func Select[T any, R any](f func(T) R, loop Enumerator[T]) Enumerator[R] {
	return func(yield func(R)) {
		loop(func(element T) {
			value := f(element)
			yield(value)
		})
	}
}
```

Therefore, the variable `fizzbuzz`, which are defined as

```Go
fizzbuzz := Select(func(i int) any {
	if i%3 == 0 {
		if i%5 == 0 {
			return "FizzBuzz"
		}
		return "Fizz"
	} else if i%5 == 0 {
		return "Buzz"
	}
	return i
}, IntsFrom(1))
```

is actually a higher order function that yields (or repeatedly calls its function
argument with an element of) FizzBuzz infinite sequence lazily.

For more examples, see [linq/linq_test.go](linq/linq_test.go).
