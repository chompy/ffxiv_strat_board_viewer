package main

import (
	"encoding/json"
	"flag"
	"image/jpeg"
	"io"
	"os"
	"strings"

	strategy_board "github.com/chompy/ffxiv_strat_board_viewer"
)

func main() {

	// parse input args
	input := flag.String("input", "", "strategy board share code")
	output := flag.String("output", "image", "format to output strategy board as (json, png, jpeg)")
	flag.Parse()
	if *input == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			rawInput, err := io.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
			*input = string(rawInput)
		}
	}

	// load board
	board, err := strategy_board.Load(strings.TrimSpace(*input))
	if err != nil {
		panic(err)
	}

	switch *output {
	case "json":
		{
			out, err := json.Marshal(board)
			if err != nil {
				panic(err)
			}
			os.Stdout.Write(out)
			break
		}
	case "image":
	case "png":
		{
			image, err := strategy_board.Draw(board)
			if err != nil {
				panic(err)
			}
			if err := image.EncodePNG(os.Stdout); err != nil {
				panic(err)
			}
			break
		}
	case "jpeg":
	case "jpg":
		{
			image, err := strategy_board.Draw(board)
			if err != nil {
				panic(err)
			}
			if err := jpeg.Encode(os.Stdout, image.Image(), nil); err != nil {
				panic(err)
			}
			break
		}
	}

}
