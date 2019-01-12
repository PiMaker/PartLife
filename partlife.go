package main

import (
	"fmt"
	"github.com/PiMaker/PartLife/sim"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"time"
)

const (
	width  = 1920
	height = 1080
)

var (
	frames = 0
	steps  = 0
	second = time.Tick(time.Second)
)

func main() {
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "PartLife | FPS: 0 | UPS: 0",
		Bounds: pixel.R(0, 0, width, height),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Black)

	sim := sim.Init(width, height)
	imd := imdraw.New(nil)

	go func() {
		for !win.Closed() {
			sim.Step()
			steps++
		}
	}()

	for !win.Closed() {
		update(win, imd, sim)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("PartLife | FPS: %d | UPS: %d", frames, steps))
			frames = 0
			steps = 0
		default:
		}
	}
}

func update(win *pixelgl.Window, imd *imdraw.IMDraw, simulation *sim.Sim) {
	//win.Clear(pixel.RGB(0, 0, 0).Mul(pixel.Alpha(0.2)))
	imd.Clear()
	imd.Color = pixel.RGB(0, 0, 0).Mul(pixel.Alpha(0.075))
	imd.Push(pixel.V(0, 0), pixel.V(width, height))
	imd.Rectangle(0)
	imd.Draw(win)

	imd.Clear()

	for _, p := range simulation.Parts {
		imd.Color = p.Color
		imd.Push(p.Position)
		imd.Circle(sim.Radius, 0)
	}

	imd.Draw(win)
}
