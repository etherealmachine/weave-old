package main

import (
	"image"
	"log"
	"math"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type UI struct {
	SelectedTileIndex *int
	Tileset           *ebiten.Image
	Tiles             []*ebiten.Image
	Width             int
}

func (ui *UI) Draw(img *ebiten.Image) {
	if ui.Tiles != nil && ui.SelectedTileIndex != nil {
		x, y := ebiten.CursorPosition()
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(math.Floor(float64(x)/16)*16, math.Floor(float64(y)/16)*16)
		img.DrawImage(ui.Tiles[*ui.SelectedTileIndex], op)
	}
}

func (ui *UI) Update() error {
	return nil
}

func (ui *UI) PasteTile(event *bento.Event) {
	log.Println(event.X, event.Y)
}

func (ui *UI) SelectTile(event *bento.Event) {
	if ui.Tileset == nil {
		var err error
		ui.Tileset, _, err = ebitenutil.NewImageFromFile("dungeon.png")
		if err != nil {
			log.Fatal(err)
		}
		bounds := ui.Tileset.Bounds()
		width := bounds.Dx() / 17
		height := bounds.Dy() / 17
		ui.Width = width
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				tile := ebiten.NewImageFromImage(
					ui.Tileset.SubImage(
						image.Rect(x*17, y*17, (x+1)*17, (y+1)*17)))
				ui.Tiles = append(ui.Tiles, tile)
			}
		}
	}
	tileX := event.X / 17
	tileY := event.Y / 17
	i := tileY*ui.Width + tileX
	ui.SelectedTileIndex = &i
}

func (ui *UI) UI() string {
	return `<col grow="1">
		<canvas grow="1" draw="Draw" update="Update" onClick="PasteTile" />
		<row grow="1 0" justify="end" margin="16px">
			<img onClick="SelectTile" src="dungeon.png" />
		</row>
	</col>`
}
