package model

import (
	"game/vec2"
	"game/vec3"
	"log"
)

var UnitCube = [12][3]vec3.Value{
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

func (m *Model) Cube(from, to vec3.Value, tileX int, tileY int) {
	fx, tx := from.X, to.X
	fy, ty := from.Y, to.Y
	fz, tz := from.Z, to.Z

	if fx > tx {
		fx, tx = tx, fx
	}
	if fy > ty {
		fy, ty = ty, fy
	}
	if fz > tz {
		fz, tz = tz, fz
	}

	for ti := range 12 {
		t := UnitCube[ti]

		for pi := range 3 {
			p := &t[pi]

			if p.X == 0 {
				p.X = fx
			} else {
				p.X = tx
			}

			if p.Y == 0 {
				p.Y = fy
			} else {
				p.Y = ty
			}

			if p.Z == 0 {
				p.Z = fz
			} else {
				p.Z = tz
			}
		}

		normal := t[1].Subtract(t[0]).CrossProduct(t[2].Subtract(t[0])).Normalize()

		uvOrigin := vec2.XY(float64(tileX), float64(tileY))
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

func (m *Model) Export() {

	err := m.generator.Export(m.path, m.matPath, m.matName)
	if err != nil {
		log.Fatal(err)
	}
}
