package game3

import (
	"fmt"
	v2 "game/vec2"
	v3 "game/vec3"
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
func (u *UniformVec3) Set(x, y, z float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(x), float32(y), float32(z)}, rl.ShaderUniformVec3)
}
func (u *UniformVec3) SetVec3(v v3.Value) {
	u.Set(v.X, v.Y, v.Z)
}
func (u *UniformVec4) Set(x, y, z, w float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(x), float32(y), float32(z), float32(w)}, rl.ShaderUniformVec4)
}
func (u *UniformVec4) SetColor(color color.RGBA) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(color.R) / 255, float32(color.G) / 255, float32(color.B) / 255, float32(color.A) / 255}, rl.ShaderUniformVec4)
}
func (u *UniformTexture) Set(texture rl.Texture2D) {
	rl.SetShaderValueTexture(u.shader, u.location, texture)
	// rl.EnableTexture(texture.ID)
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
	Position v3.Value
	// Camera target it looks-at
	Target v3.Value
	// Camera up vector (rotation over its axis)
	Up v3.Value
	// Camera field-of-view apperture in Y (degrees) in perspective, used as near plane width in orthographic
	Fovy float64
	// Camera type, controlling projection type, either CameraPerspective or CameraOrthographic.
	Projection rl.CameraProjection
}

func (c Camera3D) Raylib() rl.Camera3D {
	return rl.Camera3D{
		Position:   c.Position.Raylib(),
		Target:     c.Target.Raylib(),
		Up:         v3.Y(1).Raylib(),
		Fovy:       float32(c.Fovy),
		Projection: c.Projection,
	}
}

type CursorLayout struct{ rl.Rectangle }

func NewCursorLayout(x, y, w, h float32) *CursorLayout {
	return &CursorLayout{
		Rectangle: rl.NewRectangle(x, y, w, h),
	}
}

func (c *CursorLayout) Right() *CursorLayout {
	c.X += c.Width
	return c
}
func (c *CursorLayout) Down() *CursorLayout {
	c.Y += c.Height
	return c
}
func (c *CursorLayout) Copy() *CursorLayout {
	return NewCursorLayout(c.X, c.Y, c.Width, c.Height)
}
func (c *CursorLayout) With(fn func(cursor *CursorLayout)) {
	fn(NewCursorLayout(c.X, c.Y, c.Width, c.Height))
}

type StackLayout struct {
	pos  v2.Value
	size v2.Value
}

func NewStackLayout(x, y, w, h float64) *StackLayout {

	return &StackLayout{
		pos:  v2.XY(x, y),
		size: v2.XY(w, h),
	}
}

func (s *StackLayout) ToRectangle() rl.Rectangle {
	return rl.NewRectangle(
		float32(s.pos.X),
		float32(s.pos.Y),
		float32(s.size.X),
		float32(s.size.Y),
	)
}

func (s *StackLayout) Down(h float64) rl.Rectangle {
	rect := rl.NewRectangle(
		float32(s.pos.X),
		float32(s.pos.Y),
		float32(s.size.X),
		float32(h),
	)
	s.pos.Y += h
	return rect
}
func (s *StackLayout) Right(w float64) rl.Rectangle {
	rect := rl.NewRectangle(
		float32(s.pos.X),
		float32(s.pos.Y),
		float32(w),
		float32(s.size.Y),
	)
	s.pos.X += w
	return rect
}

type LineLayout struct {
	X      float64
	Y      float64
	Height float64
	Width  float64
}

func NewLineLayout(X float64, Y float64, Height float64, Width float64) *LineLayout {
	return &LineLayout{X, Y, Height, Width}
}

func (l *LineLayout) Next() rl.Rectangle {
	rect := rl.NewRectangle(
		float32(l.X+l.Width),
		float32(l.Y),
		float32(l.Width),
		float32(l.Height),
	)

	return rect
}

func (l *LineLayout) NextEx(width float64) rl.Rectangle {
	rect := rl.NewRectangle(
		float32(l.X+l.Width),
		float32(l.Y),
		float32(width),
		float32(l.Height),
	)

	l.Width += width

	return rect
}

func (l *LineLayout) Break() {
	l.Width = 0
	l.Y += l.Height
}
func (l *LineLayout) BreakEx(height float64) {
	l.Width = 0
	l.Y += l.Height
	l.Height = height
}

func ScreenToWorld(camera Camera3D, screen v2.Value, y float64) v3.Value {
	ray := rl.GetScreenToWorldRay(screen.Raylib(), camera.Raylib())
	origin := v3.FromRaylib(ray.Position)
	direction := v3.FromRaylib(ray.Direction)
	hitpos := origin.Add(direction.Scale((y - origin.Y) / direction.Y))
	hitpos.Y = y
	return hitpos
}
