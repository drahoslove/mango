package main

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const PCS = 32         // number of pieces per line/column
const NEIGHT_RES = 128 // resolution of neighborhood thubmnail

type Set struct {
	w, h     int       // dimensions of the grid
	grid     []float64 // int is the amount of iterations after which the value escapes the two-circle
	steps    int
	zoom     float64    // current zoom level
	mid      complex128 // current center of view
	coloring int
}

func NewSet(w, h int) *Set {
	s := Set{
		w:     w,
		h:     h,
		grid:  make([]float64, w*h),
		steps: MAX_ITERS,
		mid:   -0.5,
		zoom:  1,
	}
	return &s
}

func (s *Set) ToFileName() string {
	return fmt.Sprintf("set_%v_%v_%v_%v_.png",
		time.Now().Unix(), s.mid, s.zoom, s.coloring)
}

func (s *Set) GoByFilename(filename string) {
	filename = filepath.Base(filename)

	parts := strings.Split(filename, "_")
	if len(parts) < 5 {
		return
	}
	if parts[0] != "set" {
		return
	}

	mid, err := strconv.ParseComplex(parts[2], 64)
	if err != nil {
		return
	}
	zoom, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return
	}
	coloring, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return
	}
	s.mid = mid
	s.zoom = zoom
	s.coloring = int(coloring)

}

// converts pix from the grid to  ±2 + ±2i complex number
func (s *Set) PixToSet(x, y int) complex128 {
	zoom := s.zoom / 2.5
	span := s.h
	return complex(
		(float64(x-(s.w-s.h)/2-span/2)/float64(span))/zoom+real(s.mid),
		(float64(y-span/2)/float64(span))/zoom+imag(s.mid),
	)
}

func (s *Set) SetToPix(c complex128) (x, y int) {
	zoom := s.zoom / 2.5
	span := s.h
	x = int(
		(real(c)-real(s.mid))*zoom*float64(span) + float64(span/2) + float64(s.w-s.h)/2,
	)
	y = int(
		(imag(c)-imag(s.mid))*zoom*float64(span) + float64(span/2),
	)
	return
}

func (sc *SetComputor) Transform(newZoom float64, newMid complex128) {
	s := sc.set
	destSet := *s // copy the set
	destSet.grid = make([]float64, len(s.grid))
	destSet.mid = newMid
	destSet.zoom = newZoom
	for i := range destSet.grid {
		destX := i % s.w
		dextY := i / s.w
		c := destSet.PixToSet(destX, dextY)
		srcX, srcY := s.SetToPix(c)
		thX, thY := sc.NeighToPix(c, s)
		if srcX >= 0 && srcX < s.w && srcY >= 0 && srcY < s.h {
			j := srcY*s.w + srcX
			destSet.grid[i] = s.grid[j]
		} else { // use neighborhood
			j := thY*NEIGHT_RES + thX
			destSet.grid[i] = sc.neighborhood[j]
		}
	}

	// transform the current thumbnail
	destNeighborhood := [NEIGHT_RES * NEIGHT_RES]float64{}
	for i := range destNeighborhood { // copy the previous thumb first
		destNeighborhood[i] = sc.neighborhood[i]
	}
	for i := range destNeighborhood {
		destX := i % NEIGHT_RES
		destY := i / NEIGHT_RES
		c := sc.PixToNeigh(destX, destY, &destSet)
		srcX, srcY := sc.NeighToPix(c, s)
		if srcX >= 0 && srcX < NEIGHT_RES && srcY >= 0 && srcY < NEIGHT_RES {
			j := srcY*NEIGHT_RES + srcX
			destNeighborhood[i] = sc.neighborhood[j]
		}
	}
	*s = destSet

	sc.neighborhood = destNeighborhood

	// generate new neighborhood thumbnail to be used to prefill space when Transforming next time
	go func() {
		for x := 0; x < NEIGHT_RES; x++ {
			for y := 0; y < NEIGHT_RES; y++ {
				ok, st := isInSet(sc.PixToNeigh(x, y, s), sc.set.steps)
				if ok {
					sc.neighborhood[(y*NEIGHT_RES+x)%len(sc.neighborhood)] = 0
				} else {
					sc.neighborhood[(y*NEIGHT_RES+x)%len(sc.neighborhood)] = st
				}
			}
		}
	}()
}

// Draw paints current game state.
func (s *Set) Draw(pix []byte) {
	for i, v := range s.grid {
		if v > 0 {
			r, g, b := valueToColor(v, s.coloring, s.steps)
			pix[4*i] = r
			pix[4*i+1] = g
			pix[4*i+2] = b
			pix[4*i+3] = 0xff
		} else {
			pix[4*i] = 0
			pix[4*i+1] = 0
			pix[4*i+2] = 0
			pix[4*i+3] = 0xff
		}
	}
}

type Work [4]int // Work is piece of screen defined by rect

type ComputeWork struct {
	workload chan Work
	ctx      context.Context
}

type SetComputor struct {
	set          *Set
	neighborhood [NEIGHT_RES * NEIGHT_RES]float64
	workloads    chan ComputeWork
	cancelUpdate *context.CancelFunc
	wg           sync.WaitGroup
	mutex        sync.Mutex
	premutex     sync.Mutex
	progress     int
}

// converts pix from the grid to  ±2 + ±2i complex number
func (c *SetComputor) PixToNeigh(x, y int, s *Set) complex128 {
	zoom := s.zoom / 2.5 / (4 * math.Sqrt2)
	span := NEIGHT_RES
	return complex(
		(float64(x-span/2)/float64(span))/zoom+real(s.mid),
		(float64(y-span/2)/float64(span))/zoom+imag(s.mid),
	)
}

