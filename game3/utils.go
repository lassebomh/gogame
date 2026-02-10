package game3

import (
	"fmt"
	"game/vec2"
	"game/vec3"
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

type v2 vec2.Value
type v3 vec3.Value
