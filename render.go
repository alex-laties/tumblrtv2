package main

import "C"

import (
	"fmt"
	"image"
	"image/gif"
	"io/ioutil"
	"log"
	"os/user"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var (
	currentGIF                                             *gif.GIF
	perGIFFrames                                           uint64 = 303 // 10 seconds per gif
	globalFrameCount, currentFrame, currentCount, maxCount uint64
	currentFramesOnGIF                                     uint64
	program                                                uint32

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
	imageVAO        uint32
	maddenTextures  []*Texture
	currentTextures []*Texture
)

func goglInit() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	tagBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/.tumblrtvconfig", usr.HomeDir))
	var tags []string
	if err != nil {
		tags = []string{"sakuga", "cyberpunk", "trippy", "tumblr", "staff"}
	} else {
		tags = strings.Split(string(tagBytes), ",")
	}

	go fetchGIFs(tags...)

	err = gl.Init()
	if err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println(version)
	program = gl.CreateProgram()
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

	initGIF(<-gifPipeline)
}

func initGIF(g *gif.GIF) {
	// reset frame counters
	currentFramesOnGIF = 0
	currentCount = 0
	currentFrame = 0

	currentTextures = nil
	var lastFrame *image.Paletted
	for _, gFrame := range g.Image {
		if lastFrame != nil {
			for x := lastFrame.Rect.Min.X; x < lastFrame.Rect.Max.X; x++ {
				for y := lastFrame.Rect.Min.Y; y < lastFrame.Rect.Max.Y; y++ {
					gFrameOffset := gFrame.PixOffset(x, y)
					if 0 > gFrameOffset || gFrameOffset > len(gFrame.Pix)-1 {
						continue
					}
					if gFrame.Pix[gFrameOffset] == 255 {
						gFrame.Pix[gFrameOffset] = lastFrame.Pix[lastFrame.PixOffset(x, y)]
					}
				}
			}
		}
		tex, err := NewTexture(gFrame, gl.CLAMP_TO_BORDER, gl.CLAMP_TO_EDGE)
		if err != nil {
			panic(err)
		}
		currentTextures = append(currentTextures, tex)
		lastFrame = gFrame
	}

	currentGIF = g
}

// GoGLRender is the function called into from Objective-C... Gross
//export GoGLRender
func GoGLRender() {
	if program == 0 {
		goglInit() // have to init in initial render loop because of not having the opengl context until here
	}

	if currentFramesOnGIF > perGIFFrames {
		initGIF(<-gifPipeline)
	} else {
		currentFramesOnGIF++
	}

	//gl.ClearColor(0, 0, 0, 1.0)
	//gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
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

func main() {}
