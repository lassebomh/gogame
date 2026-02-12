package game3

import rl "github.com/gen2brain/raylib-go/raylib"

var globals struct {
	Models struct {
		MonsterBody       rl.Model
		MonsterArmSegment rl.Model
	}

	Textures struct {
		PlanetElevation rl.Texture2D
		Organic         rl.Texture2D
		Atlas           rl.Texture2D
		AtlasTiles      int
	}

	Shaders struct {
		Planet *PlanetShader
		Main   *MainShader
	}
}

func globalsInit() {

	globals.Shaders.Planet = NewShader(&PlanetShader{}, "", "./resources/glsl330/planet2.fs")
	globals.Shaders.Main = NewShader(&MainShader{}, "./resources/glsl330/lighting.vs", "./resources/glsl330/lighting.fs")

	globals.Textures.Organic = rl.LoadTexture("./resources/organic.png")
	globals.Textures.PlanetElevation = rl.LoadTexture("./resources/earth_elevation.png")
	globals.Textures.Atlas = rl.LoadTexture("./resources/atlas.png")
	globals.Textures.AtlasTiles = 15

	globals.Models.MonsterBody = rl.LoadModel("./resources/models/monster/monster_body.glb")
	mats := globals.Models.MonsterBody.GetMaterials()
	for i := range mats {
		mats[i].Shader = globals.Shaders.Main.shader
		mats[i].GetMap(rl.MapDiffuse).Texture = globals.Textures.Organic
	}
	globals.Models.MonsterArmSegment = rl.LoadModel("./resources/models/monster/monster_arm_segment.glb")
	mats = globals.Models.MonsterArmSegment.GetMaterials()
	for i := range mats {
		mats[i].Shader = globals.Shaders.Main.shader
		mats[i].GetMap(rl.MapDiffuse).Texture = globals.Textures.Organic
	}

	rl.SetLineWidth(2)

}
