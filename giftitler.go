package main

import (
	"flag"
	"image/gif"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/paddyforan/giftitler/subtitles"
	"github.com/paddyforan/giftitler/writer"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
	var fontPath, subsPath, imagePath, out string
	var fontSize, dpi float64
	var strokeWeight int
	var offset time.Duration
	flag.StringVar(&fontPath, "font", "", "path to font to use")
	flag.StringVar(&subsPath, "subtitles", "", "path subtitles file to pull text from")
	flag.DurationVar(&offset, "subtitles-offset", 0, "timestamp of the subtitles the gif starts from")
	flag.StringVar(&imagePath, "gif", "", "path to gif")
	flag.StringVar(&out, "out", "output.gif", "path to write result")
	flag.Float64Var(&fontSize, "font-size", 0.0, "font size (in points) to write text in")
	flag.Float64Var(&dpi, "font-dpi", 72.0, "dpi to write text in")
	flag.IntVar(&strokeWeight, "font-stroke-weight", 3, "weight of the font stroke (outline) in pixels")
	flag.Parse()

	imagePath = os.ExpandEnv(imagePath)
	fontPath = os.ExpandEnv(fontPath)
	subsPath = os.ExpandEnv(subsPath)

	img, err := getBaseGIF(imagePath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	subs, err := getSubs(subsPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	font, err := getFont(fontPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	g, err := writer.Generate(img, writer.FontInfo{
		Font:         font,
		Size:         fontSize,
		DPI:          dpi,
		StrokeWeight: strokeWeight,
	}, subs, offset)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	f, err := os.Create(out)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	err = gif.EncodeAll(f, g)
	if err != nil {
		log.Println(err)
		f.Close()
		os.Exit(1)
	}
}

func getBaseGIF(location string) (*gif.GIF, error) {
	f, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return gif.DecodeAll(f)
}

func getSubs(location string) ([]subtitles.Subtitle, error) {
	f, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return subtitles.Parse(f)
}

func fontNamer(data draw2d.FontData) string {
	return data.Name
}

func getFont(location string) (*truetype.Font, error) {
	b, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, err
	}
	return truetype.Parse(b)
}
