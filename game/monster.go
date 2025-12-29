package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	body   *cp.Body
	Arms   []MonsterArm
	Radius float64
}

type MonsterArm struct {
	Monster           *Monster
	Bodies            []*cp.Body
	Shapes            []*cp.Shape
	RotaryLimitJoints []*cp.Constraint
	Path              []cp.Vector
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

		arm.Path = append(arm.Path[:max(len(arm.Path)-1, 0)], playerPos)
	}

}

func (m *Monster) Update(w *World) {
	newVelocity := m.body.Velocity().Mult(0.90)
	m.body.SetVelocity(newVelocity.X, newVelocity.Y)

	for i := range m.Arms {
		arm := &m.Arms[i]
		arm.Update(w)

		// targetAngleOffset := float32(i) / float32(len(m.Arms)-1) / 2 * 0.1
		// var tangle float64 = 0
		// prevBody := m.body
		// diffs := make([]cp.Vector, len(arm.Bodies))
		// for i, body := range arm.Bodies {
		// 	diffs[i] = body.Position().Sub(prevBody.Position()).Normalize()
		// 	prevBody = body
		// }
		// for i, diff := range diffs {
		// 	if i == 0 {
		// 		continue
		// 	}
		// 	prevDiff := diffs[i-1]
		// 	a := diff.Rotate(prevDiff)
		// 	angle := math.Atan2(a.Y, a.X)
		// 	tangle += angle
		// }
		// targetPoint := arm.Path[0]

		var tangle float64 = 0
		prevBody := m.body
		diffs := make([]cp.Vector, len(arm.Bodies))
		angles := make([]float64, len(arm.Bodies)-1)

		for i, body := range arm.Bodies {
			diffs[i] = body.Position().Sub(prevBody.Position()).Normalize()
			prevBody = body
		}

		for i, diff := range diffs {
			if i == 0 {
				continue
			}
			prevDiff := diffs[i-1]
			a := diff.Rotate(prevDiff)
			angle := math.Atan2(a.Y, a.X)
			angles[i-1] = angle
			tangle += angle
		}

		// targetPoint := arm.Path[0]
		// curlRatio := math.Abs(tangle) / (rl.Pi * 2)

		// if curlRatio > 0.2 {
		// 	// Uncurl by counter-rotating each segment
		// 	uncurlStrength := 1. // Adjust this to control how aggressive the uncurling is

		// 	for i, body := range arm.Bodies {
		// 		if i == 0 {
		// 			continue // Skip the base
		// 		}

		// 		// Apply torque to counter the curl
		// 		// The sign of tangle tells us which direction we're curled
		// 		curlDirection := tangle / math.Abs(tangle)

		// 		// Apply stronger counter-torque to more curved segments
		// 		segmentCurl := angles[i-1]
		// 		torque := -curlDirection * math.Abs(segmentCurl) * uncurlStrength * body.Moment()

		// 		body.SetTorque(body.Torque() + torque)
		// 	}
		// } else {
		// 	// Normal behavior - follow path
		// 	// arm.TargetPoint(w, 0, arm.Path[0])
		// 	// arm.TargetPoint(w, 0, targetPoint, 0.1)
		// }

		// if tangle {
		// 	arm.TargetPoint(w, 0, m.body.Position().Add(arm.Path[0].Sub(m.body.Position())))

		// } else {

		// }

		arm.TargetPoint(w, 0, arm.Path[0], 1)

	}
}

func (arm *MonsterArm) TargetPoint(w *World, offsetAngle float32, point cp.Vector, power float64) {
	lastSegment := arm.Bodies[len(arm.Bodies)-1]
	lastPos := lastSegment.Position()

	direction := point.Sub(lastPos)
	distance := direction.Length()

	if distance > 0 {
		direction = direction.Normalize()
		desiredAngle := math.Atan2(direction.Y, direction.X) + float64(offsetAngle)
		direction = cp.Vector{X: math.Cos(desiredAngle), Y: math.Sin(desiredAngle)}

		force := direction.Mult(5000 * power)
		lastSegment.SetForce(force)

		currentAngle := lastSegment.Angle()

		angleDiff := desiredAngle - currentAngle
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		torque := angleDiff * 1000
		lastSegment.SetTorque(torque)
	}
}

func NewMonster(w *World, position cp.Vector) *Monster {
	var group uint = 1

	monster := &Monster{
		Arms:   make([]MonsterArm, 0),
		Radius: 2.,
	}

	mass := monster.Radius * monster.Radius * 10
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, monster.Radius, cp.Vector{})))
	body.SetPosition(position)

	shape := w.Space.AddShape(cp.NewCircle(body, monster.Radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)
	shape.Filter.Group = group

	monster.body = body

	var width float64 = monster.Radius / 2
	var height float64 = monster.Radius * 1.5

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
			mass := monster.Radius * monster.Radius
			body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
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
				arm.RotaryLimitJoints = append(arm.RotaryLimitJoints, rotaryLimit)

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
	// rl.DrawCircle(int32(m.body.Position().X), int32(m.body.Position().Y), float32(m.Radius), rl.Black)

	pos := m.body.Position()

	rl.DrawSphereEx(rl.Vector3{X: float32(pos.X), Y: 1, Z: float32(pos.Y)}, float32(m.Radius), 8, 8, rl.Black)

	for _, arm := range m.Arms {
		for i, segment := range arm.Bodies {
			pos := segment.Position()
			rl.DrawSphereEx(rl.Vector3{X: float32(pos.X), Y: 1, Z: float32(pos.Y)}, float32(arm.Monster.Radius/(1.2+float64(i)/10)), 8, 8, rl.Black)
		}
	}
}
