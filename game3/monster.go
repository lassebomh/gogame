package game3

import (
	v2 "game/vec2"
	v3 "game/vec3"
	"image/color"
	"log"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	DynamicPhysicsObject

	Radius float64

	arms  []*MonsterArm
	Aggro float64
	path  []*PathPoint
}

type MonsterArm struct {
	segments  []*MonsterArmSegment
	tipTarget v3.Value
	path      []*PathPoint
}

type MonsterArmSegment struct {
	DynamicPhysicsObject

	Length float64
	Width  float64
}

func (m *Monster) Update() {

	newVelocity := v2.FromChipmunk(m.body.Velocity()).Scale(math.Pow(0.01, m.world.TimeStep.Seconds()*6))
	m.body.SetVelocity(newVelocity.X, newVelocity.Y)

	m.UpdatePhysics()

	tipsVisibleToPlayer := 0
	tipsMinDistanceToPlayer := math.Inf(1)

	if m.world.Player != nil {
		for _, arm := range m.arms {
			tip := arm.segments[len(arm.segments)-1]
			if math.Floor(tip.Position.Y) != math.Floor(m.world.Player.Position.Y) {
				continue
			}

			result := m.world.space.SegmentQueryFirst(
				tip.Position.Chipmunk(),
				m.world.Player.Position.Chipmunk(),
				0,
				cp.NewShapeFilter(
					0,
					Category(tip.Position.Y, true, false),
					Category(tip.Position.Y, true, false),
				),
			)

			if result.Shape == nil {
				tipsVisibleToPlayer++
			}

			dist := tip.Position.Distance(m.world.Player.Position) - m.world.Player.Radius
			if dist < tipsMinDistanceToPlayer {
				tipsMinDistanceToPlayer = dist
			}
		}
	}

	if tipsVisibleToPlayer > 0 && tipsMinDistanceToPlayer < 2 {
		m.Aggro = 3

		// pathPoint := NewPathPoint(m.world, m.Position)
		// m.path, _ = pathPoint.FindPath(m.world.Player.Position)
		// if len(m.path) >= 3 {
		// 	speed := 200.
		// 	delta := m.Position.Subtract(m.path[2].Center)
		// 	m.body.SetForce(delta.Normalize().Scale(speed * m.body.Mass()).Chipmunk())
		// 	// newVelocity := v2.FromChipmunk(m.body.Velocity()).Scale(math.Pow(0.01, m.world.TimeStep.Seconds()))
		// 	// m.body.SetVelocity(newVelocity.X, newVelocity.Y)
		// }

	} else if m.Aggro > 0 {
		m.Aggro -= m.world.TimeStep.Seconds()
	}

	for i, arm := range m.arms {
		tip := arm.segments[len(arm.segments)-1]

		curlAngles := make([]float64, len(arm.segments)-2)
		for ii, segment := range arm.segments[:len(arm.segments)-2] {
			a := segment.body.Position()
			b := arm.segments[ii+1].Position.Chipmunk()
			c := arm.segments[ii+2].Position.Chipmunk()
			v1 := b.Sub(a)
			v2 := c.Sub(b)

			angle := math.Atan2(v1.Cross(v2), v1.Dot(v2))
			arm.segments[ii].body.SetTorque(angle * tip.body.Moment() * 200)
			curlAngles[ii] = angle
		}

		if m.world.Player != nil {

			pathPoint := NewPathPoint(m.world, tip.Position)
			arm.path, _ = pathPoint.FindPath(m.world.Player.Position)

			speed := 30.

			if m.Aggro > 0 {
				speed = 120
			}

			if len(arm.path) >= 3 {
				arm.tipTarget = arm.path[1].Center.Lerp(arm.path[2].Center, 0.5)

			} else {

				entangleTarget := m.world.Player.Position

				playerDist := entangleTarget.Distance(tip.Position)

				dir := entangleTarget.Subtract(tip.Position).Normalize().RotateByAxisAngle(v3.Y(-1), math.Pi/2)
				if i%2 == 0 {
					dir = dir.Negate()
				}

				entangleTarget = entangleTarget.Add(dir.Scale(math.Sqrt(playerDist*m.world.Player.Radius) + m.world.Player.Radius))

				// arm.tipTarget = m.world.Player.Position
				arm.tipTarget = entangleTarget
			}
			delta := arm.tipTarget.Subtract(tip.Position)
			currentDir := cp.ForAngle(tip.body.Angle())
			relativeAngle := math.Atan2(currentDir.Cross(delta.Chipmunk()), currentDir.Dot(delta.Chipmunk()))
			tip.body.SetTorque(relativeAngle * tip.body.Moment() * 70)
			tip.body.SetForce(delta.Normalize().Scale(speed * tip.body.Mass()).Chipmunk())
			arm.segments[len(arm.segments)-1].body.SetForce(delta.Normalize().Scale(speed * tip.body.Mass()).Chipmunk())
			arm.segments[len(arm.segments)-2].body.SetForce(delta.Normalize().Scale(speed * tip.body.Mass()).Chipmunk())

		}

		for _, segment := range arm.segments {
			segment.UpdatePhysics()
		}
	}
}

