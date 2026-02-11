package game3

import (
	v3 "game/vec3"

	"github.com/jakecoffman/cp"
)

type Door struct {
	Position v3.Value
	Angle    float64
	body     *cp.Body
	shape    *cp.Shape
}
