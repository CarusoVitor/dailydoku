package dailydoku

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

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

func (cs characteristicSquare) String() string {
	return fmt.Sprintf("%s(%s)", cs.Type, cs.Obj)
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

func (pg PokeDokuDailyResponse) getRow() []characteristicSquare {
	return []characteristicSquare{pg.X1, pg.X2, pg.X3}
}

func (pg PokeDokuDailyResponse) getColumn() []characteristicSquare {
	return []characteristicSquare{pg.Y1, pg.Y2, pg.Y3}
}

func formatCharacteristicSquares(squareA, squareB characteristicSquare) string {
	return fmt.Sprintf("%s,%s", squareA, squareB)
}

func solveGrid(pokedokuGrid PokeDokuDailyResponse) map[string][]string {
	row := pokedokuGrid.getRow()
	column := pokedokuGrid.getColumn()
	solutions := make(map[string][]string, len(row)*len(column))

	for _, xsquare := range row {
		for _, ysquare := range column {
			pokemons, err := solveForTwo(xsquare, ysquare)
			if err != nil {
				slog.Error("error solving for two", "err", err)
				pokemons = nil
			}
			key := formatCharacteristicSquares(xsquare, ysquare)
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

// Solve uses dokuex's Match to solve the daily PokeDoku's puzzle.
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

	if numToDisplay > 1 {
		printManySolutions(solution, numToDisplay)
	} else if numToDisplay == 1 {
		printOneSolution(solution, dailyPuzzle)
	} else {
		return fmt.Errorf("invalid number of pokemons to display: %d", numToDisplay)
	}

	return nil
}

func printManySolutions(solution map[string][]string, numToDisplay int) {
	for combination, pokemons := range solution {
		fmt.Printf("%s: ", combination)
		if pokemons == nil {
			fmt.Printf("unable to match this combination of characteristics :(\n")
		} else {
			toDisplay := min(len(pokemons), numToDisplay)
			fmt.Printf("%v\n", pokemons[:toDisplay])
		}
	}
}

// printOneSolution displays a random solution grid formatted on a table style.
// The randomness come from dokuex's Match using sets, so its return value is
// non deterministic,
func printOneSolution(solution map[string][]string, dailyPuzzle PokeDokuDailyResponse) {
	// margin := 0
	row := dailyPuzzle.getRow()
	column := dailyPuzzle.getColumn()

	var header strings.Builder
	header.WriteByte(' ')
	for _, square := range row {
		header.WriteString(fmt.Sprintf("\t%s", square))
	}
	header.WriteString("\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprint(w, header.String())

	for _, ysquare := range column {
		fmt.Fprintf(w, "%s", ysquare)
		for _, xsquare := range row {
			combinationKey := formatCharacteristicSquares(xsquare, ysquare)
			pokemons := solution[combinationKey]

			pokemon := "-"
			if len(pokemons) > 0 {
				pokemon = pokemons[0]
			}
			fmt.Fprintf(w, "\t%s", pokemon)
		}
		fmt.Fprint(w, "\n")
	}
	w.Flush()
}
