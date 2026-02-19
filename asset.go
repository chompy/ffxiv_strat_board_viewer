package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"strings"

	"golang.org/x/image/webp"
)

const assetJsonPath = "assets/assets.json"
const objectImageCachePath = "assets/objects/_cache"
const backgroundImagePath = "assets/backgrounds"
const defaultObjectScale = 1.0 / 200.0

type StrategyBoardAsset struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Image   string  `json:"image"`
	Offset  int     `json:"offset"`
	Scale   float64 `json:"scale"`
	Size    int     `json:"size"`
	Special bool    `json:"special"`
}

type StrategyBoardAssets []StrategyBoardAsset

var assets StrategyBoardAssets = nil
var images map[string]image.Image = nil

func LoadStrategyBoardAssets() (StrategyBoardAssets, error) {
	if assets != nil {
		return assets, nil
	}
	data, err := os.ReadFile(assetJsonPath)
	if err != nil {
		return nil, err
	}
	assets := make(StrategyBoardAssets, 0)
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, err
	}
	for i := range assets {
		if assets[i].Scale == 0 {
			assets[i].Scale = defaultObjectScale
		}
	}
	return assets, nil
}

func (a StrategyBoardAssets) GetByID(id int) (*StrategyBoardAsset, error) {
	for _, asset := range a {
		if asset.ID == id {
			return &asset, nil
		}
	}
	return nil, AssetNotFound
}

func (a StrategyBoardAsset) GetScale() float64 {
	if a.Scale > 0 {
		return a.Scale
	}
	return defaultObjectScale
}

func (a *StrategyBoardAsset) LoadImage() (image.Image, error) {
	// attempt to load previously extracted asset image
	cachePath := path.Join(objectImageCachePath, fmt.Sprintf("%d.png", a.ID))
	img, err := loadImage(cachePath)
	if err == nil {
		return img, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	// load spritesheet image to extract desired asset image
	img, err = loadImage(a.Image + ".webp")
	if err != nil {
		return nil, err
	}

	size := a.Size
	if a.Size == 0 {
		size = img.Bounds().Size().X
	}

	subImage := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(a.Offset, 0, a.Offset+size, size))

	// store extracted image in cache
	os.MkdirAll(objectImageCachePath, 0777)
	cacheFile, err := os.Create(cachePath)
	if err != nil {
		return nil, err
	}
	defer cacheFile.Close()
	if err := png.Encode(cacheFile, subImage); err != nil {
		return nil, err
	}

	// load extracted image from cache
	return loadImage(cachePath)
}

func loadImage(path string) (image.Image, error) {
	if images != nil && images[path] != nil {
		return images[path], nil
	}
	if images == nil {
		images = make(map[string]image.Image)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if strings.HasSuffix(path, ".webp") {
		images[path], err = webp.Decode(file)
		return images[path], err
	}

	images[path], err = png.Decode(file)
	return images[path], err
}

func loadBackgroundImage(id int) (image.Image, error) {
	pathTo := path.Join(backgroundImagePath, fmt.Sprintf("%d.webp", id+1))
	return loadImage(pathTo)
}
