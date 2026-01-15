package game

import (
	"log"
	"reflect"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Uniform struct {
	shader   rl.Shader
	location int32
}

func NewUniform(shader rl.Shader, name string) Uniform {
	// panic if -1 location

	location := rl.GetShaderLocation(shader, name)

	if location == -1 {
		log.Fatalf("invalid uniform \"%v\"", name)
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
func (u *UniformVec4) Set(x, y, z, w float64) {
	rl.SetShaderValue(u.shader, u.location, []float32{float32(x), float32(y), float32(z), float32(w)}, rl.ShaderUniformVec4)
}
func (u *UniformTexture) Set(texture rl.Texture2D) {
	rl.SetShaderValueTexture(u.shader, u.location, texture)
}

type Shader[T any] struct {
	shader  rl.Shader
	Uniform T
}

func (s *Shader[T]) Unload() {
	rl.UnloadShader(s.shader)
}

func NewShader[T any](vs string, fs string) *Shader[T] {

	s := &Shader[T]{
		shader: rl.LoadShader(vs, fs),
	}

	locValue := reflect.ValueOf(&s.Uniform).Elem()
	locType := locValue.Type()

	for i := 0; i < locType.NumField(); i++ {
		field := locType.Field(i)

		uniformName, ok := field.Tag.Lookup("glsl")

		if !ok {
			continue
		}

		fieldValue := locValue.Field(i)

		embeddedUniform := fieldValue.FieldByName("Uniform")
		embeddedUniform.Set(reflect.ValueOf(NewUniform(s.shader, uniformName)))
	}

	return s
}

func (s *Shader[T]) UseMode(fn func()) {
	rl.BeginShaderMode(s.shader)
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

func BeginMode3D(camera rl.Camera3D, fn func()) {
	rl.BeginMode3D(camera)
	fn()
	rl.EndMode3D()
}

func BeginMode2D(camera rl.Camera2D, fn func()) {
	rl.BeginMode2D(camera)
	fn()
	rl.EndMode2D()
}
