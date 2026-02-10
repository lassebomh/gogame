package game3

import (
	v3 "game/vec3"

	"github.com/jakecoffman/cp"
)

type Object struct {
	Position v3.Value
	body     *cp.Body
	shape    *cp.Shape
}
