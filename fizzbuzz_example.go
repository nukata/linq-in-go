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
