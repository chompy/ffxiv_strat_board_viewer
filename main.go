package main

import (
	"encoding/json"
	"flag"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		panic(MissingInput)
	}

	board := flag.String("board", "", "strategy board data")
	asImage := flag.String("image", "", "output path for strategy board image")
	asJson := flag.Bool("json", false, "output parsed strategy board as json")
	flag.Parse()

	data, err := DecodeStrategyBoard(*board)
	if err != nil {
		panic(err)
	}

	sb, err := ParseStrategyBoard(data)
	if err != nil {
		panic(err)
	}

	if *asJson {
		out, err := json.Marshal(sb)
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(out)
	}

	if *asImage != "" {
		image, err := DrawStrategyBoard(sb)
		if err != nil {
			panic(err)
		}
		image.SavePNG(*asImage)
	}

}
