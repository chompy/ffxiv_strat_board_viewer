package main

import (
	"encoding/json"
	"flag"
	"image/jpeg"
	"io"
	"os"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		panic(MissingInput)
	}

	input := flag.String("input", "", "strategy board share code")
	output := flag.String("output", "image", "format to output strategy board as (json, png, jpeg)")

	flag.Parse()

	// read input from stdin
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
	board, err := LoadBoard(strings.TrimSpace(*input))
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
	case "png":
		{
			image, err := DrawBoard(board)
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
			image, err := DrawBoard(board)
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
