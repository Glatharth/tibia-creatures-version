package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Creatures struct {
	Name        string `json:"name"`
	Implemented string `json:"implemented"`
	ActualName  string `json:"actualname"`
}

var creatures []Creatures
var index map[string]int

func ReadCreatures(file string) {
	byteValue, _ := os.ReadFile(file)
	var creaturesTmp []Creatures
	err := json.Unmarshal(byteValue, &creaturesTmp)
	if err != nil {
		return
	}
	creatures = append(creatures, creaturesTmp...)
}

func Init() {
	ReadCreatures("npc.json")
	ReadCreatures("monster.json")

	index = make(map[string]int)
	for _, creature := range creatures {
		if creature.Implemented == "--" || creature.Implemented == "" {
			continue
		}
		rx, err := regexp.Compile("\\d+\\.\\d+")
		if err != nil {
			panic(err)
		}
		version, err := strconv.Atoi(strings.Replace(rx.FindString(creature.Implemented), ".", "", -1))
		if err != nil {
			fmt.Println(creature.Name)
			panic(err)
		}
		if version < 100 {
			version = version * 10
		}
		if creature.ActualName != "" {
			index[strings.ToLower(creature.ActualName)] = version
			continue
		}
		index[strings.ToLower(creature.Name)] = version
	}
}

func main() {
	Init()
	err := filepath.Walk("creatures", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err)
		}
		rx, err := regexp.Compile(`(?:Game.createMonsterType|internalNpcName).*"(.+)"`)
		creatureName := rx.FindStringSubmatch(string(f))
		if len(creatureName) < 2 {
			fmt.Printf("[Error] Creature %s not found creature name\n", path)
			return nil
		}
		creatureName[1] = strings.ToLower(creatureName[1])
		if index[creatureName[1]] == 0 {
			fmt.Printf("[Warning] Creature %s not found in TibiaFandom or not found implemented field || path: %s\n", creatureName[1], path)
			return nil
		}

		data := []byte(fmt.Sprintf("if CLIENT_VERSION < %s then\n\treturn\nend\n\n", strconv.Itoa(index[creatureName[1]])))
		data = append(data, f...)
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			panic(err)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
