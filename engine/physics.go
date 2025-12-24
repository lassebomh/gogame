package engine

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Body struct {
	Position     rl.Vector2
	PrevPosition rl.Vector2

	Angle     float32
	PrevAngle float32

	Velocity        rl.Vector2
	AngularVelocity float32

	InvMass    float32
	InvInertia float32

	Shape Shape
}

func CreateBody(x float32, y float32, angle float32, invMass float32, shape Shape) *Body {
	return &Body{
		Position:        rl.Vector2{X: x, Y: y},
		PrevPosition:    rl.Vector2{X: x, Y: y},
		Angle:           angle,
		PrevAngle:       angle,
		AngularVelocity: 0,
		Velocity:        rl.Vector2{X: 0, Y: 0},
		InvMass:         invMass,
		InvInertia:      0,
		Shape:           shape,
	}
}

type Contact struct {
	BodyA, BodyB   *Body
	PointA, PointB rl.Vector2
	Normal         rl.Vector2
	Penetration    float32
	Lambda         float32
	Compliance     float32
}

func (c *Contact) Solve(dt float32) {
	a := c.BodyA
	b := c.BodyB

	if a.InvMass == 0 && b.InvMass == 0 {
		return
	}

	C := -c.Penetration

	wA := a.InvMass
	wB := b.InvMass
	w := wA + wB

	if w == 0 {
		return
	}

	alpha := c.Compliance / (dt * dt)
	deltaLambda := (-C - alpha*c.Lambda) / (w + alpha)

	impulse := rl.Vector2Scale(c.Normal, deltaLambda)

	a.Position = rl.Vector2Add(a.Position, rl.Vector2Scale(impulse, wA))
	b.Position = rl.Vector2Subtract(b.Position, rl.Vector2Scale(impulse, wB))

	c.Lambda += deltaLambda
}

type Shape interface{}

type Circle struct {
	Radius float32
}

type Box struct {
	Width  float32
	Height float32
}

type World struct {
	Bodies  []*Body
	Gravity rl.Vector2
}

func (w *World) Integrate(dt float32) {
	for _, body := range w.Bodies {
		if body.InvMass == 0 {
			continue
		}

		body.Velocity.X += w.Gravity.X * dt
		body.Velocity.Y += w.Gravity.Y * dt

		body.PrevPosition = body.Position
		body.PrevAngle = body.Angle

		body.Position.X += body.Velocity.X * dt
		body.Position.Y += body.Velocity.Y * dt
		body.Angle += body.AngularVelocity * dt
	}
}

func (w *World) GenerateContactConstraints() []Contact {
	var contacts []Contact

	for i := 0; i < len(w.Bodies); i++ {
		for j := i + 1; j < len(w.Bodies); j++ {
			if ok, contact := w.Bodies[i].CollidesWith(w.Bodies[j]); ok {
				contacts = append(contacts, contact)
			}
		}
	}

	return contacts
}

func (w *World) SolveConstraints(contacts []Contact, dt float32, iterations int) {
	for iter := 0; iter < iterations; iter++ {
		for i := range contacts {
			contacts[i].Solve(dt)
		}
	}
}

func (w *World) Clone() *World {
	clone := *w
	clone.Bodies = make([]*Body, 0)
	for _, v := range w.Bodies {
		bodyClone := *v
		clone.Bodies = append(clone.Bodies, &bodyClone)
	}

	return &clone
}

func (b *Body) Render(color rl.Color) {
	const dirLen float32 = 20

	dir := rl.Vector2{
		X: float32(math.Cos(float64(b.Angle))),
		Y: float32(math.Sin(float64(b.Angle))),
	}

	dirEnd := rl.Vector2{
		X: b.Position.X + dir.X*dirLen,
		Y: b.Position.Y + dir.Y*dirLen,
	}

	switch c := any(b.Shape).(type) {

	case Circle:

		rl.DrawCircleLines(
			int32(b.Position.X),
			int32(b.Position.Y),
			c.Radius,
			color,
		)

		rl.DrawLineV(b.Position, dirEnd, color)

	case Box:
		hw := c.Width / 2
		hh := c.Height / 2

		corners := [4]rl.Vector2{
			{X: -hw, Y: -hh},
			{X: hw, Y: -hh},
			{X: hw, Y: hh},
			{X: -hw, Y: hh},
		}

		cos := float32(math.Cos(float64(b.Angle)))
		sin := float32(math.Sin(float64(b.Angle)))

		for i := range corners {
			x := corners[i].X
			y := corners[i].Y

			corners[i].X = b.Position.X + x*cos - y*sin
			corners[i].Y = b.Position.Y + x*sin + y*cos
		}

		for i := 0; i < 4; i++ {
			a := corners[i]
			b := corners[(i+1)%4]
			rl.DrawLineV(a, b, color)
		}

		rl.DrawLineV(b.Position, dirEnd, color)
	}
}

