package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"net/http"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error // Fixed typo (removed 'chno ')
}

type config struct {
	commands map[string]cliCommand
	nextURL	*string
	prevURL *string
}

type RespLocations struct {
	Count int `json:"count"`
	Next *string `json:"next"`
	Previous *string `json:"previous"`
	Results []struct {
		Name string `json:"name"` // Must be lowercase "name" to match PokeAPI JSON
		URL string `json:"url"`
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

		words := cleanInput(scanner.Text()) // Added missing parentheses ()
		if len(words) == 0 {
			continue
		}

		commandName := words[0]

		// Look up command in registry from the config struct
		if cmd, exists := cfg.commands[commandName]; exists { // Changed commands -> cfg.commands
			err := cmd.callback(cfg) // Pass cfg into the callback
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

	// 1. Make the GET request
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 2. Decode the JSON response
	var locationResp RespLocations
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&locationResp); err != nil {
		return err
	}

	// 3. Update pagination URLs in config
	cfg.nextURL = locationResp.Next
	cfg.prevURL = locationResp.Previous

	// 4. Print the location area names
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

	res, err := http.Get(*cfg.prevURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var locationResp RespLocations
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&locationResp); err != nil {
		return err
	}

	cfg.nextURL = locationResp.Next
	cfg.prevURL = locationResp.Previous

	for _, loc := range locationResp.Results {
		fmt.Println(loc.Name)
	}

	return nil
}
