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
	g := NewWFC(Analyze(m), 6, 6, nil, time.Now().UnixMilli())
	if got, want := g.width, 6; got != want {
	}
	if got, want := g.height, 6; got != want {
		t.Fatalf("wrong height, got %d, want %d", got, want)
	}
	for !g.Done() {
	}
	result := g.Result()
	if got, want := len(result), 6; got != want {
		t.Fatalf("wrong width, got %d, want %d", got, want)
	}
	if got, want := len(result[0]), 6; got != want {
		t.Fatalf("wrong height, got %d, want %d", got, want)
	}
	for x := 0; x < len(result); x++ {
		for y := 0; y < len(result[x]); y++ {
			fmt.Printf("%s", result[x][y].Hash())
		}
		fmt.Println()
	}
}
