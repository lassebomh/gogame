package game3

import rl "github.com/gen2brain/raylib-go/raylib"

type MainShader struct {
	shader rl.Shader
}

func (m *MainShader) GetRaylibShader() rl.Shader {
	return m.shader
}

func (m *MainShader) SetRaylibShader(shader rl.Shader) {
	m.shader = shader
}
