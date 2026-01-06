package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	Body         *cp.Body
	Arms         []*MonsterArm
	Radius       float64
	Path         []cp.Vector
	PathDistance float64
	Target       cp.Vector
}

type MonsterArm struct {
	Index     int
	Monster   *Monster
	GaitAngle float64
	Segments  []*MonsterArmSegment
	TipTarget cp.Vector
}

type MonsterArmSegment struct {
	Body             *cp.Body
	Shape            *cp.Shape
	RotaryLimitJoint *cp.Constraint

	Length float64
	Width  float64
}

const ARMS = 4
const ARM_SEGMENTS = 12

func NewMonster(w *World, position cp.Vector) *Monster {
	var group uint = 1

	monster := &Monster{
		Arms:   make([]*MonsterArm, 0),
		Radius: 2.2,
		Path:   []cp.Vector{},
	}

	mass := monster.Radius * monster.Radius * 30
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, monster.Radius, cp.Vector{})))
	body.SetPosition(position)

	shape := w.Space.AddShape(cp.NewCircle(body, monster.Radius, cp.Vector{}))
	shape.SetElasticity(0.5)
	shape.SetFriction(1)
	shape.Filter.Group = group

	monster.Body = body

	for i := range ARMS {

		arm := &MonsterArm{
			Index:     i,
			GaitAngle: float64(i) / float64(ARMS),
			Monster:   monster,
			Segments:  make([]*MonsterArmSegment, 0),
		}
		monster.Arms = append(monster.Arms, arm)

		var prevBody *cp.Body

		for segmentI := range ARM_SEGMENTS {
			i := float64(segmentI)

			segment := &MonsterArmSegment{
				Length: monster.Radius * 0.8,
				Width:  (monster.Radius * 1.5) / (0.8 + i/5),
			}
			arm.Segments = append(arm.Segments, segment)

			mass := segment.Length * segment.Width

			pos := position.Add(cp.Vector{X: (i + 0.5) * segment.Length, Y: 0})
			segment.Body = w.Space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, segment.Length, segment.Width)))
			segment.Body.SetPosition(pos)

			segment.Shape = w.Space.AddShape(cp.NewBox(segment.Body, segment.Length, segment.Width, 0))
			segment.Shape.SetElasticity(0.5)
			segment.Shape.SetFriction(1)
			segment.Shape.Filter.Group = group

			if prevBody != nil {
				pivot := pos.Add(cp.Vector{X: -segment.Length / 2, Y: 0})
				constraint := w.Space.AddConstraint(cp.NewPivotJoint(prevBody, segment.Body, pivot))
				constraint.SetMaxForce(1000000)

				rotaryLimitAngle := rl.Pi / 3
				rotaryLimit := w.Space.AddConstraint(cp.NewRotaryLimitJoint(prevBody, segment.Body, -rotaryLimitAngle, rotaryLimitAngle))
				rotaryLimit.SetMaxForce(100000000)
				segment.RotaryLimitJoint = rotaryLimit

				stiffness := 40.0 * segment.Body.Moment()
				damping := 2 * math.Sqrt(stiffness*segment.Body.Moment())
				w.Space.AddConstraint(cp.NewDampedRotarySpring(prevBody, segment.Body, 0, stiffness, damping))

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

func (m *Monster) Update(w *World) {
	m.Body.SetVelocityVector(m.Body.Velocity().Mult(0.90))
	m.Body.SetAngularVelocity(m.Body.AngularVelocity() * 0.95)

	playerPos := w.Player.Body.Position()

	var closestPathPointDistance float64

	if len(m.Path) > 0 {
		closestPathPointDistance = m.Body.Position().Distance(m.Path[0])
		closestPathPointI := 0

		for i, pathPoint := range m.Path[:min(3, len(m.Path))] {
			pathPointDistance := pathPoint.Distance(m.Body.Position())

			if pathPointDistance < closestPathPointDistance {
				closestPathPointDistance = pathPointDistance
				closestPathPointI = i
			}
		}

		if closestPathPointDistance < w.Tilemap.Scale/2 {
			m.Path = m.Path[closestPathPointI+1:]
		}
	}

	if len(m.Path) > 0 {
		m.Target = m.Path[0]

		m.Body.SetForce(m.Target.Sub(m.Body.Position()).Normalize().Mult(175 * m.Body.Mass()))
	}

	if len(m.Path) == 0 || closestPathPointDistance > w.Tilemap.Scale || m.Path[len(m.Path)-1].Distance(playerPos) > w.Tilemap.Scale*0.9 {
		m.PathDistance, m.Path = w.Tilemap.FindPath(m.Body.Position(), playerPos)
		m.Path = append(m.Path[:max(len(m.Path)-1, 0)], playerPos)
	}

	for _, arm := range m.Arms {
		arm.Update(w)
	}
}

func (arm *MonsterArm) Update(w *World) {

	tip := arm.Segments[len(arm.Segments)-1]

	totalCurlAngle := 0.0
	curlAngles := make([]float64, len(arm.Segments)-2)

	for i, segment := range arm.Segments[:len(arm.Segments)-2] {
		a := segment.Body.Position()
		b := arm.Segments[i+1].Body.Position()
		c := arm.Segments[i+2].Body.Position()
		v1 := b.Sub(a)
		v2 := c.Sub(b)
		cross := v1.Cross(v2)
		dot := v1.Dot(v2)

		angle := math.Atan2(cross, dot)
		totalCurlAngle += angle
		curlAngles[i] = angle
	}
	for i, angle := range curlAngles {
		segment := arm.Segments[i]
		segment.Body.SetTorque(angle * tip.Body.Moment() * 1000)
	}

	entangleTarget := w.Player.Body.Position()
	playerDist := entangleTarget.Distance(tip.Body.Position())
	{
		dir := entangleTarget.Sub(tip.Body.Position()).Normalize().Rotate(cp.Vector{Y: 1})
		if arm.Index%2 == 0 {
			dir = dir.Neg()
		}

		entangleTarget = entangleTarget.Add(dir.Mult(math.Sqrt(playerDist*w.Player.Radius) + w.Player.Radius))
	}

	var pathTarget cp.Vector

	closestTipPathPointI := 0
	closestTipPathPointDistance := tip.Body.Position().Distance(arm.Monster.Path[0])

	for i, pathPoint := range arm.Monster.Path[:min(3, len(arm.Monster.Path))] {

		pathPointDistance := pathPoint.Distance(tip.Body.Position())

		if pathPointDistance < closestTipPathPointDistance {
			closestTipPathPointDistance = pathPointDistance
			closestTipPathPointI = i
		}
	}
	pathTarget = arm.Monster.Path[closestTipPathPointI]
	if closestTipPathPointI < len(arm.Monster.Path)-1 {
		pathTarget = arm.Monster.Path[closestTipPathPointI+1]
	}

	if arm.Monster.PathDistance/(w.Player.Radius*2) > 8 {
		arm.TipTarget = pathTarget
	} else {
		arm.TipTarget = entangleTarget
	}

	delta := arm.TipTarget.Sub(tip.Body.Position())
	currentDir := cp.ForAngle(tip.Body.Angle())
	relativeAngle := math.Atan2(currentDir.Cross(delta), currentDir.Dot(delta))

	tip.Body.SetTorque(relativeAngle * tip.Body.Moment() * 70)

	tip.Body.SetForce(delta.Normalize().Mult(1000 * tip.Body.Mass()))

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
