package sim

import (
	"github.com/faiface/pixel"
	"math/rand"
	"sync"
	"time"
)

const Partcount = 12 * 60
const Species = 12
const Radius = 8
const Threads = 12
const Friction = 0.88
const AttractionPrescaler = 0.0011
const AttractionScaler = 0.015
const CollisionEnabled = false

type Particle struct {
	Position pixel.Vec
	Color    pixel.RGBA

	momentum pixel.Vec
	nextPos  pixel.Vec
}

type Sim struct {
	Parts []*Particle

	attractionLookup map[pixel.RGBA]map[pixel.RGBA]float64

	width, height float64
}

func Init(width, height float64) *Sim {
	rand.Seed(time.Now().UnixNano())

	retval := &Sim{
		Parts:  genParts(Partcount, width, height),
		width:  width,
		height: height,
	}

	retval.attractionLookup = genLookup(retval.Parts[:Species])

	return retval
}

func genLookup(parts []*Particle) map[pixel.RGBA]map[pixel.RGBA]float64 {
	retval := make(map[pixel.RGBA]map[pixel.RGBA]float64)

	for _, master := range parts {
		retval[master.Color] = make(map[pixel.RGBA]float64)
		for _, checkee := range parts {
			if checkee.Color == master.Color {
				retval[master.Color][checkee.Color] = rand.Float64() - 0.5
			} else {
				retval[master.Color][checkee.Color] = rand.Float64() - 0.5
			}
		}
	}

	return retval
}

func genParts(count int, width, height float64) []*Particle {
	retval := make([]*Particle, count)

	for i := 0; i < Species; i++ {
		retval[i] = &Particle{
			Position: randVec(width, height),
			Color:    pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64()),
			momentum: pixel.V(0, 0),
		}
	}

	for i := Species; i < Partcount; i++ {
		cloneOf := retval[rand.Intn(Species)]
		retval[i] = &Particle{
			Position: randVec(width, height),
			Color:    cloneOf.Color,
			momentum: pixel.V(0, 0),
		}

		for collides(retval[i], retval) {
			retval[i].Position = randVec(width, height)
		}
	}

	return retval
}

func (sim *Sim) Step() {

	threadCount := Partcount / Threads

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(Threads)

	for i := 0; i < Threads; i++ {
		go sim.updateMomentum(i*threadCount, (i+1)*threadCount, &waitGroup)
	}

	waitGroup.Wait()
	waitGroup.Add(Threads)

	for i := 0; i < Threads; i++ {
		go sim.updateNextPosition(i*threadCount, (i+1)*threadCount, &waitGroup)
	}

	waitGroup.Wait()
	waitGroup.Add(Threads)

	for i := 0; i < Threads; i++ {
		go sim.updatePosition(i*threadCount, (i+1)*threadCount, &waitGroup)
	}

	waitGroup.Wait()
}

func randVec(maxX, maxY float64) pixel.Vec {
	return pixel.V(rand.Float64()*maxX, rand.Float64()*maxY)
}

func collides(p *Particle, parts []*Particle) bool {
	for _, other := range parts {
		if other != nil && other != p && p.Position.Sub(other.Position).Len() < Radius*2 {
			return true
		}
	}

	return false
}

func collidesNext(p *Particle, parts []*Particle) bool {
	for _, other := range parts {
		if other != nil && other != p && p.nextPos.Sub(other.nextPos).Len() < Radius*2 {
			return true
		}
	}

	return false
}

func (sim *Sim) updatePosition(from, to int, waitGroup *sync.WaitGroup) {
	for i := from; i < to; i++ {
		if !CollisionEnabled || !collidesNext(sim.Parts[i], sim.Parts) {
			sim.Parts[i].Position = sim.Parts[i].nextPos
		}
	}

	waitGroup.Done()
}

func (sim *Sim) updateNextPosition(from, to int, waitGroup *sync.WaitGroup) {
	for i := from; i < to; i++ {
		newPos := sim.Parts[i].Position
		newPos = newPos.Add(sim.Parts[i].momentum)

		for newPos.X > sim.width {
			newPos.X -= sim.width
		}
		for newPos.Y > sim.height {
			newPos.Y -= sim.height
		}

		for newPos.X < 0 {
			newPos.X = sim.width + newPos.X
		}
		for newPos.Y < 0 {
			newPos.Y = sim.height + newPos.Y
		}

		sim.Parts[i].nextPos = newPos
	}

	waitGroup.Done()
}

func (sim *Sim) updateMomentum(from, to int, waitGroup *sync.WaitGroup) {
	for i := from; i < to; i++ {
		sim.Parts[i].momentum = sim.Parts[i].momentum.Scaled(Friction)

		for _, other := range sim.Parts {

			diffWidth := sim.width
			if sim.Parts[i].Position.X < other.Position.X {
				diffWidth *= -1
			}

			diffHeight := sim.height
			if sim.Parts[i].Position.Y < other.Position.Y {
				diffHeight *= -1
			}

			diffVec := other.Position.Sub(pixel.V(sim.Parts[i].Position.X, sim.Parts[i].Position.Y)).Scaled(AttractionPrescaler)
			sim.Parts[i].momentum = sim.Parts[i].momentum.Add(diffVec.Scaled(sim.calculateAttractionMagnitude(sim.Parts[i], other)))

			if sim.Parts[i].Position.X != other.Position.X {
				diffVec = other.Position.Sub(pixel.V(sim.Parts[i].Position.X-diffWidth, sim.Parts[i].Position.Y)).Scaled(AttractionPrescaler)
				sim.Parts[i].momentum = sim.Parts[i].momentum.Add(diffVec.Scaled(sim.calculateAttractionMagnitude(sim.Parts[i], other)))
			}

			if sim.Parts[i].Position.Y != other.Position.Y {
				diffVec = other.Position.Sub(pixel.V(sim.Parts[i].Position.X, sim.Parts[i].Position.Y-diffHeight)).Scaled(AttractionPrescaler)
				sim.Parts[i].momentum = sim.Parts[i].momentum.Add(diffVec.Scaled(sim.calculateAttractionMagnitude(sim.Parts[i], other)))
			}

			if sim.Parts[i].Position.X != other.Position.X && sim.Parts[i].Position.Y != other.Position.Y {
				diffVec = other.Position.Sub(pixel.V(sim.Parts[i].Position.X-diffWidth, sim.Parts[i].Position.Y-diffHeight)).Scaled(AttractionPrescaler)
				sim.Parts[i].momentum = sim.Parts[i].momentum.Add(diffVec.Scaled(sim.calculateAttractionMagnitude(sim.Parts[i], other)))
			}
		}
	}

	waitGroup.Done()
}

func (sim *Sim) calculateAttractionMagnitude(a, b *Particle) float64 {
	return sim.attractionLookup[a.Color][b.Color] * AttractionScaler

	ahash := int(a.Color.R*255) + int(a.Color.G*255) + int(a.Color.B*255)
	bhash := int(b.Color.R*255) + int(b.Color.G*255) + int(b.Color.B*255)
	neg := 1
	if (ahash & 0x4) != 0 {
		neg = -1
	}

	return float64((ahash+bhash)*neg) * AttractionScaler
}
