package game2

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	Y          float64
	YVelocity  float64
	Radius     float64
	body       *cp.Body
	shape      *cp.Shape
	PathFinder *PathFinder

	arms []*MonsterArm

	SavePosition Vec2
}

func (p *Monster) Update(g *Game) {
	monsterPos := p.Position3D()

	p.PathFinder.SetPosition(p.Position3D())
	p.PathFinder.SetTarget(g.Player.Position3D())

	force := cp.Vector{}

	if !p.PathFinder.Idle && len(p.PathFinder.Path) >= 2 {
		a := p.PathFinder.Path[0]
		b := p.PathFinder.Path[1]

		force3D := b.Lerp(a, 0.5).Subtract(monsterPos)
		force.X = force3D.X
		force.Y = force3D.Z
	}

	forceMag := force.Length()

	if forceMag != 0 {
		force = force.Normalize().Mult(p.body.Mass() * 60)
	}

	p.body.SetForce(force)

	newVelocity := Vec2FromCP(p.body.Velocity()).Scale(math.Pow(0.01, g.TimeDelta.Seconds()*4))
	p.body.SetVelocity(newVelocity.X, newVelocity.Y)

	p.Y, p.YVelocity = UpdatePhysicsY(g, p.shape, p.Y, p.YVelocity)

	for _, arm := range p.arms {

		tip := arm.segments[len(arm.segments)-1]

		totalCurlAngle := 0.0
		curlAngles := make([]float64, len(arm.segments)-2)

		for i, segment := range arm.segments[:len(arm.segments)-2] {
			a := segment.body.Position()
			b := arm.segments[i+1].body.Position()
			c := arm.segments[i+2].body.Position()
			v1 := b.Sub(a)
			v2 := c.Sub(b)
			cross := v1.Cross(v2)
			dot := v1.Dot(v2)

			angle := math.Atan2(cross, dot)
			totalCurlAngle += angle
			curlAngles[i] = angle
		}
		for i, angle := range curlAngles {
			segment := arm.segments[i]
			segment.body.SetTorque(angle * tip.body.Moment() * 500)
		}

		if !p.PathFinder.Idle && len(p.PathFinder.Path) >= 2 && p.PathFinder.PathLength > 3 {
			closestI := 0
			closestDistance := math.Inf(1)

			for i, point := range p.PathFinder.Path[:len(p.PathFinder.Path)-1] {
				distance := point.Distance(tip.Position3D())
				if distance < closestDistance {
					closestDistance = distance
					closestI = i
				}
			}

			if closestDistance > 3 {
				arm.tipTarget = p.PathFinder.Path[closestI]
			} else {
				arm.tipTarget = p.PathFinder.Path[closestI].Lerp(p.PathFinder.Path[closestI+1], 1)
			}

		} else {
			arm.tipTarget = g.Player.Position3D()
		}

		delta := arm.tipTarget.Subtract(tip.Position3D())
		currentDir := cp.ForAngle(tip.body.Angle())
		relativeAngle := math.Atan2(currentDir.Cross(delta.Chipmunk()), currentDir.Dot(delta.Chipmunk()))

		tip.body.SetTorque(relativeAngle * tip.body.Moment() * 70)
		tip.body.SetForce(delta.Normalize().Scale(50 * tip.body.Mass()).Chipmunk())

		for _, segment := range arm.segments {
			segment.Y, segment.YVelocity = UpdatePhysicsY(g, segment.shape, segment.Y, segment.YVelocity)
		}
	}
}

type MonsterArm struct {
	segments  []*MonsterArmSegment
	tipTarget Vec3
}

func (ma *MonsterArm) Tip() *MonsterArmSegment {
	return ma.segments[len(ma.segments)-1]
}
func (ma *MonsterArm) Base() *MonsterArmSegment {
	return ma.segments[0]
}

type MonsterArmSegment struct {
	body  *cp.Body
	shape *cp.Shape

	Length float64
	Width  float64

	Y         float64
	YVelocity float64
}

func (p *Monster) Position3D() Vec3 {
	return Vec3From2D(Vec2FromCP(p.body.Position()), p.Y)
}
func (p *MonsterArmSegment) Position3D() Vec3 {
	return Vec3From2D(Vec2FromCP(p.body.Position()), p.Y)
}

func (p *Monster) ToSave(g *Game) *Monster {
	if p == nil {
		return nil
	}
	p.SavePosition = Vec2FromCP(p.body.Position())
	return p
}

