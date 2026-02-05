package model

import (
	"game/vec2"
	"game/vec3"
	"log"
	"math"
)

var UnitCubeVerticies = [12][3]vec3.Value{
	// top
	{{0, 1, 0}, {1, 1, 1}, {1, 1, 0}},
	{{0, 1, 0}, {0, 1, 1}, {1, 1, 1}},
	// bottom
	{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}},
	{{0, 0, 0}, {1, 0, 1}, {0, 0, 1}},
	// front
	{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}},
	{{0, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	// back
	{{0, 0, 0}, {1, 1, 0}, {1, 0, 0}},
	{{0, 0, 0}, {0, 1, 0}, {1, 1, 0}},
	// left
	{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}},
	{{0, 0, 0}, {0, 1, 1}, {0, 1, 0}},
	// right
	{{1, 0, 0}, {1, 1, 1}, {1, 0, 1}},
	{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}},
}

var UnitCube Mesh

func init() {
	UnitCube = NewMesh()
	UnitCube.vertices = append(UnitCube.vertices, UnitCubeVerticies[:]...)
}

type Mesh struct {
	vertices [][3]vec3.Value
	aa       vec3.Value
	bb       vec3.Value
	dirty    bool
}

func NewMesh() Mesh {
	return Mesh{
		vertices: make([][3]vec3.Value, 0),
		dirty:    true,
	}
}

func (m *Mesh) refreshAABB() {

	fx, tx := math.Inf(1), -math.Inf(1)
	fz, tz := math.Inf(1), -math.Inf(1)
	fy, ty := math.Inf(1), -math.Inf(1)

	for _, t := range m.vertices {
		for _, v := range t {
			if fx > v.X {
				fx = v.X
			}
			if tx < v.X {
				tx = v.X
			}

			if fy > v.Y {
				fy = v.Y
			}
			if ty < v.Y {
				ty = v.Y
			}

			if fz > v.Z {
				fz = v.Z
			}
			if tz < v.Z {
				tz = v.Z
			}
		}
	}

	m.aa = vec3.XYZ(fx, fy, fz)
	m.bb = vec3.XYZ(tx, ty, tz)

	m.dirty = false
}

func (m Mesh) Transform(mat vec3.Matrix) Mesh {
	m2 := NewMesh()
	m2.vertices = append(m2.vertices, m.vertices...)
	for i, t := range m.vertices {
		for ii, v := range t {
			m2.vertices[i][ii] = v.Transform(mat)
		}
	}
	m2.dirty = true
	return m2
}

func (m *Mesh) Combine(o Mesh) {
	m.vertices = append(m.vertices, o.vertices...)
	m.dirty = true
}

func (m *Mesh) GetAABB() (vec3.Value, vec3.Value) {
	if m.dirty {
		m.refreshAABB()
	}
	return m.aa, m.bb
}

type Model struct {
	generator *ObjGenerator
	tiles     float64
	path      string
	matPath   string
	matName   string
}

func NewModel(tiles int, path string, matPath string, matName string) *Model {
	return &Model{
		generator: NewModelObject(),
		tiles:     float64(tiles),
		path:      path,
		matPath:   matPath,
		matName:   matName,
	}
}

