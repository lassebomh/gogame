package game2

import (
	"encoding/gob"
	"image/color"
	"math"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type RenderFlags = int32

const (
	RENDER_FLAG_PHYSICS = RenderFlags(1 << iota)
	RENDER_FLAG_FULLBRIGHT
)

const PHYSICS_TICKRATE = time.Second / 60

type Game struct {
	Time                   time.Duration
	TimeDelta              time.Duration
	TimePhysicsAccumulator time.Duration

	Day float64

	Player  *Player
	Monster *Monster
	Space   *cp.Space
	Level   *Level
	Camera  Camera3D

	IsStation   bool
	RenderFlags RenderFlags

	EditorEnabled bool
	Editor        *Editor

	Tileset *Tileset

	MousePosition     Vec2
	MouseRayOrigin    Vec3
	MouseRayDirection Vec3

	MainTexture rl.RenderTexture2D
	MainShader  *MainShader

	Textures map[string]rl.Texture2D

	Models map[string]rl.Model
}

type GameSave struct {
	Time                   time.Duration
	TimeDelta              time.Duration
	TimePhysicsAccumulator time.Duration

	Player  PlayerSave
	Monster *Monster

	Level       Level
	RenderFlags RenderFlags

	EditorEnabled bool
	Editor        *Editor
}

func (g *Game) ToSave() GameSave {
	return GameSave{
		Time:                   g.Time,
		TimeDelta:              g.TimeDelta,
		TimePhysicsAccumulator: g.TimePhysicsAccumulator,
		Level:                  *g.Level,
		Player:                 g.Player.ToSave(g),
		Monster:                g.Monster.ToSave(g),
		RenderFlags:            g.RenderFlags,
		EditorEnabled:          g.EditorEnabled,
		Editor:                 g.Editor,
	}
}
func (g *Game) Update(dt time.Duration) {

	g.TimeDelta = dt
	g.Time += dt
	g.TimePhysicsAccumulator += dt

	g.Day += dt.Seconds() / 100

	for g.TimePhysicsAccumulator >= PHYSICS_TICKRATE {
		g.Space.Step(PHYSICS_TICKRATE.Seconds())
		g.TimePhysicsAccumulator -= PHYSICS_TICKRATE
	}

	g.Camera.Position = g.Player.Position3D().Add(NewVec3(0, 8, -5).Normalize().Scale(10))
	g.Camera.Target = g.Player.Position3D()

	mousePos := rl.GetMousePosition()
	g.MousePosition = Vec2FromRaylib(mousePos)

	mouseRay := rl.GetScreenToWorldRay(mousePos, g.Camera.Raylib())

	g.MouseRayOrigin = Vec3FromRaylib(mouseRay.Position)
	g.MouseRayDirection = Vec3FromRaylib(mouseRay.Direction)

	g.Player.Update(g)
	g.Monster.Update(g)

	cellWakeX := 8
	cellWakeZ := 8
	playerPos := g.Player.Position3D()

	for ix := range cellWakeX {
		for iz := range cellWakeZ {
			cellPos := playerPos.Add(NewVec3(
				float64(ix)-float64(cellWakeX-1)/2,
				0,
				float64(iz)-float64(cellWakeZ-1)/2,
			))

			cell := g.Level.GetCell(cellPos)
			cell.Wake(g)
		}
	}

}

func (g *Game) LoadModel(name string, path string, shader Shader, texture *rl.Texture2D) {
	model := rl.LoadModel(path)
	g.Models[name] = model
	mats := model.GetMaterials()
	for i := range mats {
		mat := &mats[i]
		mat.Shader = shader.GetRaylibShader()
		if texture != nil {
			mat.Maps.Texture = *texture
		}
	}
}

func (g *Game) GetModel(name string) rl.Model {

	return g.Models[name]
}

func (save GameSave) Load() *Game {
	downscale := int32(4)
	screenWidth, screenHeight := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())

	g := &Game{
		Time:                   save.Time,
		TimeDelta:              save.TimeDelta,
		TimePhysicsAccumulator: save.TimePhysicsAccumulator,

		RenderFlags:   save.RenderFlags,
		EditorEnabled: save.EditorEnabled,
		Editor:        save.Editor,

		Camera: Camera3D{
			Fovy:       6,
			Up:         Y,
			Projection: rl.CameraOrthographic,
		},

		Tileset: NewTileset("./models/atlas.png", 5),

		MainTexture: rl.LoadRenderTexture(int32(screenWidth/downscale), int32(screenHeight/downscale)),

		Models: map[string]rl.Model{},

		Textures:   map[string]rl.Texture2D{},
		MainShader: NewShader(&MainShader{}, "./glsl330/lighting.vs", "./glsl330/lighting.fs"),
	}

	rl.SetTextureFilter(g.MainTexture.Texture, rl.FilterPoint)

	g.Space = cp.NewSpace()
	g.Level = save.Level.Init()

	g.LoadModel("wallDebug", "./models/wallx.glb", g.MainShader, nil)
	g.LoadModel("wall", "./models/wallx.glb", g.MainShader, &g.Tileset.Texture)
	g.LoadModel("stair", "./models/stair.glb", g.MainShader, &g.Tileset.Texture)
	g.LoadModel("door", "./models/door.glb", g.MainShader, &g.Tileset.Texture)
	g.LoadModel("monster_arm_segment", "./models/monster/monster_arm_segment.glb", g.MainShader, &g.Tileset.Texture)
	g.LoadModel("monster_body", "./models/monster/monster_body.glb", g.MainShader, &g.Tileset.Texture)

	save.Player.Load(g)
	save.Monster.Load(g)

	return g
}

