package main

import (
	"fmt"
	"log"
	"math/cmplx"
	"sync"
	"time"

	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	width   int32 = 1600
	height  int32 = 1200
	maxIter int32 = 80
)

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"mandelbrot",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		width,
		height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		log.Panic(err)
	}
	defer window.Destroy()

	renderer, err := window.GetRenderer()
	if err != nil {
		log.Panic(err)
	}
	defer renderer.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		log.Panic(err)
	}

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		log.Panic(err)
	}
	defer texture.Destroy()

	// texture, err := renderer.CreateTexture(
	// 	sdl.PIXELFORMAT_RGB888,
	// 	sdl.TEXTUREACCESS_STATIC,
	// 	width,
	// 	height,
	// )
	// defer texture.Destroy()
	renderer.SetRenderTarget(texture)

	rect := &sdl.Rect{X: 0, Y: 0, W: width, H: height}

	aspect := &aspect{
		REStart: -2,
		REEnd:   1,
		IMStart: -1.5,
		IMEnd:   1.5,
	}

	drawMandelbrot(texture, rect, aspect)

	window.UpdateSurface()
	// window.UpdateSurface()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

type aspect struct {
	REStart float64
	REEnd   float64
	IMStart float64
	IMEnd   float64
}

type point struct {
	X int32
	Y int32
	C uint8
}

func (p *point) setC(c uint8) {
	p.C = c
}

func (p *point) getRGB() (r uint8, g uint8, b uint8) {
	r, g, b = p.C, p.C, p.C
	return
}

func (p *point) getColor() color.Color {
	r, g, b := p.getRGB()
	return color.RGBA{r, g, b, 0}
}

func drawMandelbrot(t *sdl.Texture, r *sdl.Rect, a *aspect) {

	var rows [][]point

	for y := int32(0); y < height; y++ {
		row := make([]point, width)
		for x := int32(0); x < width; x++ {
			row[x] = point{X: x, Y: y}
		}
		rows = append(rows, row)
	}

	ticCalc := time.Now()
	var wg sync.WaitGroup
	for _, row := range rows {
		wg.Add(1)
		go func(_row []point) {
			defer wg.Done()
			setMandelbrotColorSlice(_row, a)
		}(row)
	}
	wg.Wait()
	fmt.Printf("Calc time: %v\n", time.Since(ticCalc))

	ticDraw := time.Now()
	pixels := make([]byte, width*height*3)
	for _, row := range rows {
		for _, p := range row {
			r, g, b := p.getRGB()
			pixels[p.Y*width+p.X] = byte(r)
			pixels[p.Y*width+p.X+1] = byte(g)
			pixels[p.Y*width+p.X+2] = byte(b)
			// s.Set(int(p.X), int(p.Y), p.getColor())
		}
	}
	t.Update(r, pixels, 0)
	fmt.Printf("Draw time: %v\n", time.Since(ticDraw))
}

func setMandelbrotColorSlice(points []point, a *aspect) {
	for i, p := range points {
		points[i].setC(getMandelbrotColor(p.X, p.Y, a))
	}
}

func getMandelbrotColor(x int32, y int32, a *aspect) (c uint8) {
	real := a.REStart + (float64(x)/float64(width))*(a.REEnd-a.REStart)
	imag := a.IMStart + (float64(y)/float64(height))*(a.IMEnd-a.IMStart)
	m := mandelbrot(complex(real, imag))
	if m == maxIter {
		c = 0
	} else {
		c = uint8(255 * m / maxIter)
	}
	return
}

func mandelbrot(c complex128) (n int32) {
	var z complex128 = 0.
	for cmplx.Abs(z) <= 2 && n < maxIter {
		z = z*z + c
		n++
	}
	return
}
