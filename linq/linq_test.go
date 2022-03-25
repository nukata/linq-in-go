// H25.1.21/R4.3.25 by SUZUKI Hisao

package linq

import (
	"container/list"
	"errors"
	. "fmt"
	"strings"
)

// Generate a sequence of integers from 1 to 10.
func ExampleRange() {
	var loop Enumerator[int] = Range(1, 10)
	loop(func(num int) {
		Println(num)
	})
	//Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
	// 10
}

// Generate a sequence of integers from 1 to 10 and then select their squares.
// Compare this to the C# example found in
//   https://docs.microsoft.com/dotnet/api/system.linq.enumerable.range
func ExampleRange_select() {
	squares := Select(func(x int) int { return x * x }, Range(1, 10))
	squares(func(num int) {
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
}

func ExampleEnumerator_ToList() {
	x := Range(7, 3).ToList()
	for e := x.Front(); e != nil; e = e.Next() {
		Println(e.Value)
	}
	// Output:
	// 7
	// 8
	// 9
}

func ExampleEnumerator_ToSlice() {
	x := Range(7, 3).ToSlice()
	Printf("%#v\n", x)

	var seq Enumerator[string] = func(yield func(string)) {
		yield("a")
		yield("b")
		yield("c")
	}
	y := seq.ToSlice()
	Printf("%v\n", y)
	// Output:
	// []int{7, 8, 9}
	// [a b c]
}

func ExampleAggregate() {
	x := Aggregate(func(a, b int) int { return a * b }, 100, Range(1, 5))
	// x = 100 * 1 * 2 * 3 * 4 * 5
	Printf("%v\n", x)
	// Output:
	// 12000
}

func ExampleAggregateWithExit() {
	seq := From([]any{1, 2, 3, errors.New("poi"), 4, 5})
	x := AggregateWithExit(func(a int, b any, exit func(int)) int {
		if bi, ok := b.(int); ok {
			return a * bi
		}
		exit(-1 * a)
		return 0 // dummy
	}, 100, seq)
	Printf("%v\n", x)
	// Output:
	// -600
}

func ExampleEnumerator_LoopWithExit() {
	seq := From([]any{1, 2, 3, errors.New("poi"), 4, 5})
	seq.LoopWithExit(
		func(e any, exit func()) {
			if i, ok := e.(int); ok {
				Println(i)
				Println(i * 100)
			} else {
				Println("---")
				exit()
			}
		})
	// Output:
	// 1
	// 100
	// 2
	// 200
	// 3
	// 300
	// ---
}

func ExampleSelect() {
	seq := Select(func(e int) int { return e + 100 }, From([]int{7, 8, 9}))
	seq(func(e int) {
		Println(e)
	})
	// Output:
	// 107
	// 108
	// 109
}

func ExampleSelectMany() {
	type PetOwner struct {
		Name string
		Pets []string
	}
	owners := []PetOwner{
		PetOwner{"Taro", []string{"Koro", "Pochi", "Tama"}},
		PetOwner{"Jiro", []string{"Kuro", "Buchi", "Tora"}},
	}

	x := SelectMany(func(e PetOwner) Enumerator[string] {
		return From(e.Pets)
	}, From(owners))
	Printf("%v\n", x.ToSlice())
	// Output:
	// [Koro Pochi Tama Kuro Buchi Tora]
}

func ExampleEnumerator_Where() {
	x := Range(1, 10).Where(func(e int) bool { return e%2 == 0 })
	Printf("%v\n", x.ToSlice())
	// Output:
	// [2 4 6 8 10]
}

func ExampleEnumerator_Take() {
	x := Range(1, 6).Take(3)
	Printf("%v\n", x.ToSlice())
	// Output:
	// [1 2 3]
}

func ExampleEnumerator_TakeWhile() {
	x := Range(1, 6).TakeWhile(func(e int) bool { return e < 4 })
	Printf("%v\n", x.ToSlice())
	// Output:
	// [1 2 3]
}

func ExampleEnumerator_Skip() {
	x := Range(1, 6).Skip(3)
	Printf("%v\n", x.ToSlice())
	// Output:
	// [4 5 6]
}

func ExampleEnumerator_SkipWhile() {
	x := Range(1, 6).SkipWhile(func(e int) bool { return e < 4 })
	Printf("%v\n", x.ToSlice())
	// Output:
	// [4 5 6]
}

func ExampleEnumerator_Concat() {
	x := Range(7, 5).Concat(Range(101, 9))
	Printf("%v\n", x.ToSlice())
	// Output:
	// [7 8 9 10 11 101 102 103 104 105 106 107 108 109]
}

func ExampleZip() {
	aa := From([]int{3, 1, 4, 1, 5, 9})
	bb := From([]int{2, 7, 1, 8, 2, 8})
	x := Zip(func(a, b int) int { return a + b }, aa, bb)
	Printf("%v\n", x.ToSlice())

	aa = From([]int{3, 1, 4, 1, 5, 9})
	bb = From([]int{2, 7, 1})
	x = Zip(func(a, b int) int { return a + b }, aa, bb)
	Printf("%v\n", x.ToSlice())

	aa = From([]int{3, 1, 4})
	bb = From([]int{2, 7, 1, 8, 2, 8})
	x = Zip(func(a, b int) int { return a + b }, aa, bb)
	Printf("%v\n", x.ToSlice())
	// Output:
	// [5 8 5 9 7 17]
	// [5 8 5]
	// [5 8 5]
}

func ExampleEmpty() {
	x := Empty[int]()
	Printf("%v\n", x.ToSlice())
	// Output:
	// []
}

func ExampleRepeat() {
	x := Repeat("toi", 3)
	Printf("%v\n", x.ToSlice())
	// Output:
	// [toi toi toi]
}

func ExampleIntsFrom() {
	x := IntsFrom(-3).Take(7)
	Printf("%v\n", x.ToSlice())
	// Output:
	// [-3 -2 -1 0 1 2 3]
}

func ExampleIntsFrom_selectWhereTake() {
	x := Select(func(i int) string {
		Print(i)
		return Sprintf("%d", i)
	}, IntsFrom(1)).Where(func(s string) bool {
		Print("-")
		return strings.ContainsRune(s, '3')
	}).Take(3)
	Printf("\n%v\n", x.ToSlice())
	// Output:
	// 1-2-3-4-5-6-7-8-9-10-11-12-13-14-15-16-17-18-19-20-21-22-23-
	// [3 13 23]
}

func ExampleFrom() {
	loop := From([]int{2, 7, 1, 8})
	loop(func(num int) { Println(num) })
	// Output:
	// 2
	// 7
	// 1
	// 8
}

func ExampleFromChan() {
	ch := make(chan string)
	go func() {
		ch <- "Funa"
		ch <- "1-hachi"
		ch <- "2-hachi"
		close(ch)
	}()
	loop := FromChan((<-chan string)(ch))
	loop(func(s string) { Println(s) })
	// Output:
	// Funa
	// 1-hachi
	// 2-hachi
}

func ExampleFromString() {
	loop := FromString("2718")
	loop(func(ch rune) { Printf("%c\n", ch) })
	// Output:
	// 2
	// 7
	// 1
	// 8
}

func ExampleFromList() {
	x := list.New()
	x.PushBack("Funa")
	x.PushBack("1-hachi")
	x.PushBack("2-hachi")
	loop := FromList[string](x)
	loop(func(s string) { Println(s) })
	// Output:
	// Funa
	// 1-hachi
	// 2-hachi
}

func ExampleFromReader() {
	reader := strings.NewReader(
		`A quick brown fox
jumps over
the lazy dog.
`)
	loop := FromReader(reader)
	loop(func(s string) { Printf("%q\n", s) })
	// Output:
	// "A quick brown fox"
	// "jumps over"
	// "the lazy dog."
}

func ExampleEnumerator_fizzBuzz() {
	var fizzbuzz Enumerator[any] = Select(func(i int) any {
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
		Print(e, " ")
	})
	Println()
	// Output:
	// 1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz 16 17 Fizz 19
}
