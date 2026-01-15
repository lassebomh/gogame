package game

import (
	"fmt"
	"image/color"
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

func (u *ShaderUniform) SetVec2(x, y float32) {
	rl.SetShaderValue(u.Shader, u.Location, []float32{x, y}, rl.ShaderUniformVec2)
}
func (u *ShaderUniform) SetVec3(x, y, z float32) {
	rl.SetShaderValue(u.Shader, u.Location, []float32{x, y, z}, rl.ShaderUniformVec3)
}
func (u *ShaderUniform) SetVec4(x, y, z, w float32) {
	rl.SetShaderValue(u.Shader, u.Location, []float32{x, y, z, w}, rl.ShaderUniformVec4)
}
func (u *ShaderUniform) SetTexture(texture rl.Texture2D) {
	rl.SetShaderValueTexture(u.Shader, u.Location, texture)
}

func GetUniform(shader rl.Shader, format string, args ...any) *ShaderUniform {

	uniform := &ShaderUniform{
		Location: rl.GetShaderLocation(shader, fmt.Sprintf(format, args...)),
		Shader:   shader,
	}

	return uniform
}

// IChannel0Location:          GetUniform(backgroundShader, "iChannel0"),
// IChannel1Location:          GetUniform(backgroundShader, "iChannel1"),
// BackgroundShaderTime:       GetUniform(backgroundShader, "iTime"),
// BackgroundShaderFov:        GetUniform(backgroundShader, "iFov"),
// BackgroundShaderResolution: GetUniform(backgroundShader, "iResolution"),

// TileAA: GetUniform(shader, "tileAA"),
// TileBB: GetUniform(shader, "tileBB"),

type PlanetShaderUniforms struct {
	Channel0   UniformTexture `glsl:"iChannel0"`
	Channel1   UniformTexture `glsl:"iChannel1"`
	Time       UniformFloat   `glsl:"iTime"`
	Fov        UniformFloat   `glsl:"iFov"`
	Resolution UniformVec2    `glsl:"iResolution"`

	// AtlasLoc                   UniformTexture `glsl:"atlas"`
	// TileAA                     UniformVec2    `glsl:"tileAA"`
	// TileBB                     UniformVec2    `glsl:"tileBB"`
}

type Render struct {
	Shader rl.Shader
	// BackgroundShader rl.Shader
	Lights []*Light
	LightI int

	PlanetShader *Shader[PlanetShaderUniforms]

	// IChannel0Location          *ShaderUniform
	// IChannel1Location          *ShaderUniform
	// BackgroundShaderTime       *ShaderUniform
	// BackgroundShaderFov        *ShaderUniform
	// BackgroundShaderResolution *ShaderUniform
	// AtlasLoc                   *ShaderUniform
	TileAA *ShaderUniform
	TileBB *ShaderUniform

	Models   map[string]rl.Model
	Textures map[string]rl.Texture2D

	RenderWidth  int32
	RenderHeight int32
}

func NewRender(renderWidth int32, renderHeight int32) *Render {

	// backgroundShader := rl.LoadShader("", "./glsl330/planet2.fs")
	shader := rl.LoadShader("./glsl330/lighting.vs", "./glsl330/lighting.fs")

	*shader.Locs = rl.GetShaderLocation(shader, "viewPos")
	rl.SetShaderValue(shader, rl.GetShaderLocation(shader, "ambient"), []float32{0.1, 0.1, 0.1, 1.0}, rl.ShaderUniformVec4)

	render := &Render{
		Shader:   shader,
		Lights:   make([]*Light, 0, MAX_LIGHTS),
		Models:   map[string]rl.Model{},
		Textures: map[string]rl.Texture2D{},

		PlanetShader: NewShader[PlanetShaderUniforms]("", "./glsl330/planet2.fs"),

		// BackgroundShader: backgroundShader,
		// IChannel0Location:          GetUniform(backgroundShader, "iChannel0"),
		// IChannel1Location:          GetUniform(backgroundShader, "iChannel1"),
		// BackgroundShaderTime:       GetUniform(backgroundShader, "iTime"),
		// BackgroundShaderFov:        GetUniform(backgroundShader, "iFov"),
		// BackgroundShaderResolution: GetUniform(backgroundShader, "iResolution"),
		TileAA: GetUniform(shader, "tileAA"),
		TileBB: GetUniform(shader, "tileBB"),

		RenderWidth:  renderWidth,
		RenderHeight: renderHeight,
	}

	render.LoadTexture("organic", "./models/organic.png")
	render.LoadTexture("earth_elevation", "./models/earth_elevation.png")
	render.LoadTexture("atlas", "./models/atlas.png")

	// render.LoadModel("floor", "./models/floor.glb")
	render.LoadModel("plane", "./models/plane2.glb")
	render.LoadModel("wall", "./models/cube.glb")
	render.LoadModel("door", "./models/door.glb")
	render.LoadModel("monster_arm_segment", "./models/monster/monster_arm_segment.glb")
	render.LoadModel("monster_body", "./models/monster/monster_body.glb")

	for i := range MAX_LIGHTS {
		render.Lights = append(render.Lights, &Light{
			Type:           LIGHT_SPOT,
			Position:       rl.NewVector3(0, 0, 0),
			Target:         rl.NewVector3(0, 0, 0),
			Color:          color.RGBA{},
			Enabled:        0,
			Strength:       0,
			CutOff:         float32(math.Cos(0 * rl.Deg2rad)),
			OuterCutOff:    float32(math.Cos(30 * rl.Deg2rad)),
			enabledLoc:     GetUniform(shader, "lights[%d].enabled", i),
			lightTypeLoc:   GetUniform(shader, "lights[%d].type", i),
			positionLoc:    GetUniform(shader, "lights[%d].position", i),
			targetLoc:      GetUniform(shader, "lights[%d].target", i),
			colorLoc:       GetUniform(shader, "lights[%d].color", i),
			cutOffLoc:      GetUniform(shader, "lights[%d].cutOff", i),
			outerCutOffLoc: GetUniform(shader, "lights[%d].outerCutOff", i),
			strengthLoc:    GetUniform(shader, "lights[%d].strength", i),
		})
	}

	return render
}

func (r *Render) LoadModel(name string, path string) {
	model := rl.LoadModel(path)
	model.Materials.Shader = r.Shader

	for i := range model.MaterialCount {
		mat := &model.GetMaterials()[i]
		mat.Shader = r.Shader
		rl.SetMaterialTexture(mat, rl.MapDiffuse, r.Textures["atlas"])
		rl.SetMaterialTexture(mat, rl.MapAlbedo, r.Textures["atlas"])
	}

	rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, r.Textures["atlas"])
	rl.SetMaterialTexture(model.Materials, rl.MapAlbedo, r.Textures["atlas"])
	r.Models[name] = model
}

