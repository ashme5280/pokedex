package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ashme5280/pokedex/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	commands map[string]cliCommand
	nextURL  *string
	prevURL  *string
	cache    pokecache.Cache
}

type RespLocations struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    callbackHelp,
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
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    callbackExit,
		},
	}
}

func startRepl(cfg *config) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}

		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}

		commandName := words[0]

		if cmd, exists := cfg.commands[commandName]; exists {
			err := cmd.callback(cfg)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func callbackHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range cfg.commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println()
	return nil
}

func callbackExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func cleanInput(text string) []string {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	return words
}

func callbackMap(cfg *config) error {
	url := "https://pokeapi.co/api/v2/location-area"
	if cfg.nextURL != nil {
		url = *cfg.nextURL
	}

	// 1. Check the cache first
	if val, ok := cfg.cache.Get(url); ok {
		var locationResp RespLocations
		if err := json.Unmarshal(val, &locationResp); err != nil {
			return err
		}
		cfg.nextURL = locationResp.Next
		cfg.prevURL = locationResp.Previous

		for _, loc := range locationResp.Results {
			fmt.Println(loc.Name)
		}
		return nil
	}

	// 2. If miss, fetch from API
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	dat, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// 3. Save to cache
	cfg.cache.Add(url, dat)

	// 4. Decode and print
	var locationResp RespLocations
	if err := json.Unmarshal(dat, &locationResp); err != nil {
		return err
	}

	cfg.nextURL = locationResp.Next
	cfg.prevURL = locationResp.Previous

	for _, loc := range locationResp.Results {
		fmt.Println(loc.Name)
	}

	return nil
}

func callbackMapb(cfg *config) error {
	if cfg.prevURL == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	url := *cfg.prevURL

	// 1. Check cache first for mapb
	if val, ok := cfg.cache.Get(url); ok {
		var locationResp RespLocations
		if err := json.Unmarshal(val, &locationResp); err != nil {
			return err
		}
		cfg.nextURL = locationResp.Next
		cfg.prevURL = locationResp.Previous

		for _, loc := range locationResp.Results {
			fmt.Println(loc.Name)
		}
		return nil
	}

	// 2. If miss, fetch from API
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	dat, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// 3. Save to cache
	cfg.cache.Add(url, dat)

	// 4. Decode and print
	var locationResp RespLocations
	if err := json.Unmarshal(dat, &locationResp); err != nil {
		return err
	}

	cfg.nextURL = locationResp.Next
	cfg.prevURL = locationResp.Previous

	for _, loc := range locationResp.Results {
		fmt.Println(loc.Name)
	}

	return nil
}
