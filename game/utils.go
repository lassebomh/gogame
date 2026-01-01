package game

import "fmt"

func Debug(values ...any) {
	for _, value := range values {
		fmt.Printf("[%+v] ", value)
	}
	fmt.Print("\n")
}
