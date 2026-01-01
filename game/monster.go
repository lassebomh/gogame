package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	Body   *cp.Body
	Arms   []*MonsterArm
	Radius float64
	Path   []cp.Vector
}

type MonsterArm struct {
	Monster               *Monster
	Path                  []cp.Vector
	SmoothedDistanceToTip float64
	MaxDistanceToTip      float64
	Segments              []*MonsterArmSegment
}

type MonsterArmSegment struct {
	Body             *cp.Body
	Shape            *cp.Shape
	RotaryLimitJoint *cp.Constraint

	Length float64
	Width  float64
}

const ARMS = 1
const ARM_SEGMENTS = 12

func (arm *MonsterArm) Update(w *World) {

	tip := arm.Segments[len(arm.Segments)-1]
	base := arm.Segments[0]

	distanceToTip := base.Body.Position().Sub(tip.Body.Position()).Length()

	if distanceToTip > arm.MaxDistanceToTip {
		arm.MaxDistanceToTip = distanceToTip
	}
	arm.SmoothedDistanceToTip += (distanceToTip - arm.SmoothedDistanceToTip) / 20

	playerPos := w.Player.Body.Position()

	if tip.Body.Position().Distance(arm.Path[0]) < w.Tilemap.Scale/2 {
		arm.Path = arm.Path[1:]
	}

	if len(arm.Path) != 0 {
		destination := arm.Path[len(arm.Path)-1]

		if destination.Distance(playerPos) > w.Tilemap.Scale*0.9 {
			arm.Path = []cp.Vector{}
		}
	}

	if len(arm.Path) == 0 {
		arm.Path = w.Tilemap.FindPath(tip.Body.Position(), playerPos)
		arm.Path = append(arm.Path[:max(len(arm.Path)-1, 0)], playerPos)
	}

	tipDiff := arm.Path[0].Sub(tip.Body.Position())
	tipDiffDist := tipDiff.Length()

	if tipDiffDist > 0 {
		tipDiff = tipDiff.Normalize()
		angleDiff := math.Atan2(tipDiff.Y, tipDiff.X) - tip.Body.Angle()
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}
		tip.Body.SetTorque(angleDiff * 100 * tip.Body.Mass())

		if arm.SmoothedDistanceToTip < arm.MaxDistanceToTip/2 {
			tip.Body.SetForce(tipDiff.Mult(500 * tip.Body.Mass()))
		} else {
			tip.Body.SetVelocityVector(tip.Body.Velocity().Mult(0.5))
		}
	}

	if arm.SmoothedDistanceToTip > arm.MaxDistanceToTip/2 {
		baseDiff := tip.Body.Position().Sub(base.Body.Position())
		baseDiffDist := baseDiff.Length()

		if baseDiffDist > 0 {
			baseDiff = baseDiff.Normalize()
			base.Body.SetForce(baseDiff.Mult(500 * base.Body.Mass()))
			angleDiff := math.Atan2(baseDiff.Y, baseDiff.X) - base.Body.Angle()
			for angleDiff > math.Pi {
				angleDiff -= 2 * math.Pi
			}
			for angleDiff < -math.Pi {
				angleDiff += 2 * math.Pi
			}
			base.Body.SetTorque(angleDiff * 100 * base.Body.Mass())
		}

	}

}

func (m *Monster) Update(w *World) {
	newVelocity := m.Body.Velocity().Mult(0.90)
	m.Body.SetVelocity(newVelocity.X, newVelocity.Y)

	for _, arm := range m.Arms {
		arm.Update(w)
	}
}

func NewMonster(w *World, position cp.Vector) *Monster {
	var group uint = 1

	monster := &Monster{
		Arms:   make([]*MonsterArm, 0),
		Radius: 2.5,
	}

	mass := monster.Radius * monster.Radius * 10
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, monster.Radius, cp.Vector{})))
	body.SetPosition(position)

	shape := w.Space.AddShape(cp.NewCircle(body, monster.Radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)
	shape.Filter.Group = group

	monster.Body = body

	for range ARMS {

		arm := &MonsterArm{
			Monster:  monster,
			Segments: make([]*MonsterArmSegment, 0),
			Path:     []cp.Vector{body.Position()},
		}
		monster.Arms = append(monster.Arms, arm)

		var prevBody *cp.Body

		for bodyI := range ARM_SEGMENTS {
			i := float64(bodyI)

			segment := &MonsterArmSegment{
				Length: monster.Radius * 0.6,
				Width:  monster.Radius / (1 + i/4),
			}
			arm.Segments = append(arm.Segments, segment)

			mass := segment.Length * segment.Width * 5

			pos := position.Add(cp.Vector{X: (i + 0.5) * segment.Length, Y: 0})
			segment.Body = w.Space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, segment.Length, segment.Width)))
			segment.Body.SetPosition(pos)

			segment.Shape = w.Space.AddShape(cp.NewBox(segment.Body, segment.Length, segment.Width, 0))
			segment.Shape.SetElasticity(0.5)
			segment.Shape.SetFriction(0.7)
			segment.Shape.Filter.Group = group

			if prevBody != nil {
				pivot := pos.Add(cp.Vector{X: -segment.Length / 2, Y: 0})
				constraint := w.Space.AddConstraint(cp.NewPivotJoint(prevBody, segment.Body, pivot))
				constraint.SetMaxForce(1000000)
				rotaryLimitAngle := rl.Pi / 4
				rotaryLimit := w.Space.AddConstraint(cp.NewRotaryLimitJoint(prevBody, segment.Body, -rotaryLimitAngle, rotaryLimitAngle))
				rotaryLimit.SetMaxForce(1000000)
				segment.RotaryLimitJoint = rotaryLimit
			} else {
				pivot := pos.Add(cp.Vector{X: -segment.Length / 2, Y: 0})
				constraint := w.Space.AddConstraint(cp.NewPivotJoint(monster.Body, segment.Body, pivot))
				constraint.SetMaxForce(1000000)
			}

			prevBody = segment.Body

		}

	}
	return monster
}

func (m *Monster) Render(w *World) {
	pos := m.Body.Position()

	rl.DrawSphereEx(rl.Vector3{X: float32(pos.X), Y: 1, Z: float32(pos.Y)}, float32(m.Radius), 8, 8, rl.Black)

	for _, arm := range m.Arms {
		for i, segment := range arm.Segments {
			pos := segment.Body.Position()
			rl.DrawSphereEx(rl.Vector3{X: float32(pos.X), Y: 1, Z: float32(pos.Y)}, float32(arm.Monster.Radius/(1.2+float64(i)/10)), 8, 8, rl.Black)
		}
	}
}
