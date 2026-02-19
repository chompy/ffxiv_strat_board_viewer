package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/webp"
)

const assetJsonPath = "data/assets.json"
const imageBasePath = "data/ffxiv-strategy-board-viewer"
const outputPath = "../../assets.zip"

var additionalImages = []string{
	"assets/objects/circle_aoe",
	"assets/background/1",
	"assets/background/2",
	"assets/background/3",
	"assets/background/4",
	"assets/background/5",
	"assets/background/6",
	"assets/background/7",
}
var images map[string]image.Image

type asset struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Image   string  `json:"image"`
	Offset  int     `json:"offset"`
	Scale   float64 `json:"scale"`
	Size    int     `json:"size"`
	Special bool    `json:"special"`
}

type outputAsset struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Scale float64 `json:"scale"`
}

func loadAssetList() ([]asset, error) {
	rawData, err := os.ReadFile(assetJsonPath)
	if err != nil {
		return nil, err
	}
	out := make([]asset, 0)
	return out, json.Unmarshal(rawData, &out)
}

func loadWebpImage(path string) (image.Image, error) {
	if images == nil {
		images = make(map[string]image.Image)
	}
	if images[path] != nil {
		return images[path], nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return webp.Decode(file)
}

func (a asset) loadImage() (image.Image, error) {
	assetImage, err := loadWebpImage(filepath.Join(imageBasePath, fmt.Sprintf("%s.webp", a.Image)))
	if err != nil {
		return nil, err
	}
	size := a.Size
	if a.Size == 0 {
		size = assetImage.Bounds().Size().X
	}
	return assetImage.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(a.Offset, 0, a.Offset+size, size)), nil
}

func writeToZip(zw *zip.Writer, name string, data io.Reader) error {
	zipEntry, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(zipEntry, data)
	return err
}

func main() {
	log.Println("Load asset data")
	assets, err := loadAssetList()
	if err != nil {
		panic(err)
	}

	zipFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	outputAssetList := make([]outputAsset, 0)

	for _, asset := range assets {
		log.Printf("Process asset %d (%s)", asset.ID, asset.Name)
		if asset.Image == "" {
			log.Println("  - No image, skipping")
			continue
		}
		log.Printf("  - Reading sprite from %s.webp as offset %d", asset.Image, asset.Offset)
		assetImage, err := asset.loadImage()
		if err != nil {
			panic(err)
		}
		var buf bytes.Buffer
		log.Printf("  - Save as %d.png", asset.ID)
		if err := png.Encode(&buf, assetImage); err != nil {
			panic(err)
		}
		if err := writeToZip(writer, fmt.Sprintf("%d.png", asset.ID), &buf); err != nil {
			panic(err)
		}
		outputAssetList = append(outputAssetList, outputAsset{
			ID:    asset.ID,
			Name:  asset.Name,
			Scale: asset.Scale,
		})
	}

	for _, imagePath := range additionalImages {
		log.Printf("Process image at %s.webp", imagePath)
		pathTo := filepath.Join(imageBasePath, fmt.Sprintf("%s.webp", imagePath))
		image, err := loadWebpImage(pathTo)
		if err != nil {
			panic(err)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, image); err != nil {
			panic(err)
		}
		name := fmt.Sprintf("x%s.png", filepath.Base(imagePath))
		log.Printf("  - Save as %s", name)
		if err := writeToZip(writer, name, &buf); err != nil {
			panic(err)
		}
	}

	log.Println("Save assets.json")
	outputAssetJson, err := json.Marshal(outputAssetList)
	if err != nil {
		panic(err)
	}
	if err := writeToZip(writer, "assets.json", bytes.NewReader(outputAssetJson)); err != nil {
		panic(err)
	}

}
