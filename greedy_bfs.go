package main

/*
GreedyBFS: pick a random allowed adjacent tile
*/
type GreedyBFS struct {
	*Analysis
}

func NewGreedyBFS(analysis *Analysis, width, height int, fixed Tilemap, seed int64) *GreedyBFS {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if fixed[x][y] != nil {

				fixed[x][y].Hash()
			}
		}
	}
	var queue [][2]int
	for x := 0; x < width; x++ {
		for y := range []int{0, height - 1} {
			queue = append(queue, [2]int{x, y})
		}
	}
	for y := 0; y < height; y++ {
		for x := range []int{0, width - 1} {
			queue = append(queue, [2]int{x, y})
		}
	}
	return &GreedyBFS{Analysis: analysis}
}

func (g *GreedyBFS) Done() bool {
	return true
}

func (g *GreedyBFS) Result() [][]Stack {
	return nil
}
