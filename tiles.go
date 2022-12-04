package main

import (
	"fmt"
	"strings"
)

type Tile struct {
	Spritesheet string
	Index       int
}

func (t *Tile) Hash() string {
	return fmt.Sprintf("%s:%d", t.Spritesheet, t.Index)
}

type Stack []*Tile

func (s Stack) Hash() string {
	a := make([]string, len(s))
	for i, t := range s {
		a[i] = t.Hash()
	}
	return strings.Join(a, ",")
}

type Tilemap map[int]map[int]Stack

func (m Tilemap) Set(tile *Tile, x, y int, replace bool, z int) {
	if m[x] == nil {
		m[x] = make(map[int]Stack)
	}
	l := len(m[x][y])
	if l == 0 {
		// first tile in the stack
		m[x][y] = Stack{tile}
	} else if z >= l {
		// append
		m[x][y] = append(m[x][y], tile)
	} else if replace {
		// replace
		m[x][y][z] = tile
	} else {
		// insert
		m[x][y] = append(m[x][y][:z+1], m[x][y][z:]...)
		m[x][y][z] = tile
	}
}

func (m Tilemap) At(x, y, z int) *Tile {
	if len(m[x]) == 0 {
		return nil
	}
	if len(m[x][y]) == 0 {
		return nil
	}
	if z < 0 || z >= len(m[x][y]) {
		return nil
	}
	return m[x][y][z]
}