func (p *Monster) Load(g *Game) *Monster {
	p.Radius = 0.3
	if p.PathFinder == nil {
		p.PathFinder = NewPathFinder(g.Level)
	}
	p.PathFinder.level = g.Level

	mass := p.Radius * p.Radius
	body := g.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, p.Radius, Vec2{2, 2}.CP())))
	position := p.SavePosition
	body.SetPosition(position.CP())

	p.shape = g.Space.AddShape(cp.NewCircle(body, p.Radius, Vec2{}.CP()))
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.body = body
	g.Monster = p
	p.shape.Filter.Group = GroupMonster

	p.arms = make([]*MonsterArm, 0)

	for range 5 {

		arm := &MonsterArm{
			// Index:     i,
			// GaitAngle: float64(i) / float64(ARMS),
			// Monster:   monster,
			segments: make([]*MonsterArmSegment, 0),
		}
		p.arms = append(p.arms, arm)

		prevBody := p.body
		prevPosition := position

		for i := range 12 {

			segment := &MonsterArmSegment{
				Length: p.Radius * 0.6,
				Width:  (p.Radius * 1.5) / (1 + float64(i)/5),
				Y:      p.Y,
			}
			arm.segments = append(arm.segments, segment)

			mass := segment.Length * segment.Width * 1.5

			segment.body = g.Space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, segment.Length, segment.Width)))
			position := prevPosition.AddXY(segment.Length, 0)
			segment.body.SetPosition(position.SubtractXY(segment.Length*0.5, 0).CP())

			segment.shape = g.Space.AddShape(cp.NewBox(segment.body, segment.Length, segment.Width, 0))
			segment.shape.SetElasticity(0.)
			segment.shape.SetFriction(0.1)
			segment.shape.Filter.Group = GroupMonster

			constraint := g.Space.AddConstraint(cp.NewPivotJoint(prevBody, segment.body, prevPosition.CP()))
			constraint.SetMaxForce(1e12)

			if i != 0 {
				rotaryLimitAngle := rl.Pi / 3
				rotaryLimit := g.Space.AddConstraint(cp.NewRotaryLimitJoint(prevBody, segment.body, -rotaryLimitAngle, rotaryLimitAngle))
				rotaryLimit.SetMaxForce(1e12)
				stiffness := 5.0 * segment.body.Moment()
				damping := 2 * math.Sqrt(stiffness*segment.body.Moment())
				g.Space.AddConstraint(cp.NewDampedRotarySpring(prevBody, segment.body, 0, stiffness, damping))
			}

			prevPosition = position
			prevBody = segment.body
		}

		for i, segment := range arm.segments {
			f := float64(i)
			angle := f / 2
			pos := position.Add(NewVec2(math.Cos(f+math.Pi/2), math.Sin(f+math.Pi/2)).Scale(0.25))
			segment.body.SetAngle(-angle)
			segment.body.SetPosition(pos.CP())
		}
	}

	return p
}

func (p *Monster) Draw3D(g *Game, maxY int) {
	rl.BeginBlendMode(rl.BlendAlpha)
	defer rl.EndBlendMode()

	g.MainShader.HideOutsideView.Set(1)

	aa, bb := g.Tileset.GetAABB(0, 4)
	g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
	col := color.RGBA{30, 30, 30, 255}

	if math.Floor(p.Y) <= float64(maxY) {
		rl.DrawModelEx(g.GetModel("monster_body"), p.Position3D().Add(Y.Scale(p.Radius)).Raylib(), Y.Negate().Raylib(), float32(p.body.Angle()*rl.Rad2deg), XYZ.Scale(p.Radius).Raylib(), col)
	}

	for _, arm := range p.arms {

		for _, segment := range arm.segments {

			if math.Floor(segment.Y) <= float64(maxY) {

				position := Vec3From2D(Vec2FromCP(segment.body.Position()), segment.Y).AddXYZ(0, segment.Width/2, 0)

				scale := NewVec3(segment.Length, segment.Width, segment.Width)

				rl.DrawModelEx(g.GetModel("monster_arm_segment"), position.Raylib(), Y.Negate().Raylib(), float32(segment.body.Angle()*rl.Rad2deg), scale.Raylib(), col)
			}
		}
	}

	g.MainShader.HideOutsideView.Set(0)

}
