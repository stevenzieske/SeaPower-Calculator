package main

import (
	"seapower_calculator/configs"
	"seapower_calculator/internal/data"
	"seapower_calculator/internal/tui"
)

var cfg = configs.Load()

func main() {
	weapons := data.LoadMissileData()
	// log.Println(weapons)

	tui.RenderTUI(weapons)
}
