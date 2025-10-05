# dailydoku
DailyDoku uses [dokuex](https://github.com/CarusoVitor/dokuex) to solve the daily [PokeDoku](https://www.pokedoku.com/) puzzle.

## Installation
```
git clone https://github.com/CarusoVitor/dailydoku
cd dokuex
go build -o dailydoku .
```
## Usage
Dailydoku has an optional argument *n* which specifies the number of pokemons to display in each row-column pair. 
```
./dailydoku --n value
```

If no argument is provided, it defaults to 1, which displays the daily solution in a tabular fashion. Each call displays a different solution grid.

If the characteristic is still not implemented in dokuex, an error message will be displayed and the respective squares will be filled with `-` (or an error message will be shown if n > 1).