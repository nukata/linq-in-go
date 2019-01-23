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


## Basics

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

Here `Any` is defined as an alias of `interface{}`.

```Go
type Any = interface{}
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

`Select` method is defined as follows:

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

Here `Enumerator` is just a function type defined as `func(func(interface{}))`.

```Go
type Enumerator func(yield func(element Any))
```

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
There is _no artificial_ data structure.


## Working with Files

Like C#'s LINQ, this LINQ works well with files.
Consider the following example:

```Go
package main

import (
	"fmt"
	. "github.com/nukata/linq-in-go/linq"
	"log"
	"os"
	"strings"
)

// Open a file and print its contents in upper case up to 10 lines.
func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		file.Close()
		if r := recover(); r != nil {
			log.Fatal("panic: ", r)
		}
	}()

	lines := From(file).Select(func(line interface{}) interface{} {
		return strings.ToUpper(line.(string))
	}).Take(10)

	lines(func(line interface{}) {
		fmt.Println(line)
	})
}
```

You can run it as follows:

```
$ go build read-in-linq.go
$ ./read-in-linq read-in-linq.go
PACKAGE MAIN

IMPORT (
	"FMT"
	. "GITHUB.COM/NUKATA/LINQ-IN-GO/LINQ"
	"LOG"
	"OS"
	"STRINGS"
)

$ 
```

The definition of `From` begins as follows:

```Go
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
```

There is only a typical loop of `bufio.Scanner` here.

It might be safe to say this LINQ makes a _natural_ use of the Go language,
however, not a trivial use of it.
See the definition of [Take](linq/linq.go#L130-L145) method.
It calls an _escape procedure_ (passed in the manner of Scheme's `call/cc`)
whose mechanism is implemented with `panic` and `recover` internally.

For more examples, see [linq_test.go](linq/linq_test.go).
