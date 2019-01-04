# LINQ in Go

This is a revised implementation of "LINQ to Objects" in Go
which I wrote and once presented at 
<http://www.oki-osk.jp/esc/golang/linq3.html> in 2014.
So it is not a newcomer, however, it seems still innovative:

- The sequences are represented by functions abstractly.

- LINQ methods are defined on functions directly.

- Thus, the sequences are inherently lazy.
  You can treat infinite sequences naturally.
  The code is concise and the space complexity is usually O(1).
  You can enjoy the best essence of C# LINQ in Go.

The code example of C#'s 
[Enumerable.Range(Int32, Int32) Method](https://docs.microsoft.com/dotnet/api/system.linq.enumerable.range)
can be written in Go with this `linq` package as follows:

```Go
package main

import (
	"fmt"
	. "github.com/nukata/linq-in-go/linq"
)

// Generate a sequence of integers from 1 to 10 and then select their squares.
func main() {
	squares := Range(1, 10).Select(func(x Any) Any { return x.(int) * x.(int) })
	squares(func(num Any) {
		fmt.Println(num)
	})
}
```

If you save the code as `range_select_example.go`, you can run it as follows:

```
$ go get github.com/nukata/linq-in-go/linq
$ go run range_select_example.go
1
4
9
16
25
36
49
64
81
100
$ 
```

Note that `Any` is defined as an alias of `interface{}`.

`Select` method is defiend as follows:

```Go
// Select creates an Enumerator which applies f to each of elements.
func (loop Enumerator) Select(f func(Any) Any) Enumerator {
        return func(yield func(Any)) {
                loop(func(element Any) {
                        value := f(element)
                        yield(value)
                })
        }
}
```

Note that `Enumerator` is defined as `func(func(Any))`.

`Range` is defined as follows:

```Go
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
```


Now you have seen the _whole_ implementation of the example.
The space complexity is O(1) and you can `yield` values
infinitely if you want.
There is no artificial data structure.
