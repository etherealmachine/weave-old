package main

import (
	"image"
	"log"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type UI struct {
	TileX, TileY int
	Tileset      *ebiten.Image
	Tile         *ebiten.Image
}

func (ui *UI) Draw(img *ebiten.Image) {
	if ui.Tile != nil {
		x, y := ebiten.CursorPosition()
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(float64(x)-8, float64(y)-8)
		img.DrawImage(ui.Tile, op)
	}
}

func (ui *UI) Update() error {
	return nil
}

func (ui *UI) Click(event *bento.Event) {
	if ui.Tileset == nil {
		var err error
		ui.Tileset, _, err = ebitenutil.NewImageFromFile("dungeon.png")
		if err != nil {
			log.Fatal(err)
		}
	}
	ui.TileX = event.X / 17
	ui.TileY = event.Y / 17
	ui.Tile = ebiten.NewImageFromImage(
		ui.Tileset.SubImage(image.Rect(ui.TileX*17, ui.TileY*17, (ui.TileX+1)*17, (ui.TileY+1)*17)))
}

func (ui *UI) UI() string {
	return `<col grow="1">
		<canvas grow="1" draw="Draw" update="Update" />
		<row grow="1 0" justify="end" margin="16px">
			<img onClick="Click" src="dungeon.png" />
		</row>
	</col>`
}
