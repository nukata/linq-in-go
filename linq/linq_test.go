// H25.1/21 - H31.1/19 by SUZUKI Hisao

package linq

import (
	"container/list"
	"errors"
	. "fmt"
	"strings"
)

// Generate a sequence of integers from 1 to 10 and then select their squares.
// Compare this to the C# example found in
//   https://docs.microsoft.com/dotnet/api/system.linq.enumerable.range
func ExampleRange_select() {
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

	var seq Enumerator = func(yield func(Any)) {
		yield("a")
		yield("b")
		yield("c")
	}
	x = seq.ToSlice()
	Printf("%v\n", x)
	// Output:
	// []interface {}{7, 8, 9}
	// [a b c]
}

func ExampleEnumerator_Aggregate() {
	x := Range(1, 5).Aggregate(100, func(a, b Any) Any {
		return a.(int) * b.(int)
	}) // x = 100 * 1 * 2 * 3 * 4 * 5
	Printf("%v\n", x)
	// Output:
	// 12000
}

func ExampleEnumerator_AggregateWithExit() {
	seq := From([]Any{1, 2, errors.New("poi"), 3, 4, 5})
	x := seq.AggregateWithExit(100, func(a, b Any, exit func(Any)) Any {
		if bi, ok := b.(int); ok {
			return a.(int) * bi
		}
		exit(-1 * a.(int))
		return nil // dummy
	})
	Printf("%v\n", x)
	// Output:
	// -200
}

func ExampleEnumerator_LoopWithExit() {
	seq := From([]Any{1, 2, errors.New("poi"), 3, 4, 5})
	seq.LoopWithExit(
		func(e Any, exit func()) {
			if i, ok := e.(int); ok {
				if i == 4 {
					exit()
				}
				Println(i)
				Println(i * 100)
			} else {
				Println("---")
			}
		})
	// Output:
	// 1
	// 100
	// 2
	// 200
	// ---
	// 3
	// 300
}

func ExampleEnumerator_Select() {
	seq := Range(7, 3).Select(func(e Any) Any {
		return e.(int) + 100
	})
	seq(func(e Any) {
		Println(e)
	})
	// Output:
	// 107
	// 108
	// 109
}

func ExampleEnumerator_SelectMany() {
	type PetOwner struct {
		Name string
		Pets []string
	}
	owners := []PetOwner{
		PetOwner{"Taro", []string{"Koro", "Pochi", "Tama"}},
		PetOwner{"Jiro", []string{"Kuro", "Buchi", "Tora"}},
	}

	x := From(owners).SelectMany(func(e Any) Enumerator {
		return From(e.(PetOwner).Pets)
	}).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [Koro Pochi Tama Kuro Buchi Tora]
}

func ExampleEnumerator_Where() {
	x := Range(1, 10).Where(func(e Any) bool {
		i := e.(int)
		return i%2 == 0
	}).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [2 4 6 8 10]
}

func ExampleEnumerator_Take() {
	x := Range(1, 6).Take(3).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [1 2 3]
}

func ExampleEnumerator_TakeWhile() {
	x := Range(1, 6).TakeWhile(func(e Any) bool {
		i := e.(int)
		return i < 4
	}).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [1 2 3]
}

func ExampleEnumerator_Skip() {
	x := Range(1, 6).Skip(3).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [4 5 6]
}

func ExampleEnumerator_SkipWhile() {
	x := Range(1, 6).SkipWhile(func(e Any) bool {
		i := e.(int)
		return i < 4
	}).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [4 5 6]
}

func ExampleEnumerator_Concat() {
	x := Range(7, 5).Concat(Range(101, 9)).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [7 8 9 10 11 101 102 103 104 105 106 107 108 109]
}

