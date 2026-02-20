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
	"os/exec"
	"path/filepath"

	"golang.org/x/image/webp"
)

const assetSourcePath = "assets"
const assetSourceRepo = "https://github.com/Ennea/ffxiv-strategy-board-viewer"
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
var additionalFiles = []string{
	"Roboto-Medium.ttf",
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
	assetImage, err := loadWebpImage(filepath.Join(assetSourcePath, filepath.Base(assetSourceRepo), fmt.Sprintf("%s.webp", a.Image)))
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

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	log.Println("Clone asset source repo")
	assetSourceRepoPath := filepath.Join(assetSourcePath, filepath.Base(assetSourceRepo))
	os.Mkdir("data", 0777)
	exec.Command("git", "clone", assetSourceRepo, assetSourceRepoPath).Run()

	log.Println("Build assets.json")
	assetJson, err := exec.Command("tsx", filepath.Join(assetSourcePath, "extract_objects.ts")).Output()
	handleError(err)
	assets := make([]asset, 0)
	handleError(json.Unmarshal(assetJson, &assets))

	log.Println("Build assets.zip")
	zipFile, err := os.Create(outputPath)
	handleError(err)
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	log.Println("Compile object images")
	outputAssetList := make([]outputAsset, 0)
	for _, asset := range assets {
		log.Printf("Process asset %d (%s)", asset.ID, asset.Name)
		if asset.Image == "" {
			log.Println("  - No image, skipping")
			continue
		}
		log.Printf("  - Reading sprite from %s.webp as offset %d", asset.Image, asset.Offset)
		assetImage, err := asset.loadImage()
		handleError(err)
		var buf bytes.Buffer
		name := fmt.Sprintf("o%d.png", asset.ID)
		log.Printf("  - Save as %s", name)
		handleError(png.Encode(&buf, assetImage))
		handleError(writeToZip(writer, name, &buf))
		outputAssetList = append(outputAssetList, outputAsset{
			ID:    asset.ID,
			Name:  asset.Name,
			Scale: asset.Scale,
		})
	}

	log.Println("Compile additional images")
	for _, imagePath := range additionalImages {
		log.Printf("Process image at %s.webp", imagePath)
		pathTo := filepath.Join(assetSourceRepoPath, fmt.Sprintf("%s.webp", imagePath))
		image, err := loadWebpImage(pathTo)
		handleError(err)
		var buf bytes.Buffer
		handleError(png.Encode(&buf, image))
		name := fmt.Sprintf("x%s.png", filepath.Base(imagePath))
		log.Printf("  - Save as %s", name)
		handleError(writeToZip(writer, name, &buf))
	}

	log.Println("Compile additional files")
	for _, filePath := range additionalFiles {
		log.Printf("Process file at %s", filePath)
		pathTo := filepath.Join(assetSourcePath, filePath)
		file, err := os.Open(pathTo)
		handleError(err)
		defer file.Close()
		handleError(writeToZip(writer, filePath, file))
	}

	log.Println("Save assets.json")
	outputAssetJson, err := json.Marshal(outputAssetList)
	handleError(err)
	handleError(writeToZip(writer, "assets.json", bytes.NewReader(outputAssetJson)))

	log.Println("Done")
}
