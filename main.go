package main

import (
	"seapower_calculator/configs"
	"seapower_calculator/internal/data"
	"seapower_calculator/internal/tui"
)

var cfg = configs.Load()

func main() {
	missiles := data.LoadMissileData()
	// log.Println(missiles)

	tui.RenderTUI(missiles)
}
