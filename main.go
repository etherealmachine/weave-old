package main

import (
	"log"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"

	_ "image/png"
)

type Game struct {
	ui *bento.Box
}

func NewGame() *Game {
	ui, err := bento.Build(&UI{})
	if err != nil {
		log.Fatal(err)
	}
	return &Game{
		ui: ui,
	}
}

func (g *Game) Update() error {
	if err := g.ui.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()
	g.ui.Draw(screen)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

func main() {
	log.SetFlags(log.Lshortfile)
	ebiten.SetWindowSize(1024, 800)
	ebiten.SetWindowTitle("weave")
	ebiten.SetFullscreen(false)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