func (m *Model) AddMesh(mesh Mesh, tileX int, tileY int) {

	aa, bb := mesh.GetAABB()

	fx, tx := aa.X, bb.X
	fz, tz := aa.Z, bb.Z
	fy, ty := aa.Y, bb.Y

	for _, t := range mesh.vertices {
		for _, v := range t {
			if fx > v.X {
				fx = v.X
			}
			if tx < v.X {
				tx = v.X
			}

			if fy > v.Y {
				fy = v.Y
			}
			if ty < v.Y {
				ty = v.Y
			}

			if fz > v.Z {
				fz = v.Z
			}
			if tz < v.Z {
				tz = v.Z
			}
		}
	}

	for _, t := range mesh.vertices {
		normal := t[1].Subtract(t[0]).CrossProduct(t[2].Subtract(t[0])).Normalize()

		uvOrigin := vec2.XY(float64(tileX-1), m.tiles-float64(tileY-1)-1)
		uvs := [3]vec2.Value{}

		for i := range uvs {
			var uv vec2.Value

			if normal.Z == 1 || normal.Y == 1 {
				uv.X += t[i].X - fx
			}
			if normal.Y == -1 || normal.Z == -1 {
				uv.X += tx - t[i].X
			}
			if normal.X == 1 {
				uv.X += tz - t[i].Z
			}
			if normal.X == -1 {
				uv.X += t[i].Z - fz
			}
			if normal.Y == 0 {
				uv.Y += t[i].Y - fy
			}
			if normal.Y == 1 {
				uv.Y += tz - t[i].Z
			}
			if normal.Y == -1 {
				uv.Y += t[i].Z - fz
			}

			uvs[i] = uvOrigin.Add(uv).Scale(1 / m.tiles)
		}

		m.generator.AddTriangle(t[0], t[1], t[2], uvs[0], uvs[1], uvs[2], normal)
	}
}

// func (m *Model) Cube(from, to vec3.Value, tileX int, tileY int) {
// 	mesh := NewCube(from, to)
// 	m.AddMesh(mesh, tileX, tileY)

// 	// fx, tx := from.X, to.X
// 	// fy, ty := from.Y, to.Y
// 	// fz, tz := from.Z, to.Z

// 	// if fx > tx {
// 	// 	fx, tx = tx, fx
// 	// }
// 	// if fy > ty {
// 	// 	fy, ty = ty, fy
// 	// }
// 	// if fz > tz {
// 	// 	fz, tz = tz, fz
// 	// }

// 	// for ti := range 12 {
// 	// 	t := UnitCube[ti]

// 	// 	for pi := range 3 {
// 	// 		p := &t[pi]

// 	// 		if p.X == 0 {
// 	// 			p.X = fx
// 	// 		} else {
// 	// 			p.X = tx
// 	// 		}

// 	// 		if p.Y == 0 {
// 	// 			p.Y = fy
// 	// 		} else {
// 	// 			p.Y = ty
// 	// 		}

// 	// 		if p.Z == 0 {
// 	// 			p.Z = fz
// 	// 		} else {
// 	// 			p.Z = tz
// 	// 		}
// 	// 	}

// 	// 	normal := t[1].Subtract(t[0]).CrossProduct(t[2].Subtract(t[0])).Normalize()

// 	// 	uvOrigin := vec2.XY(float64(tileX), m.tiles-float64(tileY)-1)
// 	// 	uvs := [3]vec2.Value{}

// 	// 	for i := range uvs {
// 	// 		var uv vec2.Value

// 	// 		if normal.Z == 1 || normal.Y == 1 {
// 	// 			uv.X += t[i].X - fx
// 	// 		}
// 	// 		if normal.Y == -1 || normal.Z == -1 {
// 	// 			uv.X += tx - t[i].X
// 	// 		}
// 	// 		if normal.X == 1 {
// 	// 			uv.X += tz - t[i].Z
// 	// 		}
// 	// 		if normal.X == -1 {
// 	// 			uv.X += t[i].Z - fz
// 	// 		}
// 	// 		if normal.Y == 0 {
// 	// 			uv.Y += t[i].Y - fy
// 	// 		}
// 	// 		if normal.Y == 1 {
// 	// 			uv.Y += tz - t[i].Z
// 	// 		}
// 	// 		if normal.Y == -1 {
// 	// 			uv.Y += t[i].Z - fz
// 	// 		}

// 	// 		uvs[i] = uvOrigin.Add(uv).Scale(1 / m.tiles)
// 	// 	}

// 	// 	m.generator.AddTriangle(t[0], t[1], t[2], uvs[0], uvs[1], uvs[2], normal)
// 	// }

// }

func (m *Model) Export() {

	err := m.generator.Export(m.path, m.matPath, m.matName)
	if err != nil {
		log.Fatal(err)
	}
}
