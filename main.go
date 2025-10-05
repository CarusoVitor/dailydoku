package main

import (
	"flag"
	"log/slog"
	"net/http"
	"time"

	dailydoku "github.com/CarusoVitor/dailydoku/solver"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelError)
	client := http.Client{Timeout: time.Second * 30}

	nPtr := flag.Int("n", 1, "number of pokemons")
	flag.Parse()

	dailydoku.Solve(client, *nPtr)
}
