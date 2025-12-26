package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	body   *cp.Body
	Arms   []MonsterArm
	radius float64
}

type MonsterArm struct {
	Monster *Monster
	Bodies  []*cp.Body
	Shapes  []*cp.Shape
	Path    []cp.Vector
}

func (arm *MonsterArm) Update(w *World) {
	tipPos := arm.Bodies[len(arm.Bodies)-1].Position()
	playerPos := w.Player.Body.Position()

	if tipPos.Distance(arm.Path[0]) < w.Tilemap.Scale/2 {
		arm.Path = arm.Path[1:]
	}

	if len(arm.Path) != 0 {
		destination := arm.Path[len(arm.Path)-1]

		if destination.Distance(playerPos) > w.Tilemap.Scale*0.9 {
			arm.Path = []cp.Vector{}
		}
	}

	if len(arm.Path) == 0 {
		arm.Path = w.Tilemap.FindPath(tipPos, playerPos)
		arm.Path = append(arm.Path[:len(arm.Path)-1], playerPos)
	}

}

func (m *Monster) Update(w *World) {
	newVelocity := m.body.Velocity().Mult(0.95)
	m.body.SetVelocity(newVelocity.X, newVelocity.Y)

	for i := range m.Arms {
		arm := &m.Arms[i]
		arm.Update(w)

		// targetAngleOffset := float32(i) / float32(len(m.Arms)-1) / 2 * 0.1
		arm.TargetPoint(w, 0, arm.Path[0])

	}
}

func (arm *MonsterArm) TargetPoint(w *World, offsetAngle float32, point cp.Vector) {
	// Get the last segment
	lastSegment := arm.Bodies[len(arm.Bodies)-1]
	lastPos := lastSegment.Position()

	// Calculate direction to target
	direction := point.Sub(lastPos)
	distance := direction.Length()

	if distance > 0 {
		direction = direction.Normalize()
		desiredAngle := math.Atan2(direction.Y, direction.X) + float64(offsetAngle)
		direction = cp.Vector{X: math.Cos(desiredAngle), Y: math.Sin(desiredAngle)}

		// Apply force toward the target point
		force := direction.Mult(50000) // Adjust force strength as needed
		lastSegment.SetForce(force)

		// Calculate desired angle (pointing toward target)
		currentAngle := lastSegment.Angle()

		// Normalize angle difference to -pi to pi
		angleDiff := desiredAngle - currentAngle
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		// Apply torque to rotate toward target
		torque := angleDiff * 10000 // Adjust torque strength as needed
		lastSegment.SetTorque(torque)
	}
}

func NewMonster(w *World, position cp.Vector) *Monster {
	var group uint = 1

	monster := &Monster{
		Arms:   make([]MonsterArm, 0),
		radius: 30.,
	}

	mass := monster.radius * monster.radius / 25.0
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, monster.radius, cp.Vector{})))
	body.SetPosition(position)

	shape := w.Space.AddShape(cp.NewCircle(body, monster.radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)
	shape.Filter.Group = group

	monster.body = body

	var width float64 = monster.radius / 2
	var height float64 = monster.radius * 1.5

	for range 3 {
		arm := MonsterArm{
			Monster: monster,
			Bodies:  make([]*cp.Body, 0),
			Path:    []cp.Vector{body.Position()},
		}

		var prevBody *cp.Body

		for bodyI := range 14 {
			i := float64(bodyI)
			pos := position.Add(cp.Vector{X: (i + 0.5) * width, Y: 0})
			body := w.Space.AddBody(cp.NewBody(5, cp.MomentForBox(1, width, height)))
			body.SetPosition(pos)

			arm.Bodies = append(arm.Bodies, body)

			shape := w.Space.AddShape(cp.NewBox(body, width, height/(1+i/5), 0))
			shape.SetElasticity(0.5)
			shape.SetFriction(0.7)
			shape.Filter.Group = group
			arm.Shapes = append(arm.Shapes, shape)

			if prevBody != nil {

				pivot := pos.Add(cp.Vector{X: -width / 2, Y: 0})
				constraint := w.Space.AddConstraint(cp.NewPivotJoint(prevBody, body, pivot))
				constraint.SetMaxForce(1000000)
				rotaryLimitAngle := rl.Pi / 4
				rotaryLimit := w.Space.AddConstraint(cp.NewRotaryLimitJoint(prevBody, body, -rotaryLimitAngle, rotaryLimitAngle))
				rotaryLimit.SetMaxForce(1000000)

			} else {
				pivot := pos.Add(cp.Vector{X: -width / 2, Y: 0})
				constraint := w.Space.AddConstraint(cp.NewPivotJoint(monster.body, body, pivot))
				constraint.SetMaxForce(1000000)

			}

			prevBody = body

		}

		monster.Arms = append(monster.Arms, arm)

	}
	return monster
}

func (m *Monster) Render(w *World) {
	rl.DrawCircle(int32(m.body.Position().X), int32(m.body.Position().Y), float32(m.radius), rl.Black)

	// for _, arm := range m.Arms {
	// DrawChain(arm.Bodies, arm.Shapes, rl.Black)
	// }
}
