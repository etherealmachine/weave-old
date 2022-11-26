package main

import (
	"log"
	"math"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type UI struct {
	SelectedTileset        string
	SelectedTileIndex      int
	Tilemap                *Tilemap
	MapScale, TilesetScale float64
	OffsetX, OffsetY       float64
	DragX, DragY           int
	Frame                  *bento.NineSlice
}

func NewUI() *UI {
	ui := &UI{
		Tilemap:      NewTilemap(16, 16),
		MapScale:     1,
		TilesetScale: 2,
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
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		ui.drawHoverTile(event)
	}
	ui.drawMap(event)
	if !ebiten.IsKeyPressed(ebiten.KeyControl) {
		ui.drawHoverTile(event)
	}
}

func (ui *UI) drawMap(event *bento.Event) {
	for x, ys := range ui.Tilemap.Tiles {
		for y, tiles := range ys {
			for _, tile := range tiles {
				img := ui.Tilemap.Tilesets[tile.Tileset].GetTile(tile.Index)
				w, h := float64(ui.Tilemap.TileWidth), float64(ui.Tilemap.TileHeight)
				ox, oy := math.Floor(ui.OffsetX/w)*w, math.Floor(ui.OffsetY/h)*h
				op := new(ebiten.DrawImageOptions)
				op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
				op.GeoM.Translate(float64(x)*w, float64(y)*h)
				op.GeoM.Translate(ox, oy)
				op.GeoM.Scale(ui.MapScale, ui.MapScale)
				event.Image.DrawImage(img, op)
			}
		}
	}
}

func (ui *UI) drawHoverTile(event *bento.Event) {
	x, y := ebiten.CursorPosition()
	if tile := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTile(ui.SelectedTileIndex); tile != nil {
		bounds := tile.Bounds()
		w, h := ui.MapScale*float64(bounds.Dx()), ui.MapScale*float64(bounds.Dy())
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
		op.GeoM.Scale(ui.MapScale, ui.MapScale)
		op.GeoM.Translate(math.Floor(float64(x)/w)*w, math.Floor(float64(y)/h)*h)
		event.Image.DrawImage(tile, op)
	} else {
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
		w := float64(ui.Tilemap.TileWidth) * ui.MapScale
		h := float64(ui.Tilemap.TileHeight) * ui.MapScale
		ui.Frame.Draw(
			event.Image,
			int(math.Floor(float64(x)/w)*w),
			int(math.Floor(float64(y)/h)*h),
			int(w),
			int(h),
			op)
	}
}

func (ui *UI) OnMapScroll(event *bento.Event) bool {
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

func (ui *UI) OnTilesetScroll(event *bento.Event) bool {
	_, sy := ebiten.Wheel()
	if sy != 0 {
		if sy > 0 {
			ui.TilesetScale *= 1.1
		} else {
			ui.TilesetScale /= 1.1
		}
	}
	return false
}

func (ui *UI) Hover(event *bento.Event) {
	x, y := ebiten.CursorPosition()
	tileX, tileY := ui.mapTilePos(x, y)
	if ui.SelectedTileset != "" && ui.SelectedTileIndex != 0 && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		z := math.MaxInt
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			z = 0
		}
		ui.Tilemap.SetTile(ui.SelectedTileset, ui.SelectedTileIndex, tileX, tileY, ebiten.IsKeyPressed(ebiten.KeyShift), z)
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		ui.Tilemap.EraseTile(tileX, tileY)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		if ui.DragX != 0 || ui.DragY != 0 {
			ui.OffsetX += float64(event.X-ui.DragX) / ui.MapScale
			ui.OffsetY += float64(event.Y-ui.DragY) / ui.MapScale
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
	ui.SelectedTileIndex = ui.Tilemap.Tilesets[ui.SelectedTileset].TileAt(
		int(float64(event.X)/ui.TilesetScale),
		int(float64(event.Y)/ui.TilesetScale))
}

func (ui *UI) DrawSelectedTiles(event *bento.Event) {
	rect := ui.Tilemap.Tilesets[ui.SelectedTileset].GetTileRect(ui.SelectedTileIndex)
	if rect == nil {
		return
	}
	op := new(ebiten.DrawImageOptions)
	op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
	ui.Frame.Draw(
		event.Image,
		int(float64(rect.Min.X)*ui.TilesetScale),
		int(float64(rect.Min.Y)*ui.TilesetScale),
		int(float64(rect.Dx())*ui.TilesetScale),
		int(float64(rect.Dy())*ui.TilesetScale),
		op)
}

func (ui *UI) mapTilePos(x, y int) (int, int) {
	w, h := float64(ui.Tilemap.TileWidth), float64(ui.Tilemap.TileHeight)
	ox, oy := math.Floor(ui.OffsetX/w), math.Floor(ui.OffsetY/h)
	return int(math.Floor(float64(x)/(w*ui.MapScale)) - ox), int(math.Floor(float64(y)/(h*ui.MapScale)) - oy)
}

func (ui *UI) UI() string {
	return `<col grow="1">
		<row grow="1">
			<col grow="1">
				<canvas grow="1" onDraw="Draw" onHover="Hover" onScroll="OnMapScroll" />
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
			<img onClick="SelectTile" onDraw="DrawSelectedTiles" onScroll="OnTilesetScroll" src="{{ .SelectedTileset }}" scale="{{ .TilesetScale }}" zIndex="100" />
		</col>
	</col>`
}
