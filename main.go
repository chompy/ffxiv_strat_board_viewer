package main

func main() {
	data, err := DecodeStrategyBoard("[stgy:aAOewrSt9KJHb0X75sgGGEXUm9KNCRrzbowlYQJfLqIxC71yae9nmWBiWgA866HAHo6dY6mImz1x-JcFQOJCL-jzHFL+L4lY9HYyleFjKjR8jsxP50c-sUo2K2NFTKl+rkgO8BBquoA5uxzu91nbmlMlXPrYRdMpaP0okpMXM9m6vN7pFPydttBMT4mq-bqDqP6GTGQtZGY4UlOrrTr6o7H0jmMiBN2vCoccRJXxcPqCm]")
	if err != nil {
		panic(err)
	}

	sb, err := ParseStrategyBoard(data)
	if err != nil {
		panic(err)
	}

	image, err := DrawStrategyBoard(sb)
	if err != nil {
		panic(err)
	}
	image.SavePNG("out.png")

}
