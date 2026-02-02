package model

import (
	"bufio"
	"fmt"
	"game/vec2"
	"game/vec3"
	"os"
	"path/filepath"
)

type ObjGenerator struct {
	Verts       map[vec3.Value]uint
	VertNormals map[vec3.Value]uint
	VertUVs     map[vec2.Value]uint

	VertCount       uint
	VertNormalCount uint
	VertUVCount     uint

	Faces [][3][3]uint
}

func NewModelObject() *ObjGenerator {
	return &ObjGenerator{

		Verts:       map[vec3.Value]uint{},
		VertNormals: map[vec3.Value]uint{},
		VertUVs:     map[vec2.Value]uint{},

		VertCount:       0,
		VertNormalCount: 0,
		VertUVCount:     0,

		Faces: [][3][3]uint{},
	}
}

func (m *ObjGenerator) AddTriangle(a, b, c vec3.Value, uva, uvb, uvc vec2.Value, normal vec3.Value) {
	ni, ok := m.VertNormals[normal]
	if !ok {
		m.VertNormalCount++
		ni = m.VertNormalCount
		m.VertNormals[normal] = m.VertNormalCount
	}

	ai, ok := m.Verts[a]
	if !ok {
		m.VertCount++
		ai = m.VertCount
		m.Verts[a] = m.VertCount
	}

	bi, ok := m.Verts[b]
	if !ok {
		m.VertCount++
		bi = m.VertCount
		m.Verts[b] = m.VertCount
	}

	ci, ok := m.Verts[c]
	if !ok {
		m.VertCount++
		ci = m.VertCount
		m.Verts[c] = m.VertCount
	}

	uvai, ok := m.VertUVs[uva]
	if !ok {
		m.VertUVCount++
		uvai = m.VertUVCount
		m.VertUVs[uva] = m.VertUVCount
	}

	uvbi, ok := m.VertUVs[uvb]
	if !ok {
		m.VertUVCount++
		uvbi = m.VertUVCount
		m.VertUVs[uvb] = m.VertUVCount
	}

	uvci, ok := m.VertUVs[uvc]
	if !ok {
		m.VertUVCount++
		uvci = m.VertUVCount
		m.VertUVs[uvc] = m.VertUVCount
	}

	m.Faces = append(m.Faces, [3][3]uint{
		{ai, uvai, ni},
		{bi, uvbi, ni},
		{ci, uvci, ni},
	})
}

func (m *ObjGenerator) Export(path string, matPath string, matName string) (err error) {

	matPathRel, err := filepath.Rel(filepath.Dir(path), matPath)
	if err != nil {
		return
	}

	file, err := os.Create(path)
	if err != nil {
		return
	}
	w := bufio.NewWriter(file)

	fmt.Fprintf(w, "mtllib %s\n", matPathRel)
	fmt.Fprintf(w, "o %s\n", "MyModel")

	verts := make([]vec3.Value, m.VertCount, m.VertCount)
	for v, i := range m.Verts {
		verts[i-1] = v
	}
	for _, v := range verts {
		fmt.Fprintf(w, "v %.6f %.6f %.6f\n", v.X, v.Y, v.Z)
	}

	normals := make([]vec3.Value, m.VertNormalCount, m.VertNormalCount)
	for v, i := range m.VertNormals {
		normals[i-1] = v
	}
	for _, v := range normals {
		fmt.Fprintf(w, "vn %.6f %.6f %.6f\n", v.X, v.Y, v.Z)
	}

	uvs := make([]vec2.Value, m.VertUVCount, m.VertUVCount)
	for v, i := range m.VertUVs {
		uvs[i-1] = v
	}
	for _, v := range uvs {
		fmt.Fprintf(w, "vt %.6f %.6f\n", v.X, v.Y)
	}

	fmt.Fprint(w, "s 0\n")
	fmt.Fprintf(w, "usemtl %s\n", matName)

	for _, face := range m.Faces {
		fmt.Fprint(w, "f")
		for _, idxs := range face {
			fmt.Fprintf(w, " %d/%d/%d", idxs[0], idxs[1], idxs[2])
		}
		fmt.Fprint(w, "\n")
	}

	err = w.Flush()
	if err != nil {
		return
	}

	err = file.Close()
	if err != nil {
		return
	}

	return
}
