package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestWFC(t *testing.T) {
	m := make(Tilemap)
	tiles := []*Tile{
		{Spritesheet: ".", Index: 0}, // [.]
		{Spritesheet: "|", Index: 1}, // [|]
		{Spritesheet: "-", Index: 2}, // [-]
		{Spritesheet: "+", Index: 3}, // [+]
	}
	/*
		+-+ +-+
		|.| |.|
		+-+ +-+
	*/
	m.Set(tiles[3], 0, 0, false, 0)
	m.Set(tiles[2], 1, 0, false, 0)
	m.Set(tiles[3], 2, 0, false, 0)
	m.Set(tiles[1], 0, 1, false, 0)
	m.Set(tiles[0], 1, 1, false, 0)
	m.Set(tiles[1], 2, 1, false, 0)
	m.Set(tiles[3], 0, 2, false, 0)
	m.Set(tiles[2], 1, 2, false, 0)
	m.Set(tiles[3], 2, 2, false, 0)
	m.Set(tiles[3], 0, 0, false, 1)
	m.Set(tiles[2], 1, 0, false, 1)
	m.Set(tiles[3], 2, 0, false, 1)
	m.Set(tiles[1], 0, 1, false, 1)
	m.Set(tiles[0], 1, 1, false, 1)
	m.Set(tiles[1], 2, 1, false, 1)
	m.Set(tiles[3], 0, 2, false, 1)
	m.Set(tiles[2], 1, 2, false, 1)
	m.Set(tiles[3], 2, 2, false, 1)
	g := NewGenerator(m, 6, 6, time.Now().UnixMilli())
	if got, want := g.Width, 6; got != want {
		t.Fatalf("wrong width, got %d, want %d", got, want)
	}
	if got, want := g.Height, 6; got != want {
		t.Fatalf("wrong height, got %d, want %d", got, want)
	}
	if got, want := g.Depth, 2; got != want {
		t.Fatalf("wrong depth, got %d, want %d", got, want)
	}
	if got, want := len(g.Domain), 4; got != want {
		t.Fatalf("wrong domain, got %d, want %d", got, want)
	}
	wantAdj := make([][6][]int, 4)
	// below, above, north, south, west, east
	wantAdj[g.Index(tiles[0])] = [6][]int{{g.Index(tiles[0])}, {g.Index(tiles[0])}, {g.Index(tiles[2])}, {g.Index(tiles[2])}, {g.Index(tiles[1])}, {g.Index(tiles[1])}}
	wantAdj[g.Index(tiles[1])] = [6][]int{{g.Index(tiles[1])}, {g.Index(tiles[1])}, {g.Index(tiles[3])}, {g.Index(tiles[3])}, {g.Index(tiles[0])}, {g.Index(tiles[0])}}
	wantAdj[g.Index(tiles[2])] = [6][]int{{g.Index(tiles[2])}, {g.Index(tiles[2])}, {g.Index(tiles[0])}, {g.Index(tiles[0])}, {g.Index(tiles[3])}, {g.Index(tiles[3])}}
	wantAdj[g.Index(tiles[3])] = [6][]int{{g.Index(tiles[3])}, {g.Index(tiles[3])}, {g.Index(tiles[1])}, {g.Index(tiles[1])}, {g.Index(tiles[2])}, {g.Index(tiles[2])}}
	if got, want := g.Adj, wantAdj; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong adj, got %v, want %v", got, want)
	}
	g.Init()
	for !g.Done() {
	}
	newMap := g.Readout()
	for z := 0; z < g.Depth; z++ {
		for y := 0; y < g.Height; y++ {
			for x := 0; x < g.Width; x++ {
				if newMap[x][y] == nil || z >= len(newMap[x][y]) {
					fmt.Print(" ")
				} else {
					fmt.Printf("%s", newMap[x][y][z].Spritesheet)
				}
			}
			fmt.Println()
		}
		fmt.Println()
	}
}
