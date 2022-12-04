package main

import (
	"fmt"
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
		+-+ -+-
		|.| .|.
		+-+ -+-
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
	m.Set(tiles[2], 0, 0, false, 1)
	m.Set(tiles[3], 1, 0, false, 1)
	m.Set(tiles[2], 2, 0, false, 1)
	m.Set(tiles[0], 0, 1, false, 1)
	m.Set(tiles[1], 1, 1, false, 1)
	m.Set(tiles[0], 2, 1, false, 1)
	m.Set(tiles[2], 0, 2, false, 1)
	m.Set(tiles[3], 1, 2, false, 1)
	m.Set(tiles[2], 2, 2, false, 1)
	g := NewGenerator(m)
	if got, want := len(g.Domain), 5; got != want {
		t.Fatalf("wrong domain, got %d, want %d", got, want)
	}
	g.Verify = true
	g.Init(6, 6, nil, time.Now().UnixMilli())
	if got, want := g.Width, 6; got != want {
		t.Fatalf("wrong width, got %d, want %d", got, want)
	}
	if got, want := g.Height, 6; got != want {
		t.Fatalf("wrong height, got %d, want %d", got, want)
	}
	for !g.Done() {
	}
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			if i := g.Map.At(x, y); i != nil {
				fmt.Printf("%s", g.Domain[*i].Hash())
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}
