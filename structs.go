package main

import (
	pokecache "github.com/GitSiege7/pokedexcli/internal"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	Next  string
	Prev  string
	Cache *pokecache.Cache
}

type LocationRes struct {
	Count    int                 `json:"count"`
	Next     string              `json:"next"`
	Previous string              `json:"previous"`
	Results  []map[string]string `json:"results"`
}

type EncounterRes struct {
	Encounters []Encounter `json:"pokemon_encounters"`
}

type Encounter struct {
	Pokemon Pokemon `json:"pokemon"`
}

type Pokemon struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PokemonRes struct {
	Name    string `json:"name"`
	Base_xp int    `json:"base_experience"`
}
