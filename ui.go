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
	SelectedTileIndex int
	Tileset           *ebiten.Image
	Tiles             []*ebiten.Image
	Width             int
	Map               []int
}

func (ui *UI) Draw(img *ebiten.Image) {
	if ui.Tiles != nil && ui.SelectedTileIndex > 0 {
		x, y := ebiten.CursorPosition()
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(math.Floor(float64(x)/16)*16, math.Floor(float64(y)/16)*16)
		img.DrawImage(ui.Tiles[ui.SelectedTileIndex-1], op)
	}
	if ui.Map != nil {
		for y := 0; y < len(ui.Map)/ui.Width; y++ {
			for x := 0; x < ui.Width; x++ {
				if ui.Map[y*ui.Width+x] > 0 {
					op := new(ebiten.DrawImageOptions)
					op.GeoM.Translate(float64(x*16), float64(y*16))
					img.DrawImage(ui.Tiles[ui.Map[y*ui.Width+x]-1], op)
				}
			}
		}
	}
}

func (ui *UI) Update() error {
	return nil
}

func (ui *UI) PasteTile(event *bento.Event) {
	if ui.Map == nil {
		bounds := event.Box.Bounds()
		ui.Map = make([]int, (bounds.Dx()/16)*(bounds.Dy()/16))
	}
	if ui.SelectedTileIndex > 0 {
		tileX := event.X / 16
		tileY := event.Y / 16
		ui.Map[tileY*ui.Width+tileX] = ui.SelectedTileIndex
	}
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
	ui.SelectedTileIndex = (tileY*ui.Width + tileX) + 1
}

func (ui *UI) UI() string {
	return `<col grow="1">
		<canvas grow="1" draw="Draw" update="Update" onClick="PasteTile" />
		<row grow="1 0" justify="end" margin="16px">
			<img onClick="SelectTile" src="dungeon.png" />
		</row>
	</col>`
}
