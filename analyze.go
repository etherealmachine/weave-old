package main

type Analysis struct {
	Domain        []Stack
	DomainIndex   map[string]int
	Probabilities []float64
	Adj           *NDArray[map[int]bool]
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

/*
func (g *WFCGenerator) verifyPlacement(x, y int) {
	i := g.Map.At(x, y)
	if i == nil {
		return
	}
	for d, o := range Neighbors {
		nx, ny := x+o[0], y+o[1]
		if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height {
			continue
		}
		n := g.Map.At(nx, ny)
		if n == nil {
			continue
		}
		if !g.Adj.At(*i, d)[*n] {
			log.Fatalf("incorrect placement at (%d, %d): %d cannot have %d %s",
				x, y,
				*i, *n,
				Direction(d))
		}
	}
}
*/
