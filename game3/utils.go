package game3

import (
	"fmt"
)

func Printv(args ...any) {
	format := ""
	for i := range args {
		if i != 0 {
			format += " "
		}
		format += "%+v"
	}
	format += "\n"
	fmt.Printf(format, args...)
}

// type v2 v2.Value
// type v3 v3.Value
