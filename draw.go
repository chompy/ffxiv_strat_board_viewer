package strategy_board

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"slices"

	"github.com/fogleman/gg"
)

const canvasWidth = 1024
const canvasHeight = 768

func Draw(board Board) (*gg.Context, error) {
	// load assets for given board
	assetList, err := board.Assets()
	if err != nil {
		return nil, err
	}

	log.Println("Draw strategy board")

	// create canvas
	c := gg.NewContext(canvasWidth, canvasHeight)

	// draw background
	for _, asset := range assetList {
		if asset.ID == -1 {
			log.Printf("   - Draw background (ID=%d)", board.Background)
			c.DrawImage(asset.Image, 0, 0)
			break
		}
	}

	// draw each board object
	for _, object := range slices.Backward(board.Objects) {
		if err := drawObject(object, assetList, c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func drawObject(object Object, assets []Asset, c *gg.Context) error {
	log.Printf("  - Draw object (ID=%d)", object.TypeID)
	if !object.Visible {
		log.Println("    - Object not visible, skipping")
		return nil
	}
	switch object.TypeID {
	case 10:
		arcImage, err := loadArcImage(nil)
		if err != nil {
			return err
		}
		return drawArc(object, arcImage, c)
	case 11:
		return drawLineAoe(object, c)
	case 12:
		return drawLine(object, c)
	case 17:
		return drawArc(object, nil, c)
	case 100:
		return drawTextObject(object, c)
	default:
		{
			for _, a := range assets {
				if a.ID == object.TypeID {
					return drawImageObject(object, &a, c)
				}
			}
		}
	}
	log.Printf("Asset not found: %d", object.TypeID)
	return AssetNotFound
}

func drawTextObject(object Object, c *gg.Context) error {
	if object.TypeID != 100 {
		return DrawUnexpectedObjectError
	}
	if object.Text == "" {
		return nil
	}
	fontFace, err := loadFont(nil)
	if err != nil {
		return err
	}
	c.SetFontFace(fontFace)
	c.SetColor(color.NRGBA{0, 0, 0, object.Color.A})
	c.DrawStringAnchored(object.Text, float64(object.X), float64(object.Y), 0.5, 0.5)
	c.SetColor(object.Color)
	c.DrawStringAnchored(object.Text, float64(object.X)-2, float64(object.Y)-2, 0.5, 0.5)
	c.Identity()
	return nil
}

func drawImageObject(object Object, asset *Asset, c *gg.Context) error {
	c.Translate(float64(object.X), float64(object.Y))
	c.Scale(object.ScaleFactor(asset.Scale))
	c.Rotate(gg.Radians(float64(object.Angle)))

	// TODO image transparency

	c.DrawImageAnchored(asset.Image, 0, 0, .5, .5)
	c.Identity()
	return nil
}

func drawLineAoe(object Object, c *gg.Context) error {
	c.Translate(float64(object.X), float64(object.Y))
	c.Rotate(gg.Radians(float64(object.Angle)))
	w, h := float64(object.Params[0]), float64(object.Params[1])
	c.DrawRectangle(-w, -h, w*2, h*2)
	c.SetColor(object.Color)
	c.Fill()
	c.Identity()
	return nil
}

func drawLine(object Object, c *gg.Context) error {
	x2, y2 := math.Round(float64(object.Params[0])/5120*canvasWidth), math.Round(float64(object.Params[1])/3840*canvasHeight)
	c.SetLineWidth(float64(object.Params[2]) * 2)
	c.SetColor(object.Color)
	c.MoveTo(float64(object.X), float64(object.Y))
	c.LineTo(x2, y2)
	c.Stroke()
	c.SetColor(color.NRGBA{255, 255, 255, object.Color.A})
	c.DrawPoint(float64(object.X), float64(object.Y), float64(object.Params[2]))
	c.Fill()
	c.SetColor(color.NRGBA{255, 255, 255, object.Color.A})
	c.DrawPoint(x2, y2, float64(object.Params[2]))
	c.Fill()
	c.Identity()
	return nil
}

func drawArc(object Object, image image.Image, c *gg.Context) error {
	// calculate the angle of the arc and its radius
	arcAngle := float64(object.Params[0]) / 180.0 * math.Pi
	startAngle := -math.Pi / 2.0
	endAngle := startAngle + arcAngle
	innerRadius := float64(object.Params[1])
	outerRadius := 256.0
	if object.TypeID == 17 {
		outerRadius = 250.0
	}

	// calculate the bounding box of the arc
	leftEdge, rightEdge, bottomEdge := 0.0, 0.0, 0.0
	if arcAngle < math.Pi*1.5 && arcAngle >= math.Pi {
		leftEdge = (1 + math.Sin(arcAngle)) * outerRadius
	} else if arcAngle < math.Pi {
		leftEdge = outerRadius
	}
	if arcAngle < math.Pi {
		if arcAngle >= math.Pi*0.5 {
			bottomEdge = (1 + math.Cos(arcAngle)) * outerRadius
		} else {
			bottomEdge = outerRadius + math.Cos(arcAngle)*innerRadius
		}
	}
	if arcAngle < math.Pi*0.5 {
		rightEdge = (1 - math.Sin(arcAngle)) * outerRadius
	}

	ox, oy := -(leftEdge-rightEdge)/2.0, bottomEdge/2.0

	// draw the arc and its inner circle
	nc := gg.NewContext(canvasWidth, canvasHeight)
	nc.Translate(float64(object.X)+ox, float64(object.Y)+oy)
	nc.RotateAbout(gg.Radians(float64(object.Angle)), -ox, -oy)
	sx, sy := object.ScaleFactor(.02)
	nc.ScaleAbout(sx, sy, -ox, -oy)
	nc.DrawArc(0, 0, outerRadius, startAngle, endAngle)
	nc.LineTo(innerRadius*math.Cos(endAngle), innerRadius*math.Sin(endAngle))
	nc.DrawArc(0, 0, innerRadius, endAngle, startAngle)

	if image != nil {
		// draw arc using image as mask
		nc.Clip()
		s := float64(object.Scale) * .01
		nc.ScaleAbout(s, s, -ox, -oy)
		nc.DrawImageAnchored(image, int(ox), int(oy), 0.5, 0.5)
	} else {
		// draw arc using solid color
		nc.SetColor(color.NRGBA{254, 161, 49, object.Color.A})
		nc.Fill()
	}

	c.DrawImage(nc.Image(), 0, 0)

	return nil

}
