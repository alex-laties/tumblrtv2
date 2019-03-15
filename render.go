package main

import "C"

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var (
	program uint32

	vertexShaderSource = `
    #version 410
    in vec3 vp;
    void main() {
        gl_Position = vec4(vp, 1.0);
    }
` + "\x00"

	fragmentShaderSource = `
    #version 410
    out vec4 frag_colour;
    void main() {
        frag_colour = vec4(0.5, 1, 1, 1);
    }
` + "\x00"

	triangle = []float32{
		1, -1, 1,
		0, 1, 0,
		-1, -1, 0,
	}
	triangleVAO uint32
)

//export GoGLRender
func GoGLRender() {
	if program == 0 {
		err := gl.Init()
		if err != nil {
			panic(err)
		}
		version := gl.GoStr(gl.GetString(gl.VERSION))
		log.Println(version)
		program = gl.CreateProgram()
		triangleVAO = makeVAO(triangle)
		vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
		if err != nil {
			panic(err)
		}
		fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
		if err != nil {
			panic(err)
		}

		gl.AttachShader(program, vertexShader)
		gl.AttachShader(program, fragmentShader)
		gl.LinkProgram(program)
	}

	gl.ClearColor(0.5, 0, 0.5, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindVertexArray(triangleVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangle)/3))
}

// makeVAO initializes and returns a vertex array from the points provided.
func makeVAO(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func main() {}
