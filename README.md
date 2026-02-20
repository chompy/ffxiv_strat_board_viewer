# Golang FF14 Strategy Board Viewer

Golang library for parsing Final Fantasy XIV strategy board data and rendering it as an image.

This is more or less a Golang port of this Typescript strategy board viewer:
https://github.com/Ennea/ffxiv-strategy-board-viewer


## CLI Usage

Basic Usage:
```
echo "STRATEGY BOARD SHARE CODE" | go run cli/main.go > out.png
```

The included CLI takes a strategy board share code through STDIN and will output a PNG image through STDOUT.