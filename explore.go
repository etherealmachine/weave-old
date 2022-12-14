package main

import (
	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
)

type Explore struct {
	Map       *Map
	MapScale  float64
	Character *Character
}

type Character struct {
	TileX, TileY int
	Sprite       *ebiten.Image
}

func NewExplore(m *Map) *Explore {
	return &Explore{Map: m, MapScale: 1, Character: &Character{
		Sprite: m.TileImage(&Tile{Spritesheet: "characters.png", Index: 529}),
	}}
}

func (ui *Explore) Draw(event *bento.Event) {
	ui.drawMap(event)
	op := new(ebiten.DrawImageOptions)
	op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
	op.GeoM.Scale(ui.MapScale, ui.MapScale)
	event.Image.DrawImage(ui.Character.Sprite, op)
}

func (ui *Explore) drawMap(event *bento.Event) {
	w, h := float64(ui.Map.TileWidth), float64(ui.Map.TileHeight)
	ox, oy := float64(ui.Character.TileX), float64(ui.Character.TileY)
	for x, ys := range ui.Map.Tilemap {
		for y, tiles := range ys {
			for _, tile := range tiles {
				img := ui.Map.TileImage(tile)
				op := new(ebiten.DrawImageOptions)
				op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
				op.GeoM.Translate(float64(x)*w, float64(y)*h)
				op.GeoM.Translate(ox, oy)
				op.GeoM.Scale(ui.MapScale, ui.MapScale)
				//op.GeoM.Skew(-0.7, 0)
				event.Image.DrawImage(img, op)
			}
		}
	}
}

func (ui *Explore) Click(event *bento.Event) {
}

func (ui *Explore) Hover(event *bento.Event) {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		ui.Character.TileY++
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		ui.Character.TileY--
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		ui.Character.TileX++
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		ui.Character.TileX--
	}
}

func (ui *Explore) OnMapScroll(event *bento.Event) bool {
	_, sy := ebiten.Wheel()
	if sy != 0 {
		if sy > 0 {
			ui.MapScale *= 1.1
		} else {
			ui.MapScale /= 1.1
		}
	}
	return false
}

func (ui *Explore) UI() string {
	return `<col grow="1">
		<canvas grow="1" onDraw="Draw" onClick="Click" onHover="Hover" onScroll="OnMapScroll" />
	</col>`
}
