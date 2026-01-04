package game

import (
	"fmt"
	"math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LightType int32

const (
	LIGHT_DIRECTIONAL LightType = iota
	LIGHT_POINT
	LIGHT_SPOT
)
const MAX_LIGHTS = 4

type ShaderUniform struct {
	Location int32
	Shader   rl.Shader
}

func (u *ShaderUniform) SetInt(value int32) {
	rl.SetShaderValue(u.Shader, u.Location, unsafe.Slice((*float32)(unsafe.Pointer(&value)), 4), rl.ShaderUniformInt)
}
func (u *ShaderUniform) SetFloat(value float32) {
	rl.SetShaderValue(u.Shader, u.Location, []float32{value}, rl.ShaderUniformFloat)
}
func (u *ShaderUniform) SetVec3(x, y, z float32) {
	rl.SetShaderValue(u.Shader, u.Location, []float32{x, y, z}, rl.ShaderUniformVec3)
}
func (u *ShaderUniform) SetVec4(x, y, z, w float32) {
	rl.SetShaderValue(u.Shader, u.Location, []float32{x, y, z, w}, rl.ShaderUniformVec4)
}

func (r *Render) GetUniform(shader rl.Shader, format string, args ...any) *ShaderUniform {

	uniform := &ShaderUniform{
		Location: rl.GetShaderLocation(shader, fmt.Sprintf(format, args...)),
		Shader:   shader,
	}

	return uniform
}

type Render struct {
	Shader rl.Shader
	Lights []*Light
}

func NewRender(shader rl.Shader) *Render {
	render := &Render{
		Shader: shader,
		Lights: make([]*Light, 0, MAX_LIGHTS),
	}

	return render
}

func (r *Render) NewLight(lightType LightType, position rl.Vector3, target rl.Vector3, color rl.Color, strength float32) *Light {
	lightsCount := len(r.Lights)

	light := &Light{
		Type:           lightType,
		Position:       position,
		Target:         target,
		Color:          color,
		Enabled:        1,
		Strength:       strength,
		CutOff:         float32(math.Cos(0 * rl.Deg2rad)),
		OuterCutOff:    float32(math.Cos(30 * rl.Deg2rad)),
		enabledLoc:     r.GetUniform(r.Shader, "lights[%d].enabled", lightsCount),
		lightTypeLoc:   r.GetUniform(r.Shader, "lights[%d].type", lightsCount),
		positionLoc:    r.GetUniform(r.Shader, "lights[%d].position", lightsCount),
		targetLoc:      r.GetUniform(r.Shader, "lights[%d].target", lightsCount),
		colorLoc:       r.GetUniform(r.Shader, "lights[%d].color", lightsCount),
		cutOffLoc:      r.GetUniform(r.Shader, "lights[%d].cutOff", lightsCount),
		outerCutOffLoc: r.GetUniform(r.Shader, "lights[%d].outerCutOff", lightsCount),
		strengthLoc:    r.GetUniform(r.Shader, "lights[%d].strength", lightsCount),
	}

	r.Lights = append(r.Lights, light)

	return light
}

type Light struct {
	Type        LightType
	Position    rl.Vector3
	Target      rl.Vector3
	Color       rl.Color
	CutOff      float32
	OuterCutOff float32
	Strength    float32
	Enabled     int32
	// shader locations
	enabledLoc     *ShaderUniform
	lightTypeLoc   *ShaderUniform
	positionLoc    *ShaderUniform
	targetLoc      *ShaderUniform
	colorLoc       *ShaderUniform
	cutOffLoc      *ShaderUniform
	outerCutOffLoc *ShaderUniform
	strengthLoc    *ShaderUniform
}

func (r *Render) UpdateValues() {
	for _, lt := range r.Lights {
		if lt.Enabled != 0 {
			lt.enabledLoc.SetInt(lt.Enabled)
			lt.lightTypeLoc.SetInt(int32(lt.Type))
			lt.cutOffLoc.SetFloat(lt.CutOff)
			lt.outerCutOffLoc.SetFloat(lt.OuterCutOff)
			lt.strengthLoc.SetFloat(lt.Strength)
			lt.positionLoc.SetVec3(lt.Position.X, lt.Position.Y, lt.Position.Z)
			lt.targetLoc.SetVec3(lt.Target.X, lt.Target.Y, lt.Target.Z)
			lt.colorLoc.SetVec4(float32(lt.Color.R)/255, float32(lt.Color.G)/255, float32(lt.Color.B)/255, float32(lt.Color.A)/255)
		}
	}
}
