package main

import "math/rand"

type Direction int

const (
	North = Direction(0)
	South = Direction(1)
	West  = Direction(2)
	East  = Direction(3)
)

func (d Direction) String() string {
	switch d {
	case North:
		return "north"
	case South:
		return "south"
	case West:
		return "west"
	case East:
		return "east"
	default:
		return "unknown"
	}
}

func (d Direction) Inverse() Direction {
	if d%2 == 0 {
		return d + 1
	}
	return d - 1
}

var Neighbors = [4][2]int{
	{0, -1}, // North
	{0, 1},  // South
	{-1, 0}, // West
	{1, 0},  // East
}

type Analysis struct {
	Domain        []Stack
	DomainIndex   map[string]int
	Probabilities []float64
	Adj           *NDArray[map[int]bool] // Domain, Neighbors
}

func Analyze(tilemap Tilemap) *Analysis {
	domainIndex := map[string]int{
		"": 0,
	}
	for _, ys := range tilemap {
		for _, tiles := range ys {
			if h := tiles.Hash(); domainIndex[h] == 0 {
				domainIndex[h] = len(domainIndex)
			}
		}
	}
	probs := make([]float64, len(domainIndex))
	adj := NewNDArray[map[int]bool](len(domainIndex), len(Neighbors))
	domain := make([]Stack, len(domainIndex))
	for x, ys := range tilemap {
		for y, tiles := range ys {
			i := domainIndex[tiles.Hash()]
			domain[i] = tiles
			probs[i]++
			for d, o := range Neighbors {
				nx, ny := x+o[0], y+o[1]
				n := domainIndex[tilemap[nx][ny].Hash()]
				a := adj.At(i, d)
				if a == nil {
					a = make(map[int]bool)
					adj.Set(a, i, d)
				}
				a[n] = true
				di := int(Direction(d).Inverse())
				a = adj.At(n, di)
				if a == nil {
					a = make(map[int]bool)
					adj.Set(a, n, di)
				}
				a[i] = true
			}
		}
	}
	var sum float64
	for _, count := range probs {
		sum += count
	}
	for i, count := range probs {
		probs[i] = count / sum
	}
	return &Analysis{
		Domain:        domain,
		DomainIndex:   domainIndex,
		Probabilities: probs,
		Adj:           adj,
	}
}

func (a *Analysis) Lottery(rng *rand.Rand, allowed func(i int) bool) int {
	var ticketCount float64
	tickets := make(map[int]float64)
	for i := range a.Domain {
		if !allowed(i) {
			continue
		}
		tickets[i] = a.Probabilities[i]
		ticketCount += a.Probabilities[i]
	}
	ticket := rng.Float64() * ticketCount
	winner := -1
	for i := 0; i < len(a.Domain); i++ {
		ticket -= tickets[i]
		if winner == -1 && ticket <= 0 {
			winner = i
		}
	}
	return winner
}
