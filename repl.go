package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/ashme5280/pokedex/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error // Updated signature
}

type config struct {
	commands 		map[string]cliCommand
	nextURL  		*string
	prevURL  		*string
	cache    		pokecache.Cache
	caughtPokemon	map[string]Pokemon
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
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

type RespLocationDetail struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
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
		"explore": {
    		name:        "explore <area_name>",
      		description: "List the pokemons in a location area",
       		callback:    callbackExplore,
		},
		"catch": {
    		name:        "catch <pokemon_name>",
      		description: "Attempt to catch a pokemon",
       		callback:    callbackCatch,
		},
		"inspect": {
			name:        "inspect <pokemon_name>",
			description: "View details about a caught pokemon",
			callback:    callbackInspect,
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
		args := []string{}
		if len(words) > 1 {
			args = words [1:]
		}

		if cmd, exists := cfg.commands[commandName]; exists {
					err := cmd.callback(cfg, args) // Pass args here!
					if err != nil {
						fmt.Println(err)
					}
				} else {
					fmt.Println("Unknown command")
				}
	}
}

func callbackHelp(cfg *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range cfg.commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println()
	return nil
}

func callbackExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func cleanInput(text string) []string {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	return words
}

func callbackMap(cfg *config, args []string) error {
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

func callbackMapb(cfg *config, args []string) error {
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

func callbackExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a location area name")
	}

	areaName := args[0]
	url := "https://pokeapi.co/api/v2/location-area/" + areaName

	fmt.Printf("Exploring %s...\n", areaName)

	// 1. Check Cache First
	if val, ok := cfg.cache.Get(url); ok {
		var locationDetail RespLocationDetail
		if err := json.Unmarshal(val, &locationDetail); err != nil {
			return err
		}

		fmt.Println("Found Pokemon:")
		for _, encounter := range locationDetail.PokemonEncounters {
			fmt.Printf(" - %s\n", encounter.Pokemon.Name)
		}
		return nil
	}

	// 2. Fetch from API on cache miss
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("location area '%s' not found", areaName)
	}

	dat, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// 3. Save to Cache
	cfg.cache.Add(url, dat)

	// 4. Decode and Print
	var locationDetail RespLocationDetail
	if err := json.Unmarshal(dat, &locationDetail); err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range locationDetail.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func callbackCatch(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a pokemon name")
	}

	pokemonName := args[0]
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemonName

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	var dat []byte

	// 1. Check cache or fetch from API
	if val, ok := cfg.cache.Get(url); ok {
		dat = val
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return fmt.Errorf("pokemon '%s' not found", pokemonName)
		}

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		dat = bodyBytes
		cfg.cache.Add(url, dat)
	}

	// 2. Decode response
	var pokemon Pokemon
	if err := json.Unmarshal(dat, &pokemon); err != nil {
		return err
	}

	// 3. Determine catch chance using base_experience
	resVal := rand.Intn(pokemon.BaseExperience)
	if resVal > 50 {
		fmt.Printf("%s escaped!\n", pokemon.Name)
		return nil
	}

	fmt.Printf("%s was caught!\n", pokemon.Name)

	// 4. Save to user's Pokedex
	cfg.caughtPokemon[pokemon.Name] = pokemon

	return nil
}

func callbackInspect(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a pokemon name")
	}

	pokemonName := args[0]
	pokemon, exists := cfg.caughtPokemon[pokemonName]
	if !exists {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}
