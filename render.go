package main

import "C"

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var (
	program uint32

	vertexShaderSource = `
		#version 410 core

		layout (location = 0) in vec3 position;
		layout (location = 1) in vec3 color;
		layout (location = 2) in vec2 texCoord;

		out vec3 ourColor;
		out vec2 TexCoord;

		void main()
		{
				gl_Position = vec4(position, 1.0);
				TexCoord = texCoord;    // pass the texture coords on to the fragment shader
		}
` + "\x00"

	fragmentShaderSource = `
    #version 410 core

		in vec3 ourColor;
		in vec2 TexCoord;

		out vec4 color;

		uniform sampler2D wat;

		void main()
		{
				color = texture(wat, TexCoord);
		}
` + "\x00"

	triangle = []float32{
		1, -1, 1,
		0, 1, 0,
		-1, -1, 0,
	}
	image_vertices = []float32{
		// top left
		-1, 1, 0, // position
		0, 0, 0, // Color
		0, 0, // texture coordinates

		// top right
		1, 1, 0.0,
		0.0, 0.0, 0.0,
		1, 0,

		// bottom right
		1, -1, 0.0,
		0.0, 0.0, 0.0,
		1, 1,

		// bottom left
		-1, -1, 0,
		0, 0, 0,
		0, 1,
	}

	image_indices = []uint32{
		// rectangle
		0, 1, 2, // top triangle
		0, 2, 3, // bottom triangle
	}
	triangleVAO   uint32
	imageVAO      uint32
	maddenTexture *Texture
)

func goglInit() {
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println(version)
	program = gl.CreateProgram()
	triangleVAO = makeVAO(triangle)
	imageVAO = makeImageVAO(image_vertices, image_indices)
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

	maddenTexture, err = NewTextureFromFile("/tmp/madden.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err)
	}
}

//export GoGLRender
func GoGLRender() {
	if program == 0 {
		goglInit() // have to init in initial render loop because of not having the opengl context until here
	}

	gl.ClearColor(0, 0, 0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	maddenTexture.Bind(gl.TEXTURE0)
	maddenTexture.SetUniform(gl.GetUniformLocation(program, gl.Str("wat\x00")))
	gl.BindVertexArray(imageVAO)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
	gl.BindVertexArray(0)
	maddenTexture.UnBind()
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

func makeImageVAO(vertices []float32, indices []uint32) uint32 {
	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// copy indices into element buffer
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 3*4 + 2*4
	var offset int

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// color
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 3 * 4

	// texture position
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(2)
	offset += 2 * 4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
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
