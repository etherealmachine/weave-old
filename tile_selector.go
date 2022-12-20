package main

import (
	"strconv"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
)

type TileSelector struct {
	Selected *Tile
	Tileset  *Tileset
}

func NewTileSelector(tileset *Tileset) *TileSelector {
	return &TileSelector{Tileset: tileset}
}

func (ui *TileSelector) Draw(event *bento.Event) {
	i, err := strconv.Atoi(event.Box.Attrs["index"])
	if err != nil {
		return
	}
	op := new(ebiten.DrawImageOptions)
	op.GeoM.Translate(float64(event.Box.X), float64(event.Box.Y))
	event.Image.DrawImage(ui.Tileset.Tiles()[i].Image, op)
}

func (ui *TileSelector) Click(event *bento.Event) {

}

func (ui *TileSelector) Hover(event *bento.Event) {

}

func (ui *TileSelector) UI() string {
	return `<row grow="1">
		<row grow="1" justify="between" margin="0 0 12px 0">
			{{ range $index, $img := .Tileset.Tiles }}
				<canvas width="16" height="16" onDraw="Draw" onClick="Click" onHover="Hover" index="{{ $index }}" />
			{{ end }}
			{{ range $name, $sheet := .Tileset.Spritesheets }}
				<button
						font="NotoSans 18"
						btn="ui/button.png 6"
						color="#ffffff"
						padding="12px"
						onClick="SelectTileset"
				>{{ $name }}</button>
			{{ end }}
		</row>
	</row>`
}