func CircleVsCircle(a, b *Body) (bool, Contact) {
	delta := rl.Vector2Subtract(b.Position, a.Position)
	distSq := delta.X*delta.X + delta.Y*delta.Y
	aRadius := a.Shape.(Circle).Radius
	bRadius := b.Shape.(Circle).Radius
	radiusSum := aRadius + bRadius

	if distSq <= radiusSum*radiusSum {
		dist := float32(math.Sqrt(float64(distSq)))
		normal := rl.Vector2{X: 0, Y: 0}

		if dist != 0 {
			normal = rl.Vector2Scale(delta, 1/dist)
		}

		penetration := radiusSum - dist

		contactA := rl.Vector2Add(a.Position, rl.Vector2Scale(normal, aRadius))
		contactB := rl.Vector2Add(b.Position, rl.Vector2Scale(normal, -bRadius))

		return true, Contact{
			BodyA:       a,
			BodyB:       b,
			PointA:      contactA,
			PointB:      contactB,
			Normal:      normal,
			Penetration: penetration,
		}
	}

	return false, Contact{}
}

func CircleVsBox(circle *Body, box *Body) (bool, Contact) {

	delta := rl.Vector2Subtract(circle.Position, box.Position)

	cos := float32(math.Cos(float64(-box.Angle)))
	sin := float32(math.Sin(float64(-box.Angle)))
	localX := delta.X*cos - delta.Y*sin
	localY := delta.X*sin + delta.Y*cos

	circleShape := circle.Shape.(Circle)
	boxShape := box.Shape.(Box)

	halfW := boxShape.Width / 2
	halfH := boxShape.Height / 2
	closestX := clamp(localX, -halfW, halfW)
	closestY := clamp(localY, -halfH, halfH)

	inside := false
	if localX == closestX && localY == closestY {
		inside = true

		distLeft := localX + halfW
		distRight := halfW - localX
		distTop := localY + halfH
		distBottom := halfH - localY

		minDist := min(distLeft, distRight, distTop, distBottom)
		switch minDist {
		case distLeft:
			closestX = -halfW
		case distRight:
			closestX = halfW
		case distTop:
			closestY = -halfH
		default:
			closestY = halfH
		}
	}

	cosA := float32(math.Cos(float64(box.Angle)))
	sinA := float32(math.Sin(float64(box.Angle)))
	worldClosestX := closestX*cosA - closestY*sinA
	worldClosestY := closestX*sinA + closestY*cosA
	closestWorld := rl.Vector2Add(box.Position, rl.Vector2{X: worldClosestX, Y: worldClosestY})

	normal := rl.Vector2Subtract(circle.Position, closestWorld)
	distSq := normal.X*normal.X + normal.Y*normal.Y
	radius := circleShape.Radius

	if distSq > radius*radius && !inside {
		return false, Contact{}
	}

	dist := float32(math.Sqrt(float64(distSq)))

	if inside {
		normal = rl.Vector2Scale(normal, -1)
		dist = -dist
	}

	if dist != 0 {
		normal = rl.Vector2Scale(normal, 1/dist)
	} else {

		normal = rl.Vector2{X: 0, Y: -1}
	}

	penetration := radius - dist

	contactA := rl.Vector2Add(circle.Position, rl.Vector2Scale(normal, -radius))
	contactB := closestWorld

	return true, Contact{
		BodyA:       circle,
		BodyB:       box,
		PointA:      contactA,
		PointB:      contactB,
		Normal:      normal,
		Penetration: penetration,
	}
}

func clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func min(vals ...float32) float32 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func BoxVsBox(a, b *Body) (bool, Contact) {

	aShape := a.Shape.(Box)
	bShape := a.Shape.(Box)

	aHW := aShape.Width / 2
	aHH := aShape.Height / 2
	bHW := bShape.Width / 2
	bHH := bShape.Height / 2

	aCos := float32(math.Cos(float64(a.Angle)))
	aSin := float32(math.Sin(float64(a.Angle)))
	bCos := float32(math.Cos(float64(b.Angle)))
	bSin := float32(math.Sin(float64(b.Angle)))

	aAxisX := rl.Vector2{X: aCos, Y: aSin}
	aAxisY := rl.Vector2{X: -aSin, Y: aCos}

	bAxisX := rl.Vector2{X: bCos, Y: bSin}
	bAxisY := rl.Vector2{X: -bSin, Y: bCos}

	axes := []rl.Vector2{aAxisX, aAxisY, bAxisX, bAxisY}

	minPenetration := float32(math.MaxFloat32)
	var bestAxis rl.Vector2

	for _, axis := range axes {

		aMin, aMax := projectBox(a.Position, aHW, aHH, aCos, aSin, axis)
		bMin, bMax := projectBox(b.Position, bHW, bHH, bCos, bSin, axis)

		if aMax < bMin || bMax < aMin {
			return false, Contact{}
		}

		penetration := min(aMax-bMin, bMax-aMin)

		if penetration < minPenetration {
			minPenetration = penetration
			bestAxis = axis

			delta := rl.Vector2Subtract(b.Position, a.Position)
			dot := delta.X*axis.X + delta.Y*axis.Y
			if dot < 0 {
				bestAxis = rl.Vector2Scale(axis, -1)
			}
		}
	}

	contactA, contactB := findBoxContactPoints(a, b, bestAxis, aHW, aHH, bHW, bHH, aCos, aSin, bCos, bSin)

	return true, Contact{
		BodyA:       a,
		BodyB:       b,
		PointA:      contactA,
		PointB:      contactB,
		Normal:      bestAxis,
		Penetration: minPenetration,
	}
}

func projectBox(center rl.Vector2, halfW, halfH, cos, sin float32, axis rl.Vector2) (float32, float32) {

	corners := [4]rl.Vector2{
		{X: -halfW, Y: -halfH},
		{X: halfW, Y: -halfH},
		{X: halfW, Y: halfH},
		{X: -halfW, Y: halfH},
	}

	minProj := float32(math.MaxFloat32)
	maxProj := float32(-math.MaxFloat32)

	for _, corner := range corners {

		worldX := center.X + corner.X*cos - corner.Y*sin
		worldY := center.Y + corner.X*sin + corner.Y*cos

		proj := worldX*axis.X + worldY*axis.Y

		if proj < minProj {
			minProj = proj
		}
		if proj > maxProj {
			maxProj = proj
		}
	}

	return minProj, maxProj
}

func findBoxContactPoints(a, b *Body, normal rl.Vector2,
	aHW, aHH, bHW, bHH, aCos, aSin, bCos, bSin float32) (rl.Vector2, rl.Vector2) {

	aCornersLocal := [4]rl.Vector2{
		{X: -aHW, Y: -aHH}, {X: aHW, Y: -aHH},
		{X: aHW, Y: aHH}, {X: -aHW, Y: aHH},
	}
	bCornersLocal := [4]rl.Vector2{
		{X: -bHW, Y: -bHH}, {X: bHW, Y: -bHH},
		{X: bHW, Y: bHH}, {X: -bHW, Y: bHH},
	}

	var aCorners, bCorners [4]rl.Vector2
	for i := 0; i < 4; i++ {
		aCorners[i] = rl.Vector2{
			X: a.Position.X + aCornersLocal[i].X*aCos - aCornersLocal[i].Y*aSin,
			Y: a.Position.Y + aCornersLocal[i].X*aSin + aCornersLocal[i].Y*aCos,
		}
		bCorners[i] = rl.Vector2{
			X: b.Position.X + bCornersLocal[i].X*bCos - bCornersLocal[i].Y*bSin,
			Y: b.Position.Y + bCornersLocal[i].X*bSin + bCornersLocal[i].Y*bCos,
		}
	}

	maxDepthA := float32(-math.MaxFloat32)
	var deepestA rl.Vector2

	for _, corner := range aCorners {
		depth := dotProduct(rl.Vector2Subtract(corner, b.Position), normal)
		if depth > maxDepthA {
			maxDepthA = depth
			deepestA = corner
		}
	}

	maxDepthB := float32(-math.MaxFloat32)
	var deepestB rl.Vector2

	for _, corner := range bCorners {
		depth := -dotProduct(rl.Vector2Subtract(corner, a.Position), normal)
		if depth > maxDepthB {
			maxDepthB = depth
			deepestB = corner
		}
	}

	return deepestA, deepestB
}

func dotProduct(a, b rl.Vector2) float32 {
	return a.X*b.X + a.Y*b.Y
}

func (a *Body) CollidesWith(b *Body) (bool, Contact) {
	switch a.Shape.(type) {
	case Circle:
		switch b.Shape.(type) {
		case Circle:
			return CircleVsCircle(a, b)
		case Box:
			return CircleVsBox(a, b)
		}
	case Box:
		switch b.Shape.(type) {
		case Circle:
			ok, contact := CircleVsBox(b, a)
			if ok {

				contact.PointA, contact.PointB = contact.PointB, contact.PointA
				contact.Normal = rl.Vector2Scale(contact.Normal, -1)
			}
			return ok, contact
		case Box:
			return BoxVsBox(a, b)
		}
	}
	return false, Contact{}
}
