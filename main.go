package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/llgcode/draw2d/draw2dimg"
)

// A glyph is a single character as displayed on the LCD
type glyph struct {
	data []byte
}

func (g *glyph) draw(ctx *draw2dimg.GraphicContext, colour color.RGBA, x, y float64, scalefactor int) {
	ctx.SetFillColor(colour)
	for col, coldata := range g.data {
		for row := uint8(0); row < 8; row++ {
			if coldata&(1<<row) != 0 {
				rect(ctx, x+float64(col*scalefactor), y+float64(int(row)*scalefactor), x+float64((col+1)*scalefactor), y+float64((int(row)+1)*scalefactor))
				ctx.Fill()
			}
		}
	}
}

type fontTable map[rune]glyph

type charSize struct {
	w, h int
}

type colorSet struct {
	background, pixelOff, pixelOn color.RGBA
}

var colorSetStandard = colorSet{
	color.RGBA{149, 210, 3, 255},
	color.RGBA{138, 195, 19, 255},
	color.RGBA{3, 13, 14, 255},
}

type lcd struct {
	rows, cols int
	charSize   charSize
	charSep    int // separation between characters
	colorSet   colorSet
	fonttable  fontTable
}

func rect(ctx *draw2dimg.GraphicContext, x1, y1, x2, y2 float64) {
	ctx.MoveTo(x1, y1)
	ctx.LineTo(x2, y1)
	ctx.LineTo(x2, y2)
	ctx.LineTo(x1, y2)
	ctx.LineTo(x1, y1)
}

func (d *lcd) Draw(message string) image.Image {
	// Convert message to glyphs
	glyphs := map[int]map[int]glyph{}
	row, col := 0, 0
	for _, char := range message {
		if char == '\n' {
			row++
			col = 0
			continue
		}
		if _, ok := glyphs[row]; !ok {
			glyphs[row] = map[int]glyph{}
		}
		// Lookup glyph replacing unknown characters with ?
		if g, ok := d.fonttable[char]; !ok {
			fmt.Println(char)
			glyphs[row][col] = d.fonttable['?']
		} else {
			glyphs[row][col] = g
		}
		col++
		if col >= d.cols { // just did last column on row
			row++
			col = 0
		}
	}
	scalefactor := 10
	pixelsW := (d.cols * d.charSize.w) + ((d.cols * d.charSep) + 1)
	pixelsH := (d.rows * d.charSize.h) + ((d.rows * d.charSep) + 1)
	img := image.NewRGBA(image.Rect(0, 0, pixelsW*scalefactor, pixelsH*scalefactor))
	ctx := draw2dimg.NewGraphicContext(img)
	ctx.SetFillColor(d.colorSet.background)
	ctx.SetLineWidth(0)
	rect(ctx, 0, 0, float64(pixelsW*scalefactor), float64(pixelsH*scalefactor))
	ctx.Fill()
	for r := 0; r < d.rows; r++ {
		for c := 0; c < d.cols; c++ {
			charX1 := float64(((c * (d.charSize.w + d.charSep)) + d.charSep) * scalefactor)
			charY1 := float64(((r * (d.charSize.h + d.charSep)) + d.charSep) * scalefactor)
			charX2 := float64(((c * (d.charSize.w + d.charSep)) + d.charSep + d.charSize.w) * scalefactor)
			charY2 := float64(((r * (d.charSize.h + d.charSep)) + d.charSep + d.charSize.h) * scalefactor)
			ctx.SetFillColor(d.colorSet.pixelOff)
			rect(
				ctx,
				charX1,
				charY1,
				charX2,
				charY2,
			)
			ctx.Fill()
			if g, ok := glyphs[r][c]; ok {
				g.draw(ctx, d.colorSet.pixelOn, charX1, charY1, scalefactor)
			}
		}
	}

	return img

}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s \"Message\"", os.Args[0])
		os.Exit(1)
	}
	display := lcd{
		2,
		16,
		charSize{5, 7},
		1,
		colorSetStandard,
		font,
	}
	img := display.Draw(strings.Join(os.Args[1:], "\n"))
	f, _ := os.Create("lcd.png")
	defer f.Close()
	png.Encode(f, img)

}