func (m *Monster) Draw() {

	minTipDistance := math.Inf(1)

	if m.world.Player != nil {
		for _, arm := range m.arms {
			tip := arm.segments[len(arm.segments)-1]
			dist := m.world.Player.Position.Distance(tip.Position) - m.world.Player.Radius
			if dist < minTipDistance {
				minTipDistance = dist
			}
		}
	}

	minTipDistance /= 1.5

	globals.Shaders.Main.HideOutsideView.Set(1)
	globals.Shaders.Main.Visibility.Set(v2.Clamp(1-minTipDistance, 0, 1))
	defer globals.Shaders.Main.HideOutsideView.Set(0)

	col := color.RGBA{50, 50, 50, 255}

	rl.DrawModelEx(globals.Models.MonsterBody, m.Position.Add(v3.Y(m.Radius)).Raylib(), v3.Y(-1).Raylib(), float32(m.body.Angle()*rl.Rad2deg), v3.Fill(m.Radius).Raylib(), col)

	for _, arm := range m.arms {

		positions := make([]v3.Value, len(arm.segments)+1)
		positions[0] = m.Position.Add(v3.Y(m.Radius))

		for i, segment := range arm.segments {

			segmentOffset := v3.XYZ(
				math.Cos(segment.body.Angle()),
				0,
				math.Sin(segment.body.Angle()),
			).Scale(segment.Length / 2)

			heightOffset := v3.XYZ(0, segment.Width/2+0.25, 0)

			positions[i+1] = segment.Position.Add(segmentOffset).Add(heightOffset)
		}

		positionsSoft := append([]v3.Value{}, positions...)

		for i := 1; i < len(positions)-1; i++ {
			positionsSoft[i] = positions[i].Lerp(positions[i-1].Lerp(positions[i+1], 0.5), 0.5)
		}

		for i, segment := range arm.segments {

			from := positionsSoft[i]
			to := positionsSoft[i+1]

			middle := from.Lerp(to, 0.5)

			direction := to.Subtract(from).Normalize()

			yaw := math.Atan2(float64(-direction.X), float64(direction.Z))
			pitch := math.Asin(float64(direction.Y))

			mat := (v3.NewMatrix().RotateY(
				-math.Pi/2,
			).Scale(
				segment.Width*1.2, segment.Width*1.2, from.Distance(to)*0.9,
			).RotateZ(
				segment.body.AngularVelocity() / 20,
			).RotateX(
				pitch,
			).RotateY(
				yaw,
			).Translate(
				middle,
			))

			globals.Models.MonsterArmSegment.Transform = mat.Raylib()
			rl.DrawModel(globals.Models.MonsterArmSegment, rl.Vector3{}, 1, col)

		}
	}
}

func SpawnNewMonster(world *World) {
	m := &Monster{
		Radius: 0.25,
		DynamicPhysicsObject: DynamicPhysicsObject{
			Position: v3.XYZ(0, 0, 0),
		},
	}
	m.Spawn(world)
}

func (m *Monster) Spawn(world *World) *Monster {
	if m.world != nil {
		log.Fatal("not implemented")
		m.world.Monster = nil
	}
	m.world = world
	m.world.Monster = m
	m.Position = v3.XYZ(0, 0, -5)

	m.Radius = 0.22

	mass := m.Radius * m.Radius / 1.5
	body := m.world.space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, m.Radius, cp.Vector{2, 2})))

	body.SetPosition(m.Position.Chipmunk())

	m.shape = m.world.space.AddShape(cp.NewCircle(body, m.Radius, cp.Vector{}))
	m.shape.SetElasticity(0)
	m.shape.SetFriction(0)
	m.body = body
	m.world.Monster = m
	m.shape.Filter.Group = GroupMonster
	m.UpdatePhysics()

	m.arms = make([]*MonsterArm, 0)

	for range 5 {

		arm := &MonsterArm{
			segments: make([]*MonsterArmSegment, 0),
		}
		m.arms = append(m.arms, arm)

		prevBody := m.body
		prevPosition := m.Position

		for i := range 16 {

			segment := &MonsterArmSegment{
				// Length: m.Radius * 1,
				DynamicPhysicsObject: DynamicPhysicsObject{
					world: world,
				},
				Length: m.Radius * 0.7,
				Width:  (m.Radius * 2) / (1 + float64(i)/6),
			}
			arm.segments = append(arm.segments, segment)

			mass := segment.Length * segment.Width * 0.5

			segment.body = m.world.space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, segment.Length, segment.Width)))
			position := prevPosition.Add(v3.X(segment.Length))
			segment.body.SetPosition(position.Subtract(v3.X(segment.Length * 0.5)).Chipmunk())

			segment.shape = m.world.space.AddShape(cp.NewBox(segment.body, segment.Length, segment.Width, 0))
			segment.shape.SetElasticity(0.5)
			segment.shape.SetFriction(0.5)
			segment.shape.Filter.Group = GroupMonster

			constraint := m.world.space.AddConstraint(cp.NewPivotJoint(prevBody, segment.body, prevPosition.Chipmunk()))
			constraint.SetMaxForce(math.Inf(1))

			if i != 0 {
				rotaryLimitAngle := rl.Pi / 2.5
				rotaryLimit := m.world.space.AddConstraint(cp.NewRotaryLimitJoint(prevBody, segment.body, -rotaryLimitAngle, rotaryLimitAngle))
				rotaryLimit.SetMaxForce(math.Inf(1))
				stiffness := 5.0 * segment.body.Moment()
				damping := 2 * math.Sqrt(stiffness*segment.body.Moment())
				m.world.space.AddConstraint(cp.NewDampedRotarySpring(prevBody, segment.body, 0, stiffness, damping))
			}

			segment.UpdatePhysics()

			prevPosition = position
			prevBody = segment.body
		}

		for i, segment := range arm.segments {
			f := float64(i)
			angle := f / 2
			pos := m.Position.Add(v3.XZ(math.Cos(f+math.Pi/2), math.Sin(f+math.Pi/2)).Scale(0.25))
			segment.body.SetAngle(-angle)
			segment.body.SetPosition(pos.Chipmunk())
		}
	}

	return m
}
