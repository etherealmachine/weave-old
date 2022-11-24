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
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if tile := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTile(ui.SelectedTileIndex); tile != nil {
			x, y := ebiten.CursorPosition()
			bounds := tile.Bounds()
			w, h := ui.Scale*float64(bounds.Dx()), ui.Scale*float64(bounds.Dy())
			tileX := int(float64(x) / w)
			tileY := int(float64(y) / h)
			ui.Tilemap.SetTile(ui.SelectedTileset, ui.SelectedTileIndex, tileX, tileY)
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
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
}

func (ui *UI) SelectTileset(event *bento.Event) {
	if event.Box.Content != ui.SelectedTileset {
		ui.SelectedTileset = event.Box.Content
		ui.SelectedTileIndex = 0
	}
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
		<col float="true" justifySelf="end" margin="16px">
			<row grow="1" justify="between" margin="0 0 12px 0">
				{{ range $name, $tileset := .Tilemap.Tilesets }}
					<button
							font="NotoSans 18"
							btn="button.png 6"
							color="#ffffff"
							padding="12px"
							underline="{{ eq $.SelectedTileset $name }}"
							onClick="SelectTileset"
					>{{ $name }}</button>
				{{ end }}
			</row>
			<img onClick="SelectTile" src="{{ .SelectedTileset }}" scale="2" zIndex="100" />
		</col>
	</col>`
}
