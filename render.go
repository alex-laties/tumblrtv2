package main

import "C"

import (
	"bytes"
	"fmt"
	"image/gif"
	"log"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/gobuffalo/packr"
)

var (
	currentGIF                                             *gif.GIF
	perGIFFrames                                           uint64 = 303 // 10 seconds per gif
	globalFrameCount, currentFrame, currentCount, maxCount uint64
	currentFramesOnGIF                                     uint64
	program                                                uint32

	shaderBox = packr.NewBox("./assets/shaders")
	gifBox    = packr.NewBox("./assets/gifs")

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
	triangleVAO     uint32
	imageVAO        uint32
	maddenTextures  []*Texture
	currentTextures []*Texture
)

func goglInit() {
	go fetchGIFs("cat")
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println(version)
	program = gl.CreateProgram()
	triangleVAO = makeVAO(triangle)
	imageVAO = makeImageVAO(image_vertices, image_indices)
	vertexShaderSource, _ := shaderBox.FindString("basic.vert")
	vertexShader, err := compileShader(vertexShaderSource+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShaderSource, _ := shaderBox.FindString("basic.frag")
	fragmentShader, err := compileShader(fragmentShaderSource+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	maddenGIFBytes := gifBox.Bytes("madden.gif")
	maddenReader := bytes.NewBuffer(maddenGIFBytes)
	maddenGIF, err := gif.DecodeAll(maddenReader)
	if err != nil {
		panic(err)
	}

	currentGIF = maddenGIF
	for _, image := range maddenGIF.Image {
		tex, err := NewTexture(image, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
		if err != nil {
			panic(err)
		}
		maddenTextures = append(maddenTextures, tex)
	}

	currentTextures = maddenTextures
}

func initGIF(g *gif.GIF) {
	currentTextures = nil

	for _, image := range g.Image {
		tex, err := NewTexture(image, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
		if err != nil {
			panic(err)
		}
		currentTextures = append(currentTextures, tex)
	}

	currentGIF = g
}

//export GoGLRender
func GoGLRender() {
	if program == 0 {
		goglInit() // have to init in initial render loop because of not having the opengl context until here
	}

	if currentFramesOnGIF > perGIFFrames {
		currentFramesOnGIF = 0
		initGIF(<-gifPipeline)
	} else {
		currentFramesOnGIF++
	}

	gl.ClearColor(0, 0, 0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)
	maxCount = uint64(currentGIF.Delay[currentFrame])
	if currentCount < maxCount {
		currentCount += 3
	} else {
		currentFrame = (currentFrame + 1) % uint64(len(currentGIF.Image))
		currentCount = 0
		maxCount = uint64(currentGIF.Delay[currentFrame])
	}

	textureToUse := currentTextures[currentFrame]

	textureToUse.Bind(gl.TEXTURE0)
	textureToUse.SetUniform(gl.GetUniformLocation(program, gl.Str("wat\x00")))
	gl.BindVertexArray(imageVAO)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
	gl.BindVertexArray(0)
	textureToUse.UnBind()
	globalFrameCount++
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
