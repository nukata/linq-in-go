# LINQ in Go

This is a revised implementation of "LINQ to Objects" in Go
which I wrote and once presented at 
<http://www.oki-osk.jp/esc/golang/linq3.html> in 2014.
So it is not new, however, alas, it seems still innovative:

- The sequences are represented by functions abstractly.

- LINQ methods are defined on functions directly.

- Thus, the sequences are inherently lazy.
  You can treat infinite sequences naturally.
  The code is concise and the space complexity is usually O(1).
  You can enjoy the best essence of C# LINQ in Go.

## Examples

```Go
squares := Range(1, 10).Select(func(x Any) Any { return x.(int) * x.(int) })
squares(func(num Any) {
        Println(num)
})
// Output:
// 1
// 4
// 9
// 16
// 25
// 36
// 49
// 64
// 81
// 100
```

Note that `type Any = interface{}`.

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

Note that `Enumerator` is defined as `type Enumerator func(func(Any))`.

Now you have seen the whole definition of `Select` method here.
It is functional, reactive and infinite.
