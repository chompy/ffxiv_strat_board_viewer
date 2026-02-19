package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
)

const assetZipPath = "assets.zip"
const assetJsonPath = "assets.json"
const defaultObjectScale = 1.0 / 200.0

type Asset struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Scale float64     `json:"scale"`
	Image image.Image `json:"-"`
}

var additionalImages = []string{"xcircle_aoe.png"}

/* Load asset list from zip reader */
func loadAssetList(zr *zip.ReadCloser) ([]Asset, error) {
	f, err := zr.Open(assetJsonPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	assets := make([]Asset, 0)
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

/* Load image from zip reader */
func loadImage(zr *zip.ReadCloser, name string) (image.Image, error) {
	f, err := zr.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

/* Load image for asset from zip reader */
func (a *Asset) loadAssetImage(zr *zip.ReadCloser) error {
	image, err := loadImage(zr, fmt.Sprintf("%d.png", a.ID))
	a.Image = image
	return err
}

/* Load background image from zip read */
func loadBackgroundImage(zr *zip.ReadCloser, id int) (image.Image, error) {
	return loadImage(zr, fmt.Sprintf("x%d.png", id+1))
}

func LoadBoardAssets(board Board) ([]Asset, error) {
	zr, err := zip.OpenReader(assetZipPath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	assets, err := loadAssetList(zr)
	if err != nil {
		return nil, err
	}

	boardAssets := make([]Asset, 0)
	for _, obj := range board.Objects {
		hasAsset := false
		for _, existingBoardAsset := range boardAssets {
			if existingBoardAsset.ID == obj.TypeID {
				hasAsset = true
				break
			}
		}
		if !hasAsset {
			for _, asset := range assets {
				if asset.ID == obj.TypeID {
					if err := asset.loadAssetImage(zr); err != nil {
						return nil, err
					}
					boardAssets = append(boardAssets, asset)
					break
				}
			}
		}
	}

	bgImage, err := loadBackgroundImage(zr, board.Background)
	boardAssets = append(boardAssets, Asset{Name: "Background", ID: -1, Image: bgImage})

	for i, imageName := range additionalImages {
		additionalImage, err := loadImage(zr, imageName)
		if err != nil {
			return nil, err
		}
		boardAssets = append(boardAssets, Asset{Name: imageName, ID: -i - 2, Image: additionalImage})
	}

	return boardAssets, nil
}
