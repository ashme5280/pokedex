package main

import (
	"time"

	"github.com/ashme5280/pokedex/internal/pokecache"
)

func main() {
	cfg := &config{
		cache:    pokecache.NewCache(5 * time.Minute),
		commands: getCommands(),
	}

	startRepl(cfg)
}
