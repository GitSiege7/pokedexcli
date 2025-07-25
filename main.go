package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	pokecache "github.com/GitSiege7/pokedexcli/internal"
)

var commands map[string]cliCommand
var pokemon map[string]Pokemon

func init() {
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Display a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays map and pages map to next page",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Pages map to previous page",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Takes location area name from map and retrieves list of pokemon in the area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Takes pokemon name as parameter and attempts to catch and add to pokedex",
			callback:    commandCatch,
		},
	}

	pokemon = make(map[string]Pokemon)
}

func cleanInput(text string) []string {
	words := strings.Fields(text)

	var clean []string
	for _, word := range words {
		clean = append(clean, strings.ToLower(strings.TrimSpace(word)))
	}

	return clean
}

func commandExit(c *config, params []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(c *config, params []string) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	fmt.Println()
	for key := range commands {
		fmt.Printf("%s: %s\n", commands[key].name, commands[key].description)
	}
	return nil
}

func commandMap(c *config, params []string) error {
	baseUrl := "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"

	var url string

	if c.Next == "" {
		url = baseUrl
	} else {
		url = c.Next
	}

	var data LocationRes

	err := getData(url, &data, c)
	if err != nil {
		return err
	}

	c.Next = data.Next
	c.Prev = data.Previous

	for _, result := range data.Results {
		fmt.Println(result["name"])
	}

	return nil
}

func commandMapBack(c *config, params []string) error {
	if c.Prev == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	url := c.Prev

	var data LocationRes

	err := getData(url, &data, c)
	if err != nil {
		return err
	}

	c.Next = data.Next
	c.Prev = data.Previous

	for _, result := range data.Results {
		fmt.Println(result["name"])
	}

	return nil
}

// helper function to get data, unmarshal, store in data struct, return status
func getData(url string, data any, c *config) error {
	var body []byte
	var res *http.Response
	var err error

	if val, ok := c.Cache.Get(url); !ok {
		res, err = http.Get(url)

		if err != nil {
			return fmt.Errorf("failed to get data")
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body")
		}

		c.Cache.Add(url, body)

		//fmt.Println("Didn't find cache for: ", url)

	} else {
		body = val

		//fmt.Println("Found cache for: ", url)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json data")
	}

	return nil
}

func commandExplore(c *config, params []string) error {
	if len(params) > 1 {
		return fmt.Errorf("excess parameters: expected 1, got %v", len(params))
	}

	fmt.Println("Exploring " + strings.ToLower(params[0]) + "...")

	url := "https://pokeapi.co/api/v2/location-area/" + params[0]

	var data EncounterRes

	err := getData(url, &data, c)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")

	for _, encounter := range data.Encounters {
		fmt.Println(" - " + encounter.Pokemon.Name)
	}

	return nil
}

func catch(data PokemonRes) {
	if _, ok := pokemon[data.Name]; !ok {
		pokemon[data.Name] = Pokemon{Name: data.Name, url: }
	}
}

func commandCatch(c *config, params []string) error {
	if len(params) > 1 {
		return fmt.Errorf("excess parameters: expected 1, got %v", len(params))
	}

	url := "https://pokeapi.co/api/v2/pokemon/" + params[0]

	var data PokemonRes

	err := getData(url, &data, c)
	if err != nil {
		return err
	}

	fmt.Println("Throwing a Pokeball at", data.Name, "...")

	chance := float64(1.0 - (float64(data.Base_xp) / float64(310)))

	outcome := rand.Float64()

	if outcome < chance {
		fmt.Println(data.Name, "was caught!")
		catch(&data)
	} else {
		fmt.Println(data.Name, "escaped!")
	}

	return nil
}

func main() {
	init()

	scanner := bufio.NewScanner(os.Stdin)

	cache := pokecache.NewCache(5 * time.Second)

	c := config{
		Next:  "",
		Prev:  "",
		Cache: cache,
	}

	for {
		fmt.Print("Pokedex > ")

		scanner.Scan()
		clean := cleanInput(scanner.Text())

		cmd, ok := commands[clean[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		var err error

		if len(clean) > 1 {
			err = cmd.callback(&c, clean[1:])
		} else {
			err = cmd.callback(&c, nil)
		}

		if err != nil {
			fmt.Println(err)
		}

	}
}
