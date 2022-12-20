package main

import (
	"image"
	"log"
	"math"
	"time"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Editor struct {
	Selection              *image.Rectangle
	Map                    *Map
	MapScale, TilesetScale float64
	OffsetX, OffsetY       float64
	Drag                   *[2]int
	HoverX, HoverY         int
	Frame                  *bento.NineSlice
	TileSelector           *TileSelector
}

func NewEditor() *Editor {
	tileset := NewTileset("tilesets")
	ui := &Editor{
		Map:          NewMap(16, 16, tileset),
		MapScale:     1,
		TilesetScale: 1,
		TileSelector: NewTileSelector(tileset),
	}
	img, _, err := ebitenutil.NewImageFromFile("ui/frame.png")
	if err != nil {
		log.Fatal(err)
	}
	ui.Frame = bento.NewNineSlice(img, [3]int{4, 24, 4}, [3]int{4, 24, 4}, 0, 0)
	return ui
}

func (ui *Editor) Draw(event *bento.Event) {
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		ui.drawHoverTile(event)
	}
	ui.drawMap(event)
	if !ebiten.IsKeyPressed(ebiten.KeyControl) {
		ui.drawHoverTile(event)
	}
	if ui.Selection != nil {
		ui.drawSelection(event)
	}
}

func (ui *Editor) drawMap(event *bento.Event) {
	w, h := float64(ui.Map.TileWidth), float64(ui.Map.TileHeight)
	ox, oy := math.Floor(ui.OffsetX/w)*w, math.Floor(ui.OffsetY/h)*h
	for x, ys := range ui.Map.Tilemap {
		for y, tiles := range ys {
			for _, tile := range tiles {
				img := ui.Map.Image(tile)
				if img == nil {
					log.Fatal(tile)
				}
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

func (ui *Editor) drawHoverTile(event *bento.Event) {
	if tile := ui.Map.Image(ui.TileSelector.Selected); tile != nil {
		bounds := tile.Bounds()
		w, h := ui.MapScale*float64(bounds.Dx()), ui.MapScale*float64(bounds.Dy())
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
		op.GeoM.Scale(ui.MapScale, ui.MapScale)
		op.GeoM.Translate(math.Floor(float64(event.X)/w)*w, math.Floor(float64(event.Y)/h)*h)
		event.Image.DrawImage(tile, op)
	} else {
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
		w := float64(ui.Map.TileWidth) * ui.MapScale
		h := float64(ui.Map.TileHeight) * ui.MapScale
		ui.Frame.Draw(
			event.Image,
			int(math.Floor(float64(event.X)/w)*w),
			int(math.Floor(float64(event.Y)/h)*h),
			int(w),
			int(h),
			op)
	}
}

func (ui *Editor) drawSelection(event *bento.Event) {
	if ui.Selection == nil || ui.Selection.Dx() == 0 || ui.Selection.Dy() == 0 {
		return
	}
	w, h := float64(ui.Map.TileWidth), float64(ui.Map.TileHeight)
	ox, oy := math.Floor(ui.OffsetX/w)*w, math.Floor(ui.OffsetY/h)*h
	op := new(ebiten.DrawImageOptions)
	op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
	ui.Frame.Draw(event.Image,
		int((float64(ui.Selection.Min.X)*w+ox)*ui.MapScale),
		int((float64(ui.Selection.Min.Y)*h+oy)*ui.MapScale),
		int(float64(ui.Selection.Dx())*w*ui.MapScale),
		int(float64(ui.Selection.Dy())*h*ui.MapScale),
		op)
}

func (ui *Editor) OnMapScroll(event *bento.Event) bool {
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

func (ui *Editor) OnTilesetScroll(event *bento.Event) bool {
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

func (ui *Editor) Click(event *bento.Event) {
	ui.HoverX, ui.HoverY = ui.mapTilePos(event.X, event.Y)
	if ui.TileSelector.Selected == nil {
		ui.Drag = &[2]int{ui.HoverX, ui.HoverY}
		selection := image.Rect(ui.HoverX, ui.HoverY, ui.HoverX, ui.HoverY)
		ui.Selection = &selection
	}
}

func (ui *Editor) Hover(event *bento.Event) {
	ui.HoverX, ui.HoverY = ui.mapTilePos(event.X, event.Y)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.TileSelector.Selected = nil
		ui.Selection = nil
	} else if ui.Selection != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyG) {
			ui.Map.Generate(*ui.Selection, time.Now().UnixMilli())
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			ui.Map.Erase(*ui.Selection)
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		ui.OffsetY++
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		ui.OffsetY--
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		ui.OffsetX++
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		ui.OffsetX--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		game.SetScene(NewExplore(ui.Map))
	}

	tileX, tileY := ui.mapTilePos(event.X, event.Y)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		ui.Map.EraseTile(tileX, tileY)
	} else if ui.TileSelector.Selected == nil && ui.Drag != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		selection := image.Rect(ui.Drag[0], ui.Drag[1], tileX+1, tileY+1)
		ui.Selection = &selection
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		ui.Drag = nil
		if ui.Selection != nil && ui.Selection.Dx() == 1 && ui.Selection.Dy() == 1 {
			ui.Selection = nil
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		z := math.MaxInt
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			z = 0
		}
		ui.Map.Tilemap.Set(ui.TileSelector.Selected, tileX, tileY, ebiten.IsKeyPressed(ebiten.KeyShift), z)
		ui.Map.Save("map.json")
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		if ui.Drag != nil {
			ui.OffsetX += float64(event.X-ui.Drag[0]) / ui.MapScale
			ui.OffsetY += float64(event.Y-ui.Drag[1]) / ui.MapScale
		}
		ui.Drag = &[2]int{event.X, event.Y}
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		ui.Drag = nil
	}
}

func (ui *Editor) DrawSelectedTiles(event *bento.Event) {
	if ui.TileSelector.Selected == nil {
		return
	}
	rect := ui.Map.Spritesheets[ui.TileSelector.Selected.Spritesheet].Rect(ui.TileSelector.Selected.Index)
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

func (ui *Editor) mapTilePos(x, y int) (int, int) {
	w, h := float64(ui.Map.TileWidth), float64(ui.Map.TileHeight)
	ox, oy := math.Floor(ui.OffsetX/w), math.Floor(ui.OffsetY/h)
	return int(math.Floor(float64(x)/(w*ui.MapScale)) - ox), int(math.Floor(float64(y)/(h*ui.MapScale)) - oy)
}

func (ui *Editor) UI() string {
	return `<col grow="1">
		<row grow="1">
			<col grow="1">
				<canvas grow="1" onDraw="Draw" onClick="Click" onHover="Hover" onScroll="OnMapScroll" />
			</col>
		</row>
		<col float="true" justifySelf="start end" margin="16px">
			{{ if ne .TileSelector.Selected nil }}
				<text font="RobotoMono 14" color="#ffffff">{{ .TileSelector.Selected.Spritesheet }} {{ .TileSelector.Selected.Index }}</text>
			{{ end }}
			<text font="RobotoMono 14" color="#ffffff">{{ .HoverX }}, {{ .HoverY }}</text>
		</col>
		<col float="true" justifySelf="end" margin="16px">
			<TileSelector zIndex="100" />
		</col>
	</col>`
}
