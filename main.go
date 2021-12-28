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

	surface, err := window.GetSurface()
	if err != nil {
		log.Panic(err)
	}

	aspect := &aspect{
		REStart: -2,
		REEnd:   1,
		IMStart: -1,
		IMEnd:   1,
	}

	drawMandelbrot(surface, aspect)
	window.UpdateSurface()

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

func (p *point) getColor() color.Color {
	return color.RGBA{p.C, p.C, p.C, 0}
}

func drawMandelbrot(s *sdl.Surface, a *aspect) {
	s.Lock()
	defer s.Unlock()
	s.FillRect(nil, 0)

	tic := time.Now()

	var rows [][]point

	for y := int32(0); y < s.H; y++ {
		row := make([]point, s.W)
		for x := int32(0); x < s.W; x++ {
			row[x] = point{X: x, Y: y}
		}
		rows = append(rows, row)
	}

	var wg sync.WaitGroup

	for _, row := range rows {
		wg.Add(1)
		go func(_row []point) {
			defer wg.Done()
			setMandelbrotColorSlice(_row, a, s)
		}(row)
	}

	wg.Wait()

	for _, row := range rows {
		for _, p := range row {
			s.Set(int(p.X), int(p.Y), p.getColor())
		}
	}

	t := time.Since(tic)
	fmt.Printf("Draw time: %v\n", t)
}

func setMandelbrotColorSlice(points []point, a *aspect, s *sdl.Surface) {
	for i, p := range points {
		points[i].setC(getMandelbrotColor(p.X, p.Y, a, s))
	}
}

func getMandelbrotColor(x int32, y int32, a *aspect, s *sdl.Surface) (c uint8) {
	real := a.REStart + (float64(x)/float64(s.W))*(a.REEnd-a.REStart)
	imag := a.IMStart + (float64(y)/float64(s.H))*(a.IMEnd-a.IMStart)
	m := mandelbrot(complex(real, imag))
	c = uint8(255 * m / maxIter)
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
