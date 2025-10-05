package dailydoku

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/CarusoVitor/dokuex/characteristics"
)

const pokedokuApiDaily = "https://api.pokedoku.com/api/puzzle/current"

var characteristicNameMap = map[string]string{
	"POKEMON_TYPE":       "type",
	"GENERATION":         "generation",
	"POKEMON_MOVE":       "move",
	"POKEMON_ABILITY":    "ability",
	"BABY":               "baby",
	"MYTHICAL":           "mythical",
	"LEGENDARY":          "legendary",
	"LEGENDARY_TRIO":     "legendary_trio",
	"ULTRA_BEAST":        "ultra-beast",
	"MEGA":               "mega",
	"DUAL_TYPE":          "dual-type",
	"MONOTYPE":           "monotype",
	"GMAX":               "gmax",
	"EVOLVED_BY":         "evolved-by",
	"EVOLUTION_POSITION": "evolution-position",
	"EVOLUTION_BRANCHED": "evolution-branched",
	"HISUI":              "hisui",
	"FIRST_PARTNER":      "first-partner",
	"FOSSIL":             "fossil",
	"PARADOX":            "paradox",
}

type characteristicSquare struct {
	Type          string   `json:"type"`
	Obj           string   `json:"obj"`
	ExcludedForms []string `json:"excludedForms"`
}

type PokeDokuDailyResponse struct {
	X1   characteristicSquare `json:"x1"`
	X2   characteristicSquare `json:"x2"`
	X3   characteristicSquare `json:"x3"`
	Y1   characteristicSquare `json:"y1"`
	Y2   characteristicSquare `json:"y2"`
	Y3   characteristicSquare `json:"y3"`
	Date string               `json:"date"`
}

func solveGrid(pokedokuGrid PokeDokuDailyResponse) map[string][]string {
	row := []characteristicSquare{
		pokedokuGrid.X1,
		pokedokuGrid.X2,
		pokedokuGrid.X3,
	}
	column := []characteristicSquare{
		pokedokuGrid.Y1,
		pokedokuGrid.Y2,
		pokedokuGrid.Y3,
	}
	solutions := make(map[string][]string, len(row)*len(column))
	for _, xsquare := range row {
		for _, ysquare := range column {
			pokemons, err := solveForTwo(xsquare, ysquare)
			if err != nil {
				slog.Error("error solving for two", "err", err)
				pokemons = nil
			}
			key := fmt.Sprintf("%s(%s),%s(%s)", xsquare.Type, xsquare.Obj, ysquare.Type, ysquare.Obj)
			solutions[key] = pokemons
		}
	}
	return solutions
}

func formatCharacteristicsToValues(squareA, squareB characteristicSquare) map[string][]string {
	if squareA.Type == squareB.Type {
		return map[string][]string{
			characteristicNameMap[squareA.Type]: {squareA.Obj, squareB.Obj},
		}
	}
	return map[string][]string{
		characteristicNameMap[squareA.Type]: {squareA.Obj},
		characteristicNameMap[squareB.Type]: {squareB.Obj},
	}
}

func solveForTwo(squareA, squareB characteristicSquare) ([]string, error) {
	characteristicToValues := formatCharacteristicsToValues(squareA, squareB)
	slog.Debug("matching with", "values", characteristicToValues)
	pokemons, err := characteristics.Match(characteristicToValues)
	return pokemons, formatMatchError(err)
}

func formatMatchError(err error) error {
	var invalidErr characteristics.InvalidCharacteristicError
	if err != nil {
		if errors.As(errors.Unwrap(err), &invalidErr) {
			return fmt.Errorf("Characteristic %s was not yet implemented", invalidErr.Name)
		}
		return fmt.Errorf("Unknown error when querying pokemons")
	}
	return nil
}

func Solve(client http.Client, numToDisplay int) error {
	resp, err := client.Get(pokedokuApiDaily)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dailyPuzzle PokeDokuDailyResponse
	json.Unmarshal(body, &dailyPuzzle)

	solution := solveGrid(dailyPuzzle)

	for combination, pokemons := range solution {
		fmt.Printf("%s: ", combination)
		if pokemons == nil {
			fmt.Printf("unable to match this combination of characteristics :(\n")
		} else {
			toDisplay := min(len(pokemons), numToDisplay)
			fmt.Printf("%v\n", pokemons[:toDisplay])
		}
	}

	return nil
}
