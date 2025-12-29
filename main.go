package main

import (
	"github.com/stevenzieske/SeaPower-Calculator/configs"
	"github.com/stevenzieske/SeaPower-Calculator/internal/data"
	"github.com/stevenzieske/SeaPower-Calculator/internal/tui"
)

var cfg = configs.Load()

func main() {
	weapons := data.LoadMissileData()
	// log.Println(weapons)

	tui.RenderTUI(weapons)
}
