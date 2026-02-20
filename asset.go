package strategy_board

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed assets.zip
var assetsZipArchive []byte

const assetJsonPath = "assets.json"
const assetFontPath = "Roboto-Medium.ttf"
const assetFontSize = 30
const defaultObjectScale = 1.0 / 200.0
const arcImagePath = "xcircle_aoe.png"

type Asset struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Scale float64     `json:"scale"`
	Image image.Image `json:"-"`
}

var assetList []Asset
var boardFont font.Face
var arcImage image.Image

/* Read asset zip archive stored as go embed */
func loadAssetsZip() (*zip.Reader, error) {
	log.Println("Read assets zip archive")
	return zip.NewReader(bytes.NewReader(assetsZipArchive), int64(len(assetsZipArchive)))
}

/* Load asset from zip reader */
func loadAsset(zr *zip.Reader, name string) ([]byte, error) {
	if zr == nil {
		var err error
		zr, err = loadAssetsZip()
		if err != nil {
			return nil, err
		}
	}
	log.Printf("  - Load asset %s", name)
	f, err := zr.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

/* Load asset list from zip reader */
func loadAssetList(zr *zip.Reader) ([]Asset, error) {
	if assetList != nil {
		return assetList, nil
	}
	data, err := loadAsset(zr, assetJsonPath)
	if err != nil {
		return nil, err
	}
	assetList = make([]Asset, 0)
	if err := json.Unmarshal(data, &assetList); err != nil {
		return nil, err
	}
	for i := range assetList {
		if assetList[i].Scale == 0 {
			assetList[i].Scale = defaultObjectScale
		}
	}
	return assetList, nil
}

/* Load font used for text in strategy board */
func loadFont(zr *zip.Reader) (font.Face, error) {
	if boardFont != nil {
		return boardFont, nil
	}
	var err error
	if zr == nil {
		zr, err = loadAssetsZip()
		if err != nil {
			return nil, err
		}
	}
	fontBytes, err := loadAsset(zr, assetFontPath)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	boardFont = truetype.NewFace(font, &truetype.Options{Size: assetFontSize})
	return boardFont, nil
}

/* Load image from zip reader */
func loadImage(zr *zip.Reader, name string) (image.Image, error) {
	if zr == nil {
		var err error
		zr, err = loadAssetsZip()
		if err != nil {
			return nil, err
		}
	}
	log.Printf("  - Load asset %s", name)
	f, err := zr.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

/* Load image for asset from zip reader */
func (a *Asset) loadAssetImage(zr *zip.Reader) error {
	image, err := loadImage(zr, fmt.Sprintf("o%d.png", a.ID))
	a.Image = image
	return err
}

/* Load background image from zip read */
func loadBackgroundImage(zr *zip.Reader, id int) (image.Image, error) {
	return loadImage(zr, fmt.Sprintf("x%d.png", id+1))
}

/* Load arc image, aka circle aoe, used by a few objects */
func loadArcImage(zr *zip.Reader) (image.Image, error) {
	if arcImage != nil {
		return arcImage, nil
	}
	var err error
	arcImage, err = loadImage(zr, arcImagePath)
	return arcImage, err
}

/* Load assets needed by given strategy board */
func (b Board) Assets() ([]Asset, error) {
	// load asset data
	zr, err := loadAssetsZip()
	assets, err := loadAssetList(zr)
	if err != nil {
		return nil, err
	}

	// build list of board specific assets
	boardAssets := make([]Asset, 0)
	for _, obj := range b.Objects {
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

	// load background image as special asset (ID: -1)
	bgImage, err := loadBackgroundImage(zr, b.Background)
	boardAssets = append(boardAssets, Asset{Name: "Background", ID: -1, Image: bgImage})

	// preload additional assets
	loadArcImage(zr)
	loadFont(zr)

	return boardAssets, nil
}
