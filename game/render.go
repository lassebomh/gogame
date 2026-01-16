package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LightType int32

const (
	LIGHT_DIRECTIONAL LightType = iota
	LIGHT_POINT
	LIGHT_SPOT
)
const MAX_LIGHTS = 4

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

type MainShader struct {
	shader rl.Shader

	LightI int

	TileAA UniformVec2 `glsl:"tileAA"`
	TileBB UniformVec2 `glsl:"tileBB"`

	Ambient UniformVec4 `glsl:"ambient"`

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

func (m *MainShader) LightDirectional(direction Vec, color rl.Color, strength float64) {
	m.LightEnabled[m.LightI].Set(1)
	m.LightType[m.LightI].Set(int32(LIGHT_DIRECTIONAL))
	m.LightPosition[m.LightI].SetVec(NewVec(0, 0, 0))
	m.LightTarget[m.LightI].SetVec(direction)
	m.LightColor[m.LightI].SetColor(color)
	m.LightStrength[m.LightI].Set(strength)

	m.LightI++
}

func (m *MainShader) LightSpot(position Vec, target Vec, cutoff float64, outerCutOff float64, color rl.Color, strength float64) {
	m.LightEnabled[m.LightI].Set(1)
	m.LightType[m.LightI].Set(int32(LIGHT_SPOT))
	m.LightPosition[m.LightI].SetVec(position)
	m.LightTarget[m.LightI].SetVec(target)
	m.LightCutOff[m.LightI].Set(math.Cos(float64(cutoff * rl.Deg2rad)))
	m.LightOuterCutOff[m.LightI].Set(math.Cos(float64(outerCutOff * rl.Deg2rad)))
	m.LightColor[m.LightI].SetColor(color)
	m.LightStrength[m.LightI].Set(strength)

	m.LightI++
}

func (m *MainShader) LightPoint(position Vec, color rl.Color, strength float64) {
	m.LightEnabled[m.LightI].Set(1)
	m.LightType[m.LightI].Set(int32(LIGHT_POINT))
	m.LightPosition[m.LightI].SetVec(position)
	m.LightColor[m.LightI].SetColor(color)
	m.LightStrength[m.LightI].Set(strength)

	m.LightI++
}

func (m *MainShader) UpdateValues() {

	for i := m.LightI + 1; i < MAX_LIGHTS; i++ {
		m.LightEnabled[m.LightI].Set(0)
	}
	m.LightI = 0
}

type Render struct {
	MainShader *MainShader

	PlanetShader *PlanetShader

	Models   map[string]rl.Model
	Textures map[string]rl.Texture2D

	RenderWidth  int32
	RenderHeight int32
}

func NewRender(renderWidth int32, renderHeight int32) *Render {

	render := &Render{
		MainShader:   &MainShader{},
		PlanetShader: &PlanetShader{},

		Models:   map[string]rl.Model{},
		Textures: map[string]rl.Texture2D{},

		RenderWidth:  renderWidth,
		RenderHeight: renderHeight,
	}

	InstantiateShader(render.MainShader, "./glsl330/lighting.vs", "./glsl330/lighting.fs")
	InstantiateShader(render.PlanetShader, "", "./glsl330/planet2.fs")

	render.LoadTexture("organic", "./models/organic.png")
	render.LoadTexture("earth_elevation", "./models/earth_elevation.png")
	render.LoadTexture("atlas", "./models/atlas.png")

	render.LoadModel("floor", "./models/floor.glb")
	render.LoadModel("plane", "./models/plane2.glb")
	render.LoadModel("wall", "./models/cube.glb")
	render.LoadModel("door", "./models/door.glb")
	render.LoadModel("monster_arm_segment", "./models/monster/monster_arm_segment.glb")
	render.LoadModel("monster_body", "./models/monster/monster_body.glb")

	return render
}

func (r *Render) LoadModel(name string, path string) {
	model := rl.LoadModel(path)
	model.Materials.Shader = r.MainShader.GetRaylibShader()

	for i := range model.MaterialCount {
		mat := &model.GetMaterials()[i]
		mat.Shader = r.MainShader.GetRaylibShader()
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
	rl.UnloadShader(r.MainShader.GetRaylibShader())
	rl.UnloadShader(r.PlanetShader.GetRaylibShader())
}

func (r *Render) DrawModel(model rl.Model, tileX float64, tileY float64, pos Vec, scale Vec, rotationAxis Vec, rotationRadians float32) {
	atlas := r.Textures["atlas"]
	width, _ := float64(atlas.Width), float64(atlas.Height)
	s := 64 / width
	x, y := float64(tileX)*s, float64(tileY)*s
	r.MainShader.TileAA.Set(x, y)
	r.MainShader.TileBB.Set(x+s, y+s)
	rl.DrawModelEx(model, pos.Vector3, rotationAxis.Vector3, rotationRadians, scale.Vector3, rl.White)
}