func (r *Render) LoadTexture(name string, path string) {
	texture := rl.LoadTexture(path)
	rl.SetTextureWrap(texture, rl.WrapRepeat)
	r.Textures[name] = texture
}

func (r *Render) Unload() {
	for _, model := range r.Models {
		rl.UnloadModel(model)
	}
	for _, texture := range r.Textures {
		rl.UnloadTexture(texture)
	}
	rl.UnloadShader(r.Shader)
	// rl.UnloadShader(r.BackgroundShader)
	r.PlanetShader.Unload()
}

func (r *Render) DrawModel(model rl.Model, tileX float64, tileY float64, pos Vec, scale Vec, rotationAxis Vec, rotationRadians float32) {
	atlas := r.Textures["atlas"]
	width, _ := float64(atlas.Width), float64(atlas.Height)
	s := 64 / width
	x, y := float64(tileX)*s, float64(tileY)*s
	// r.PlanetShader.Uniform.TileAA.Set(x, y)
	// r.PlanetShader.Uniform.TileBB.Set(x+s, y+s)
	r.TileAA.SetVec2(float32(x), float32(y))
	r.TileBB.SetVec2(float32(x+s), float32(y+s))
	rl.DrawModelEx(model, pos.Vector3, rotationAxis.Vector3, rotationRadians, scale.Vector3, rl.White)
}

func (r *Render) LightDirectional(direction Vec, color rl.Color, strength float32) {
	light := r.Lights[r.LightI]
	light.Enabled = 1
	light.Type = LIGHT_DIRECTIONAL
	light.Position = rl.Vector3{}
	light.Target = direction.Vector3
	light.Color = color
	light.Strength = strength

	r.LightI++
}

func (r *Render) LightSpot(position Vec, target Vec, cutoff float32, outerCutOff float32, color rl.Color, strength float32) {
	light := r.Lights[r.LightI]
	light.Enabled = 1
	light.Type = LIGHT_SPOT
	light.Position = position.Vector3
	light.Target = target.Vector3
	light.CutOff = float32(math.Cos(float64(cutoff * rl.Deg2rad)))
	light.OuterCutOff = float32(math.Cos(float64(outerCutOff * rl.Deg2rad)))
	light.Color = color
	light.Strength = strength

	r.LightI++
}

func (r *Render) LightPoint(position Vec, color rl.Color, strength float32) {
	light := r.Lights[r.LightI]
	light.Enabled = 1
	light.Type = LIGHT_POINT
	light.Position = position.Vector3
	light.Color = color
	light.Strength = strength

	r.LightI++
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
			lt.Enabled = 0
		}
	}
	r.LightI = 0
}
