package main

import (
	"time"

	"github.com/ashme5280/pokedex/internal/pokecache" // Added missing import
)

func main() {
    cfg := &config{
    	cache: pokecache.NewCache(5 * time.Minute),
    	commands: map[string]cliCommand{
            "help": {
                name:        "help",
                description: "Displays a help message",
                callback:    callbackHelp,
            },
            "exit": {
                name:        "exit",
                description: "Exit the Pokedex",
                callback:    callbackExit,
            },
            "map": {
                name:        "map",
                description: "Displays the next 20 location areas",
                callback:    callbackMap,
            },
            "mapb": {
                name:        "mapb",
                description: "Displays the previous 20 location areas",
                callback:    callbackMapb,
            },
        },
    }

    startRepl(cfg)
}