func ExampleEnumerator_Zip() {
	aa := From([]int{3, 1, 4, 1, 5, 9})
	bb := From([]int{2, 7, 1, 8, 2, 8})
	x := aa.Zip(bb, func(a, b Any) Any {
		return a.(int) + b.(int)
	}).ToSlice()
	Printf("%v\n", x)

	aa = From([]int{3, 1, 4, 1, 5, 9})
	bb = From([]int{2, 7, 1})
	x = aa.Zip(bb, func(a, b Any) Any {
		return a.(int) + b.(int)
	}).ToSlice()
	Printf("%v\n", x)

	aa = From([]int{3, 1, 4})
	bb = From([]int{2, 7, 1, 8, 2, 8})
	x = aa.Zip(bb, func(a, b Any) Any {
		return a.(int) + b.(int)
	}).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [5 8 5 9 7 17]
	// [5 8 5]
	// [5 8 5]
}

func ExampleEmpty() {
	x := Empty().ToSlice()
	Printf("%v\n", x)
	// Output:
	// []
}

func ExampleRange() {
	x := Range(-3, 7).ToSlice()
	Printf("%v\n", x)
	// Output:
	// [-3 -2 -1 0 1 2 3]
}

func ExampleRange_selectWhereTake() {
	x := Range(1, 30).Select(func(i Any) Any {
		Print(i)
		return Sprintf("%d", i)
	}).Where(func(s Any) bool {
		Print("-")
		return strings.ContainsRune(s.(string), '3')
	}).Take(3).ToSlice()
	Printf("\n%v\n", x)
	// Output:
	// 1-2-3-4-5-6-7-8-9-10-11-12-13-14-15-16-17-18-19-20-21-22-23-
	// [3 13 23]
}

func ExampleRepeat() {
	x := Repeat("toi", 3).ToSlice()
	Printf("\n%v\n", x)
	// Output:
	// [toi toi toi]
}

func ExampleIntsFrom() {
	x := IntsFrom(-3).Take(7).ToSlice()
	Printf("\n%v\n", x)
	// Output:
	// [-3 -2 -1 0 1 2 3]
}

func ExampleFrom() {
	p := func(e Any) {
		Printf(" %#v", e)
	}

	reader := strings.NewReader(
		`A quick brown fox
jumps over
the lazy dogs.
`)
	From(reader)(p) // io.Reader
	Println()

	x := list.New()
	x.PushBack("Funa")
	x.PushBack("1-hachi")
	x.PushBack("2-hachi")
	From(x)(p) // *list.List
	Println()

	From("271828")(p) // string
	Println()

	From([]int{2, 7, 1, 8, 2, 8})(p) // slice
	Println()
	From([6]int{2, 7, 1, 8, 2, 8})(p) // array
	Println()

	From([]float64{2.7, 1, 8, 2, 8})(p)
	Println()
	From([]string{"27", "1", "8", "2", "8"})(p)
	Println()
	From([]interface{}{2, 7, "1", 8, 2, 8})(p)
	Println()

	ch := make(chan string)
	go func() {
		ch <- "Funa"
		ch <- "1-hachi"
		ch <- "2-hachi"
		close(ch)
	}()
	From(ch)(p) // chan
	Println()

	From(2.71828)(p)
	Println()
	// Output:
	//  "A quick brown fox" "jumps over" "the lazy dogs."
	//  "Funa" "1-hachi" "2-hachi"
	//  50 55 49 56 50 56
	//  2 7 1 8 2 8
	//  2 7 1 8 2 8
	//  2.7 1 8 2 8
	//  "27" "1" "8" "2" "8"
	//  2 7 "1" 8 2 8
	//  "Funa" "1-hachi" "2-hachi"
	//  2.71828
}

func ExampleEnumerator_fizzBuzz() {
	fizzbuzz := IntsFrom(1).Select(func(e Any) Any {
		i := e.(int)
		if i%3 == 0 {
			if i%5 == 0 {
				return "FizzBuzz"
			}
			return "Fizz"
		} else if i%5 == 0 {
			return "Buzz"
		}
		return i
	})

	fizzbuzz.Take(19)(func(e Any) {
		Print(e, " ")
	})
	Println()
	// Output:
	// 1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz 16 17 Fizz 19
}