func NewGameSave() GameSave {
	gameSave := GameSave{
		Time:                   0,
		TimeDelta:              0,
		TimePhysicsAccumulator: 0,
		Player: PlayerSave{
			Position: NewVec2(0, 0),
		},
		Monster: &Monster{
			SavePosition: NewVec2(0, 0),
		},
		EditorEnabled: false,
		Editor:        NewEditor(),
		Level:         Level{},
	}

	return gameSave
}

func (g *Game) Draw() {
	// downscale := int32(8)
	screenWidth := int32(rl.GetScreenWidth())
	screenHeight := int32(rl.GetScreenHeight())

	g.Player.RenderViewTexture(g)

	BeginTextureMode(g.MainTexture, func() {
		rl.ClearBackground(rl.Black)

		BeginMode3D(g.Camera, func() {

			g.MainShader.ShadowMap.Set(g.Player.ViewTexture.Texture)
			g.MainShader.Resolution.Set(float64(g.Player.ViewTexture.Texture.Width), float64(g.Player.ViewTexture.Texture.Height))

			if g.RenderFlags&RENDER_FLAG_FULLBRIGHT != 0 {
				g.MainShader.FullBright.Set(1)
			} else {
				g.MainShader.FullBright.Set(0)
			}

			g.MainShader.Ambient.SetColor(color.RGBA{50, 50, 50, 255})

			if !g.IsStation {

				hour := math.Mod(g.Day, 1) * 24
				day := c(hour-HOUR_MORNING) - c(hour-HOUR_NIGHT)
				transitionColor := 1 + ((c(2*(hour-HOUR_MORNING-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_MORNING+HOURS_TRANSITION/2))) + (c(2*(hour-HOUR_NIGHT-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_NIGHT+HOURS_TRANSITION/2))))
				transitionAngle := 1 + ((c((hour - HOUR_MORNING - HOURS_TRANSITION/2)) - c((hour - HOUR_MORNING + HOURS_TRANSITION/2))) + (c((hour - HOUR_NIGHT - HOURS_TRANSITION/2)) - c((hour - HOUR_NIGHT + HOURS_TRANSITION/2))))

				sunColor := DAWN.Lerp(NIGHT.Lerp(DAY, (day)), (transitionColor))

				g.MainShader.LightDirectional(NewVec3((1-transitionAngle), (1-day*2), 0).Normalize(), rl.ColorFromHSV(float32(sunColor.X), float32(sunColor.Y), float32(sunColor.Z)), 1)

			} else {
				g.MainShader.LightDirectional(NewVec3(0, -1, 0).Normalize(), rl.White, 0.25)
			}

			g.MainShader.LightSpot(g.Player.Position3D().Add(Y.Scale(g.Player.Radius)), g.Player.LookPosition.Add(Y.Scale(g.Player.Radius)), 30, 40, rl.White, 1.5)

			g.MainShader.UpdateValues()

			g.Draw3D(int(g.Player.Y))
		})

	})

	rl.DrawTexturePro(
		g.MainTexture.Texture,
		rl.NewRectangle(0, 0, float32(g.MainTexture.Texture.Width), -float32(g.MainTexture.Texture.Height)),
		rl.NewRectangle(0, 0, float32(screenWidth), float32(screenHeight)),
		rl.Vector2{0, 0},
		0,
		rl.White,
	)

	rl.DrawTexturePro(
		g.Player.ViewTexture.Texture,
		rl.NewRectangle(0, 0, float32(g.Player.ViewTexture.Texture.Width), -float32(g.Player.ViewTexture.Texture.Height)),
		rl.NewRectangle(0, 0, float32(screenWidth)/4, float32(screenHeight)/4),
		rl.Vector2{0, 0},
		0,
		rl.White,
	)
}

var NIGHT = Vec3{-115, 0.3, .1}
var DAWN = Vec3{5, 0.5, 1}
var DAY = Vec3{55, 0.1, 1}
var HOUR_MORNING float64 = 9
var HOUR_NIGHT float64 = 21
var HOURS_TRANSITION float64 = 1

func c(x float64) float64 {
	return (1 + math.Tanh(x)) / 2
}

func (g *Game) Draw3D(maxY int) {
	g.Monster.Draw3D(g, maxY)
	g.Level.Draw(g, maxY)
	g.Player.Draw(g)
}

func LoadSaveFromFile(path string, save *GameSave) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(save); err != nil {
		return err
	}

	return nil
}

func (save GameSave) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(save); err != nil {
		return err
	}

	return nil
}
