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

func (m *Model) AddMesh(mesh Mesh, tileX int, tileY int, rotation int) {

	aa, _ := mesh.GetAABB()

	// fx, tx := aa.X, bb.X
	// fz, tz := aa.Z, bb.Z
	// fy, _ := aa.Y, bb.Y

	for _, t := range mesh.vertices {
		a, b, c := t[0], t[1], t[2]
		normal := b.Subtract(a).CrossProduct(c.Subtract(a)).Normalize()

		uvOrigin := vec2.XY(float64(tileX-1), m.tiles-float64(tileY-1)-1)
		uvs := [3]vec2.Value{{0, 0}, {0, 1}, {1, 0}}

		for i, uv := range uvs {
			vp := t[i].Subtract(aa)
			absN := normal.Abs()

			u, v := 0., 0.

			if absN.X > absN.Y && absN.X > absN.Z {
				u, v = vp.Z, vp.Y
			} else if absN.Y > absN.X && absN.Y > absN.Z {
				u, v = vp.X, vp.Z
			} else {
				u, v = vp.X, vp.Y
			}

			uv.X = u
			uv.Y = v

			if normal.Y != 0 {
				switch rotation {
				case 2:
					uv.X = 1 - u
					uv.Y = v
				case 3:
					uv.X = 1 - v
					uv.Y = 1 - u
				case 0:
					uv.X = u
					uv.Y = 1 - v
				case 1:
					uv.X = v
					uv.Y = u
				}
			}

			uvs[i] = uvOrigin.Add(uv).Scale(1 / m.tiles)
		}

		m.generator.AddTriangle(t[0], t[1], t[2], uvs[0], uvs[1], uvs[2], normal)
	}
}

func (m *Model) Export() {

	err := m.generator.Export(m.path, m.matPath, m.matName)
	if err != nil {
		log.Fatal(err)
	}
}
