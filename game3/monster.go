package game3

import (
	"fmt"
	"game/vec2"
	"game/vec3"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	YVelocity float64
	Radius    float64
	body      *cp.Body
	shape     *cp.Shape
	// PathFinder *PathFinder

	arms []*MonsterArm

	world    *World
	Position vec3.Value

	bodyModel    rl.Model
	segmentModel rl.Model
}

type MonsterArm struct {
	segments  []*MonsterArmSegment
	tipTarget vec3.Value
	path      []vec3.Value
}

type MonsterArmSegment struct {
	body  *cp.Body
	shape *cp.Shape

	Length float64
	Width  float64

	YVelocity float64

	Position vec3.Value
}

func (m *Monster) Update() {

	newVelocity := vec2.FromChipmunk(m.body.Velocity()).Scale(math.Pow(0.01, m.world.TimeStep.Seconds()*4))
	m.body.SetVelocity(newVelocity.X, newVelocity.Y)

	m.Position, m.YVelocity = UpdatePhysicsY(m.world, m.shape, m.Position, m.YVelocity)

	for i, arm := range m.arms {
		tip := arm.segments[len(arm.segments)-1]

		curlAngles := make([]float64, len(arm.segments)-2)
		for i, segment := range arm.segments[:len(arm.segments)-2] {
			a := segment.body.Position()
			b := arm.segments[i+1].Position.Chipmunk()
			c := arm.segments[i+2].Position.Chipmunk()
			v1 := b.Sub(a)
			v2 := c.Sub(b)

			angle := math.Atan2(v1.Cross(v2), v1.Dot(v2))
			arm.segments[i].body.SetTorque(angle * tip.body.Moment() * 200)
			curlAngles[i] = angle
		}

		if m.world.Player != nil {

			var length float64
			arm.path, length = FindPath(m.world, tip.Position, m.world.Player.Position)
			speed := 70.
			
			if length < 4 {
				speed = 180
			}

			if len(arm.path) >= 3 {
				arm.tipTarget = arm.path[1].Lerp(arm.path[2], 0.5)

			} else {
				
				entangleTarget := m.world.Player.Position
				
				playerDist := entangleTarget.Distance(tip.Position)
				
				dir := entangleTarget.Subtract(tip.Position).Normalize().RotateByAxisAngle(vec3.Y(-1), math.Pi/2)
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

		}

		for _, segment := range arm.segments {
			segment.Position, segment.YVelocity = UpdatePhysicsY(m.world, segment.shape, segment.Position, segment.YVelocity)
		}
	}
}

func (m *Monster) Draw() {
	m.world.shader.Visibility.Set(1)
	// m.world.shader.HideOutsideView.Set(1)
	col := color.RGBA{50, 50, 50, 255}

	rl.DrawModelEx(m.bodyModel, m.Position.Add(vec3.Y(m.Radius)).Raylib(), vec3.Y(-1).Raylib(), float32(m.body.Angle()*rl.Rad2deg), vec3.Fill(m.Radius).Raylib(), col)

	for _, arm := range m.arms {

		positions := make([]vec3.Value, len(arm.segments)+1)
		positions[0] = m.Position.Add(vec3.Y(m.Radius))

		for i, segment := range arm.segments {

			segmentOffset := vec3.XYZ(
				math.Cos(segment.body.Angle()),
				0,
				math.Sin(segment.body.Angle()),
			).Scale(segment.Length / 2)

			heightOffset := vec3.XYZ(0, segment.Width/2+0.15, 0)

			positions[i+1] = segment.Position.Add(segmentOffset).Add(heightOffset)
		}

		positionsSoft := append([]vec3.Value{}, positions...)

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

			mat := (vec3.NewMatrix().RotateY(
				-math.Pi/2,
			).Scale(
				segment.Width, segment.Width, from.Distance(to)*0.9,
			).RotateZ(
				segment.body.AngularVelocity() / 20,
			).RotateX(
				pitch,
			).RotateY(
				yaw,
			).Translate(
				middle,
			))

			m.segmentModel.Transform = mat.Raylib()
			rl.DrawModel(m.segmentModel, rl.Vector3{}, 1, col)

		}
	}

	// m.world.shader.HideOutsideView.Set(0)
}

func SpawnNewMonster(world *World) {
	m := &Monster{
		Radius:   0.25,
		Position: vec3.XYZ(0, 0, 0),
	}
	m.Spawn(world)
}

func (m *Monster) Spawn(world *World) *Monster {
	if m.world != nil {
		m.world.Monster = nil
	}
	m.world = world
	m.world.Monster = m
	m.Position = vec3.XYZ(0, 0, -5)

	if !rl.IsModelValid(m.bodyModel) {
		m.bodyModel = rl.LoadModel("./models/monster/monster_body.glb")
		mats := m.bodyModel.GetMaterials()
		for i := range mats {
			mats[i].Shader = m.world.shader.shader
			mats[i].GetMap(rl.MapDiffuse).Texture = world.planetOrganicTexture
		}
	}
	if !rl.IsModelValid(m.segmentModel) {
		m.segmentModel = rl.LoadModel("./models/monster/monster_arm_segment.glb")
		mats := m.segmentModel.GetMaterials()
		for i := range mats {
			mats[i].Shader = m.world.shader.shader
			mats[i].GetMap(rl.MapDiffuse).Texture = world.planetOrganicTexture
		}
	}

	m.Radius = 0.3
	// if p.PathFinder == nil {
	// 	p.PathFinder = NewPathFinder(g.Level)
	// }
	// p.PathFinder.level = g.Level

	mass := m.Radius * m.Radius
	body := m.world.space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, m.Radius, cp.Vector{2, 2})))

	body.SetPosition(m.Position.Chipmunk())

	m.shape = m.world.space.AddShape(cp.NewCircle(body, m.Radius, cp.Vector{}))
	m.shape.SetElasticity(0)
	m.shape.SetFriction(0)
	m.body = body
	m.world.Monster = m
	m.shape.Filter.Group = GroupMonster

	m.arms = make([]*MonsterArm, 0)

	for range 3 {

		arm := &MonsterArm{
			segments: make([]*MonsterArmSegment, 0),
		}
		m.arms = append(m.arms, arm)

		prevBody := m.body
		prevPosition := m.Position

		for i := range 12 {

			segment := &MonsterArmSegment{
				// Length: m.Radius * 1,
				Length: m.Radius * 0.7,
				Width:  (m.Radius * 1.5) / (1 + float64(i)/5),
			}
			arm.segments = append(arm.segments, segment)

			mass := segment.Length * segment.Width * 0.5

			segment.body = m.world.space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, segment.Length, segment.Width)))
			position := prevPosition.Add(vec3.X(segment.Length))
			segment.body.SetPosition(position.Subtract(vec3.X(segment.Length * 0.5)).Chipmunk())

			segment.shape = m.world.space.AddShape(cp.NewBox(segment.body, segment.Length, segment.Width, 0))
			segment.shape.SetElasticity(0.5)
			segment.shape.SetFriction(0.5)
			segment.shape.Filter.Group = GroupMonster

			fmt.Printf("%+v\n", prevBody.UserData)

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

			prevPosition = position
			prevBody = segment.body
		}

		// for i, segment := range arm.segments {
		// 	f := float64(i)
		// 	angle := f / 2
		// 	pos := m.Position.Add(vec3.XZ(math.Cos(f+math.Pi/2), math.Sin(f+math.Pi/2)).Scale(0.25))
		// 	segment.body.SetAngle(-angle)
		// 	segment.body.SetPosition(pos.Chipmunk())
		// }
	}

	return m
}
