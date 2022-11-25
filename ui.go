package main

import (
	"log"
	"math"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type UI struct {
	SelectedTileset   string
	SelectedTileIndex int
	Tilemap           *Tilemap
	Scale             float64
	OffsetX, OffsetY  float64
	DragX, DragY      int
	Frame             *bento.NineSlice
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
	img, _, err := ebitenutil.NewImageFromFile("frame.png")
	if err != nil {
		log.Fatal(err)
	}
	ui.Frame = bento.NewNineSlice(img, [3]int{4, 24, 4}, [3]int{4, 24, 4}, 0, 0)
	return ui
}

func (ui *UI) Draw(event *bento.Event) {
	for x, ys := range ui.Tilemap.Tiles {
		for y, tiles := range ys {
			for _, tile := range tiles {
				img := ui.Tilemap.Tilesets[tile.Tileset].GetTile(tile.Index)
				bounds := img.Bounds()
				w, h := float64(bounds.Dx()), float64(bounds.Dy())
				op := new(ebiten.DrawImageOptions)
				op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
				op.GeoM.Translate(float64(x)*w, float64(y)*h)
				op.GeoM.Translate(math.Floor(float64(ui.OffsetX)/16)*16, math.Floor(float64(ui.OffsetY)/16)*16)
				op.GeoM.Scale(ui.Scale, ui.Scale)
				event.Image.DrawImage(img, op)
			}
		}
	}
	if tile := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTile(ui.SelectedTileIndex); tile != nil {
		x, y := ebiten.CursorPosition()
		bounds := tile.Bounds()
		w, h := ui.Scale*float64(bounds.Dx()), ui.Scale*float64(bounds.Dy())
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
		op.GeoM.Scale(ui.Scale, ui.Scale)
		op.GeoM.Translate(math.Floor(float64(x)/w)*w, math.Floor(float64(y)/h)*h)
		event.Image.DrawImage(tile, op)
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
	if tile := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTile(ui.SelectedTileIndex); tile != nil {
		x, y := ebiten.CursorPosition()
		bounds := tile.Bounds()
		w, h := ui.Scale*float64(bounds.Dx()), ui.Scale*float64(bounds.Dy())
		tileX := int((float64(x) - ui.Scale*math.Floor(ui.OffsetX/16)*16) / w)
		tileY := int((float64(y) - ui.Scale*math.Floor(ui.OffsetY/16)*16) / h)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			ui.Tilemap.SetTile(ui.SelectedTileset, ui.SelectedTileIndex, tileX, tileY, ebiten.IsKeyPressed(ebiten.KeyShift))
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			ui.Tilemap.SetTile("", 0, tileX, tileY, false)
		}
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

func (ui *UI) DrawSelectedTiles(event *bento.Event) {
	rect := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTileRect(ui.SelectedTileIndex)
	if rect == nil {
		return
	}
	ui.Frame.Draw(event.Image, rect.Min.X*2, rect.Min.Y*2, rect.Dx()*2, rect.Dy()*2, event.Op)
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
			<img onClick="SelectTile" onDraw="DrawSelectedTiles" src="{{ .SelectedTileset }}" scale="2" zIndex="100" />
		</col>
	</col>`
}
