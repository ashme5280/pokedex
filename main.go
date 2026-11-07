package main

func main() {
    cfg := &config{
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
