package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/sqweek/dialog"

	// profiling

	"net/http"
	_ "net/http/pprof"
)

func init() {
	go http.ListenAndServe("localhost:8080", nil)
}

const (
	screenWidth  = int(2 << 9 * 1.5)
	screenHeight = 2 << 9
)

var MAX_ITERS = 1 << 10 // max iterations before determining whther the point is in set or not

var WORKERS = 1

func init() {
	WORKERS = runtime.NumCPU()/2 - 1
	if WORKERS < 1 {
		WORKERS = 2
	}

	if w := os.Getenv("WORKERS"); w != "" {
		if ww, err := strconv.ParseUint(w, 10, 32); err == nil {
			WORKERS = int(ww)
		}
	}
}

type Game struct {
	SetComputor
	imageComputor SetComputor
	pixels        []byte
	ticks         int
}

// input handling
func (g *Game) Update() error {
	mid := g.set.mid
	zoom := g.set.zoom

	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.set.steps = 1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		mid = -0.5
		zoom = 1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		zoom = nextSqrt(zoom, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		zoom = nextSqrt(zoom, +1)
	}
	if _, yoff := ebiten.Wheel(); yoff != 0 {
		xx, yy := ebiten.CursorPosition()
		mid = g.set.PixToSet(xx, yy)

		if yoff > 0 {
			zoom = nextSqrt(zoom, +1)
		} else {
			zoom = nextSqrt(zoom, -1)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		xx, yy := ebiten.CursorPosition()
		mid = g.set.PixToSet(xx, yy)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		mid -= complex(0, 0.5/zoom)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		mid += complex(0, 0.5/zoom)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		mid -= complex(0.5/zoom, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		mid += complex(0.5/zoom, 0)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.set.coloring = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.set.coloring = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.set.coloring = 2
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.set.coloring = 3
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) &&
		ebiten.IsKeyPressed(ebiten.KeyControl) {
		filename, err := dialog.File().
			SetStartDir("images").
			Filter("Mandelbrot png file", "png").
			Load()
		if err != nil {
			fmt.Println(err)
		} else {
			go func() {
				g.set.GoByFilename(filename)
				g.Compute()
			}()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) &&
		ebiten.IsKeyPressed(ebiten.KeyControl) {
		if g.imageComputor.progress != PCS*PCS {
			ok := dialog.
				Message("Saving in progress\ninterrupt?").
				YesNo()
			if !ok {
				return nil
			}
		}
		fileName, err := dialog.File().
			SetStartDir("images").
			SetStartFile(g.set.ToFileName()).
			Title("Save image").Save()
		if err != nil {
			fmt.Println(err)
		} else {
			go g.SaveImage(fileName)
		}
	}

	// boundaries
	if zoom < 0.25 {
		zoom = 0.25
	}
	if size(mid+1) >= 6 {
		mid = g.set.mid
	}

	if g.set.zoom != zoom || g.set.mid != mid {
		g.Transform(zoom, mid)

		go g.Compute()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.set.steps = MAX_ITERS
		go g.Compute()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		g.set.steps /= 2
		go g.Compute()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.set.steps *= 2
		go g.Compute()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.set.steps = MAX_ITERS << 5
		go g.Compute()
	}

	// if g.set.steps < MAX_ITERS {
	// 	if g.ticks%6 == 0 {
	// 		go g.Compute()
	// 		g.set.steps += 1 + int(math.Pow((math.Log(float64(g.set.steps))), 3)/10)
	// 		if g.set.steps > MAX_ITERS {
	// 			g.set.steps -= g.set.steps % MAX_ITERS
	// 		}
	// 	}
	// 	g.ticks++
	// }

	return nil
}

func (g *Game) SaveImage(fileName string) {
	w := g.set.w * 2
	h := g.set.h * 2
	if g.imageComputor.set == nil {
		g.imageComputor.set = NewSet(w, h)
	} else {
		g.imageComputor.set.w = w
		g.imageComputor.set.h = h
		g.imageComputor.set.grid = make([]float64, w*h)
	}

	set := g.imageComputor.set
	set.mid = g.set.mid
	set.zoom = g.set.zoom
	set.steps = g.set.steps
	set.coloring = g.set.coloring

	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})

	g.imageComputor.Compute()
	set.Draw(img.Pix)

	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	e := png.Encode(f, img)
	if e != nil {
		fmt.Println(err)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	view := screen.Bounds()
	vx, vy := view.Dx(), view.Dy()
	if g.pixels == nil || len(g.pixels)/4 != vx*vy { // adjust pixels to match screen
		g.pixels = make([]byte, vx*vy*4)
	}
	if vx != g.set.w || vy != g.set.h {
		println("Mismatched set and pixels size")
		return
	}
	g.ticks++
	if g.ticks%2 == 0 {
		// render the set
		g.set.Draw(g.pixels)
		// g.DrawNeigh(g.pixels) // minimap
	}
	screen.WritePixels(g.pixels)

	// render hud texts
	x, y := ebiten.CursorPosition()
	mid := g.set.mid
	if x >= 0 && x < g.set.w && y >= 0 && y < g.set.h { // not oustise of screen
		mid = g.set.PixToSet(x, y)
	}
	progress := fmt.Sprintf("%v/%v", g.progress, PCS*PCS)
	imgProgress := fmt.Sprintf("%v/%v", g.imageComputor.progress, PCS*PCS)
	debugText := fmt.Sprintf(
		// "Arrows to navigate\n"+
		// "PgUp/PgDn to zoom\n"+
		"c: %v\nzoom: %v\niters: %v\nchunks: %v\n",
		mid,
		toHumNum(g.set.zoom),
		toHumNum(float64(g.set.steps)),
		progress,
	)
	if g.imageComputor.progress < PCS*PCS {
		debugText += fmt.Sprintf("saving: %v\n", imgProgress)
	}
	ebitenutil.DebugPrint(screen, debugText)
}

func (g *Game) Layout(outW, outH int) (int, int) {
	if g.set.w != outW || g.set.h != outH { // adjust on resize
		if g.cancelUpdate != nil {
			(*g.cancelUpdate)()
		}
		g.set.w = outW
		g.set.h = outH
		g.set.grid = make([]float64, outW*outH)
		go g.Compute()
	}
	return outW, outH
}

func main() {
	g := &Game{}
	g.set = NewSet(screenWidth, screenHeight)
	g.imageComputor.progress = PCS * PCS
	if len(os.Args) > 1 { // lodad set position from filename
		g.set.GoByFilename(os.Args[1])
	}
	go g.Compute()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Mandelbrot set - Draho")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
