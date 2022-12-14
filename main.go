package main

import (
	"log"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"

	_ "image/png"
)

var game *Game

type Game struct {
	scene *bento.Box
}

func (g *Game) SetScene(scene bento.Component) {
	ui, err := bento.Build(scene)
	if err != nil {
		log.Fatal(err)
	}
	g.scene = ui
}

func (g *Game) Update() error {
	if err := g.scene.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()
	g.scene.Draw(screen)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

func main() {
	log.SetFlags(log.Lshortfile)
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("weave")
	//ebiten.SetFullscreen(true)
	game = &Game{}
	game.SetScene(NewEditor())
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
