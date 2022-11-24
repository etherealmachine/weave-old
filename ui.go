package main

import (
	"log"
	"math"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
)

type UI struct {
	SelectedTileset   string
	SelectedTileIndex int
	Tilemap           *Tilemap
	Scale             float64
	OffsetX, OffsetY  float64
	DragX, DragY      int
}

func NewUI() *UI {
	ui := &UI{
		Tilemap: NewTilemap(),
		Scale:   1,
	}
	if err := ui.Tilemap.AddTileset("dungeon.png", 16, 1); err != nil {
		log.Fatal(err)
	}
	if err := ui.Tilemap.AddTileset("general.png", 16, 1); err != nil {
		log.Fatal(err)
	}
	if err := ui.Tilemap.AddTileset("indoors.png", 16, 1); err != nil {
		log.Fatal(err)
	}
	if err := ui.Tilemap.AddTileset("characters.png", 16, 1); err != nil {
		log.Fatal(err)
	}
	ui.SelectedTileset = "dungeon.png"
	return ui
}

func (ui *UI) Draw(img *ebiten.Image) {
	if tile := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTile(ui.SelectedTileIndex); tile != nil {
		x, y := ebiten.CursorPosition()
		bounds := tile.Bounds()
		w, h := ui.Scale*float64(bounds.Dx()), ui.Scale*float64(bounds.Dy())
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Scale(ui.Scale, ui.Scale)
		op.GeoM.Translate(math.Floor(float64(x)/w)*w, math.Floor(float64(y)/h)*h)
		img.DrawImage(tile, op)
	}
}

func (ui *UI) Update(event *bento.Event) bool {
	_, sy := ebiten.Wheel()
	if sy != 0 {
		if sy > 0 {
			ui.Scale *= 1.1
		} else {
			ui.Scale /= 1.1
		}
	}
	return false
}

func (ui *UI) Hover(event *bento.Event) {
	/*
		tileX := int(float64(event.X) / (float64(ui.TileWidth) * ui.Scale))
		tileY := int(float64(event.Y) / (float64(ui.TileWidth) * ui.Scale))
		if ui.Input.ActionIsPressed(ActionPaste) && !ebiten.IsKeyPressed(ebiten.KeyControl) && ui.SelectedTileIndex > 0 {
			layer := ui.Layers[ui.SelectedLayer]
			layer.Tiles[tileY*layer.Width+tileX] = ui.SelectedTileIndex
			ui.Save()
		} else if ui.Input.ActionIsPressed(ActionPaste) {
			layer := ui.Layers[ui.SelectedLayer]
			layer.Tiles[tileY*layer.Width+tileX] = 0
			ui.Save()
		} else if ui.Input.ActionIsPressed(ActionDrag) {
			if ui.DragX != 0 || ui.DragY != 0 {
				ui.OffsetX += float64(event.X-ui.DragX) / ui.Scale
				ui.OffsetY += float64(event.Y-ui.DragY) / ui.Scale
			}
			ui.DragX = event.X
			ui.DragY = event.Y
		} else {
			ui.DragX = 0
			ui.DragY = 0
		}
	*/
}

func (ui *UI) SelectTile(event *bento.Event) {
	ui.SelectedTileIndex = ui.Tilemap.Tilesets[ui.SelectedTileset].TileAt(event.X/2, event.Y/2)
}

func (ui *UI) UI() string {
	return `<col grow="1">
		<row grow="1">
			<col grow="1">
				<canvas grow="1" onDraw="Draw" onHover="Hover" onUpdate="Update" />
			</col>
		</row>
		<row float="true" justify="end" margin="16px">
			<img onClick="SelectTile" src="{{ .SelectedTileset }}" scale="2" zIndex="100" />
		</row>
	</col>`
}
