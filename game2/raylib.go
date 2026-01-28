package game2

import (
	"fmt"
	"image/color"
	"log"
	"reflect"
	"strings"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Uniform struct {
	shader   rl.Shader
	location int32
}

func NewUniform(shader rl.Shader, name string) Uniform {
	location := rl.GetShaderLocation(shader, name)

	if location == -1 {
		log.Printf("WARNING! invalid uniform \"%v\"", name)
	}

	return Uniform{
		shader:   shader,
		location: location,
	}
}

type UniformFloat struct{ Uniform }
type UniformInt struct{ Uniform }
type UniformTexture struct{ Uniform }
type UniformVec2 struct{ Uniform }
type UniformVec3 struct{ Uniform }
type UniformVec4 struct{ Uniform }
type UniformMat4 struct{ Uniform }

func (u *UniformInt) Set(value int32) {
	rl.SetShaderValue(u.shader, u.location, unsafe.Slice((*float32)(unsafe.Pointer(&value)), 4), rl.ShaderUniformInt)
}
func (u *UniformFloat) Set(value float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(value)}, rl.ShaderUniformFloat)
}
func (u *UniformVec2) Set(x, y float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(x), float32(y)}, rl.ShaderUniformVec2)
}
func (u *UniformVec2) SetVec2(vec Vec2) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(vec.X), float32(vec.Y)}, rl.ShaderUniformVec2)
}
func (u *UniformVec3) Set(x, y, z float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(x), float32(y), float32(z)}, rl.ShaderUniformVec3)
}
func (u *UniformVec3) SetVec3(vec Vec3) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(vec.X), float32(vec.Y), float32(vec.Z)}, rl.ShaderUniformVec3)
}
func (u *UniformVec4) Set(x, y, z, w float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(x), float32(y), float32(z), float32(w)}, rl.ShaderUniformVec4)
}
func (u *UniformVec4) SetVec4(vec Vec4) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(vec.X), float32(vec.Y), float32(vec.Z), float32(vec.W)}, rl.ShaderUniformVec4)
}
func (u *UniformVec4) SetColor(color color.RGBA) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(color.R) / 255, float32(color.G) / 255, float32(color.B) / 255, float32(color.A) / 255}, rl.ShaderUniformVec4)
}
func (u *UniformTexture) Set(texture rl.Texture2D) {
	rl.SetShaderValueTexture(u.shader, u.location, texture)
}
func (u *UniformMat4) Set(mat rl.Matrix) {
	rl.SetShaderValueMatrix(u.shader, u.location, mat)
}

type Shader interface {
	GetRaylibShader() rl.Shader
	SetRaylibShader(value rl.Shader)
}

func NewShader[T Shader](shader T, vs string, fs string) T {
	raylibShader := rl.LoadShader(vs, fs)
	shader.SetRaylibShader(raylibShader)

	locValue := reflect.ValueOf(shader).Elem()
	locType := locValue.Type()

	for i := 0; i < locType.NumField(); i++ {
		field := locType.Field(i)

		uniformName, ok := field.Tag.Lookup("glsl")

		if !ok {
			continue
		}

		fieldValue := locValue.Field(i)

		if strings.Contains(uniformName, "%d") {

			for i := range fieldValue.Len() {

				uniformName := fmt.Sprintf(uniformName, i)
				embeddedUniform := fieldValue.Index(i).FieldByName("Uniform")
				embeddedUniform.Set(reflect.ValueOf(NewUniform(raylibShader, uniformName)))
			}
		} else {
			embeddedUniform := fieldValue.FieldByName("Uniform")
			embeddedUniform.Set(reflect.ValueOf(NewUniform(raylibShader, uniformName)))

		}

	}

	return shader
}

func BeginShaderMode(shader Shader, fn func()) {
	rl.BeginShaderMode(shader.GetRaylibShader())
	fn()
	rl.EndShaderMode()
}

func BeginDrawing(fn func()) {
	rl.BeginDrawing()
	fn()
	rl.EndDrawing()
}

func BeginTextureMode(texture rl.RenderTexture2D, fn func()) {
	rl.BeginTextureMode(texture)
	fn()
	rl.EndTextureMode()
}

func BeginMode3D(camera Camera3D, fn func()) {
	rl.BeginMode3D(camera.Raylib())
	fn()
	rl.EndMode3D()
}

func BeginMode2D(camera rl.Camera2D, fn func()) {
	rl.BeginMode2D(camera)
	fn()
	rl.EndMode2D()
}

func BeginOverlayMode(fn func()) {
	rl.DrawRenderBatchActive()
	rl.DisableDepthTest()

	fn()
	rl.DrawRenderBatchActive()
	rl.EnableDepthTest()

}

// Camera3D type, defines a camera position/orientation in 3d space
type Camera3D struct {
	// Camera position
	Position Vec3
	// Camera target it looks-at
	Target Vec3
	// Camera up vector (rotation over its axis)
	Up Vec3
	// Camera field-of-view apperture in Y (degrees) in perspective, used as near plane width in orthographic
	Fovy float64
	// Camera type, controlling projection type, either CameraPerspective or CameraOrthographic.
	Projection rl.CameraProjection
}

func (c Camera3D) Raylib() rl.Camera3D {
	return rl.Camera3D{
		Position:   c.Position.Raylib(),
		Target:     c.Target.Raylib(),
		Up:         Y.Raylib(),
		Fovy:       float32(c.Fovy),
		Projection: c.Projection,
	}
}

type LineLayout struct {
	X      float64
	Y      float64
	Height float64
	Width  float64
}

func NewLineLayout(X float64, Y float64, Height float64) *LineLayout {
	return &LineLayout{X, Y, Height, 0}
}

func (l *LineLayout) Next(width float64) rl.Rectangle {
	rect := rl.NewRectangle(
		float32(l.X+l.Width),
		float32(l.Y),
		float32(width),
		float32(l.Height),
	)

	l.Width += width

	return rect
}

func (l *LineLayout) Break(height float64) {
	l.Width = 0
	l.Y += l.Height
	l.Height = height
}

func ScreenToWorld(camera Camera3D, screen Vec2, y float64) Vec3 {
	ray := rl.GetScreenToWorldRay(screen.Raylib(), camera.Raylib())
	origin := Vec3FromRaylib(ray.Position)
	direction := Vec3FromRaylib(ray.Direction)
	hitpos := origin.Add(direction.Scale((y - origin.Y) / direction.Y))
	hitpos.Y = y
	return hitpos
}