func (sc *SetComputor) NeighToPix(c complex128, s *Set) (x, y int) {
	zoom := s.zoom / 2.5 / (4 * math.Sqrt2)
	span := NEIGHT_RES
	x = int(
		(real(c)-real(s.mid))*zoom*float64(span) + float64(span/2),
	)
	y = int(
		(imag(c)-imag(s.mid))*zoom*float64(span) + float64(span/2),
	)
	return
}

// Draw paints current game state.
func (c *SetComputor) DrawNeigh(pix []byte) {
	for i, v := range c.neighborhood {
		x := i % NEIGHT_RES
		y := i / NEIGHT_RES
		destI := y*c.set.w + x + c.set.w - NEIGHT_RES
		if v > 0 {
			r, g, b := valueToColor(v, c.set.coloring, c.set.steps)
			pix[4*destI] = r
			pix[4*destI+1] = g
			pix[4*destI+2] = b
			pix[4*destI+3] = 0xff
		} else {
			pix[4*destI] = 0
			pix[4*destI+1] = 0
			pix[4*destI+2] = 0
			pix[4*destI+3] = 0xff
		}
	}
}

func (g *SetComputor) Init() {
	s := g.set
	work := func(w Work) {
		setBlackPix := func(x, y int) {
			s.grid[(y*s.w+x)%len(s.grid)] = 0 // set all black
		}
		computePix := func(x, y int) bool {
			ok, st := isInSet(s.PixToSet(x, y), s.steps)
			if ok {
				setBlackPix(x, y)
			} else {
				s.grid[(y*s.w+x)%len(s.grid)] = st
			}
			return ok
		}
		x0, x1, y0, y1 := w[0], w[1], w[2], w[3]
		for x0 <= x1 && y0 <= y1 {
			isThisFrame := true
			for x := x0; x < x1; x++ { // top and bottom border
				if !computePix(x, y0) {
					isThisFrame = false
				}
				if !computePix(x, y1-1) {
					isThisFrame = false
				}
			}
			y0++
			y1--
			for y := y0; y < y1; y++ { // left and right border
				if !computePix(x0, y) {
					isThisFrame = false
				}
				if !computePix(x1-1, y) {
					isThisFrame = false
				}
			}
			x0++
			x1--
			if isThisFrame {
				break
			}
		}
		for y, y_ := y0, y1-1; y < y1; y, y_ = y+1, y_-1 { // each line
			y := y
			if y0 < s.h/2 { // upper half, change direction of line handling
				y = y_
			}
			for x := x0; x < x1; x++ {
				setBlackPix(x, y)
			}
		}
	}
	worker := func(workload chan Work) {
		for w := range workload {
			work(w)
			g.wg.Done()
			g.progress++
		}
	}
	manager := func(computeWorks chan ComputeWork) {
		var workload chan Work
		for cw := range computeWorks { // cw represent current Compute
			if workload != nil {
				close(workload)
			}
			workload = make(chan Work)     // spawn new workers for each cw
			for i := 0; i < WORKERS; i++ { // spawn workers
				go worker(workload)
			}
		currentWorkLoop:
			for {
				select {
				case <-cw.ctx.Done():
					for range cw.workload {
						// g.progress++
						g.wg.Done()
					}
					break currentWorkLoop
				case w, ok := <-cw.workload:
					if !ok {
						break currentWorkLoop
					}
					workload <- w
				}
			}
		}
	}

	g.workloads = make(chan ComputeWork) // TODO close when app exits
	go manager(g.workloads)
}

func (g *SetComputor) Compute() {
	if g.workloads == nil { // init workers on firt run
		g.Init()
	}

	if g.cancelUpdate != nil { // cancel existing Compute run
		(*g.cancelUpdate)()
	}

	if !g.premutex.TryLock() { // seomeone already waiting for mutex
		return
	}
	g.mutex.Lock() // allow only one gorutine waiting for this lock by wrapping it to premutex
	g.premutex.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelUpdate = &cancel
	defer func() {
		g.cancelUpdate = nil
		g.mutex.Unlock()
	}()

	// poush current workload to workloads
	workload := make(chan Work)
	g.workloads <- ComputeWork{ // cw set to the manager
		workload,
		ctx,
	}

	g.progress = 0 // reset progress
	pcs := PCS     // number of rows and cols of chunks, must be even
	g.wg.Add(pcs * pcs)
	go func() {
		s := g.set
		i0 := pcs/2 - 1
		j0 := pcs/2 - 1
		i1 := pcs / 2
		j1 := pcs / 2
		sendWork := func(i, j int) {
			ysize := s.h / pcs
			ystart := j * ysize
			xsize := s.w / pcs
			xstart := i * xsize
			workload <- Work{xstart, xstart + xsize, ystart, ystart + ysize}
		}
	loop:
		for {
			for j, j_ := pcs/2-1, pcs/2; j >= j0; j, j_ = j-1, j_+1 {
				sendWork(i0, j)
				sendWork(i1, j)
				sendWork(i0, j_)
				sendWork(i1, j_)
			}
			if j0 == 0 && i0 == 0 {
				break loop
			}
			j0--
			j1++
			for i, i_ := pcs/2-1, pcs/2; i >= i0; i, i_ = i-1, i_+1 {
				sendWork(i, j0)
				sendWork(i, j1)
				sendWork(i_, j1)
				sendWork(i_, j0)
			}
			i0--
			i1++
		}
		close(workload)
	}()

	g.wg.Wait()
	g.progress = 1024
}
