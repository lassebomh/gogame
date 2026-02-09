package game3

import (
	"game/vec3"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LightType int32

const (
	LIGHT_DIRECTIONAL LightType = iota
	LIGHT_POINT
	LIGHT_SPOT
)

const MAX_LIGHTS = 8

type MainShader struct {
	shader rl.Shader

	LightI int

	// UVClamp UniformVec4 `glsl:"uvClamp"`

	ShadowMap UniformTexture `glsl:"shadowMap"`

	PlayerPosition UniformVec3 `glsl:"playerPosition"`
	// PlayerViewResolution UniformVec2 `glsl:"iResolution"`
	Visibility UniformFloat `glsl:"visibility"`

	HideOutsideView UniformFloat `glsl:"hideOutsideView"`
	FullBright      UniformInt   `glsl:"fullBright"`
	// Ambient          UniformVec4              `glsl:"ambient"`

	LightEnabled     [MAX_LIGHTS]UniformInt   `glsl:"lights[%d].enabled"`
	LightType        [MAX_LIGHTS]UniformInt   `glsl:"lights[%d].type"`
	LightPosition    [MAX_LIGHTS]UniformVec3  `glsl:"lights[%d].position"`
	LightTarget      [MAX_LIGHTS]UniformVec3  `glsl:"lights[%d].target"`
	LightColor       [MAX_LIGHTS]UniformVec4  `glsl:"lights[%d].color"`
	LightCutOff      [MAX_LIGHTS]UniformFloat `glsl:"lights[%d].cutOff"`
	LightOuterCutOff [MAX_LIGHTS]UniformFloat `glsl:"lights[%d].outerCutOff"`
	LightStrength    [MAX_LIGHTS]UniformFloat `glsl:"lights[%d].strength"`
}

func (m *MainShader) GetRaylibShader() rl.Shader {
	return m.shader
}

func (m *MainShader) SetRaylibShader(shader rl.Shader) {
	m.shader = shader
}

func (m *MainShader) LightDirectional(direction vec3.Value, color rl.Color, strength float64) {
	m.LightI++
	m.LightEnabled[m.LightI].Set(1)
	m.LightType[m.LightI].Set(int32(LIGHT_DIRECTIONAL))
	m.LightPosition[m.LightI].Set(0, 0, 0)
	m.LightTarget[m.LightI].SetVec3(direction)
	m.LightColor[m.LightI].SetColor(color)
	m.LightStrength[m.LightI].Set(strength)

}

func (m *MainShader) LightSpot(position vec3.Value, target vec3.Value, cutoff float64, outerCutOff float64, color rl.Color, strength float64) {
	m.LightI++
	m.LightEnabled[m.LightI].Set(1)
	m.LightType[m.LightI].Set(int32(LIGHT_SPOT))
	m.LightPosition[m.LightI].SetVec3(position)
	m.LightTarget[m.LightI].SetVec3(target)
	m.LightCutOff[m.LightI].Set(math.Cos(float64(cutoff * rl.Deg2rad)))
	m.LightOuterCutOff[m.LightI].Set(math.Cos(float64(outerCutOff * rl.Deg2rad)))
	m.LightColor[m.LightI].SetColor(color)
	m.LightStrength[m.LightI].Set(strength)

}

func (m *MainShader) LightPoint(position vec3.Value, color rl.Color, strength float64) {
	m.LightI++
	m.LightEnabled[m.LightI].Set(1)
	m.LightType[m.LightI].Set(int32(LIGHT_POINT))
	m.LightPosition[m.LightI].SetVec3(position)
	m.LightColor[m.LightI].SetColor(color)
	m.LightStrength[m.LightI].Set(strength)

}

func (m *MainShader) UpdateLights() {

	for i := m.LightI + 1; i < MAX_LIGHTS; i++ {
		m.LightEnabled[i].Set(0)
	}
	m.LightI = -1
}

type PlanetShader struct {
	shader rl.Shader

	Channel0   UniformTexture `glsl:"iChannel0"`
	Channel1   UniformTexture `glsl:"iChannel1"`
	Time       UniformFloat   `glsl:"iTime"`
	Fov        UniformFloat   `glsl:"iFov"`
	Resolution UniformVec2    `glsl:"iResolution"`
}

func (p *PlanetShader) GetRaylibShader() rl.Shader {
	return p.shader
}

func (p *PlanetShader) SetRaylibShader(shader rl.Shader) {
	p.shader = shader
}
