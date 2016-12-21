package writer

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/truetype"
	"github.com/paddyforan/giftitler/subtitles"
)

type FontInfo struct {
	Font         *truetype.Font
	Size         float64
	DPI          float64
	StrokeWeight int
}

func Generate(img *gif.GIF, fontInfo FontInfo, subs []subtitles.Subtitle, offset time.Duration) (*gif.GIF, error) {
	var titles []subtitles.Subtitle
	startTime := offset
	var wg sync.WaitGroup
	if fontInfo.Size == 0.0 {
		fontInfo.Size = calculateFontSize(img, fontInfo, subs, offset)
		log.Println("calculated font size to be", fontInfo.Size)
	}
	for pos := range img.Image {
		delay := time.Duration(img.Delay[pos]) * time.Second / 100
		if len(titles) < 1 || titles[len(titles)-1].End <= startTime {
			titles = subtitles.GetSubtitles(subs, startTime, delay)
		}
		startTime += delay
		if len(titles) == 0 {
			continue
		}
		text := ""
		for _, title := range titles {
			text = text + "\n" + title.Text
		}
		strings.TrimSpace(text)
		wg.Add(1)
		go genFrame(img.Image[pos], fontInfo, text, &wg)
	}
	wg.Wait()
	return img, nil
}

func calculateFontSize(img *gif.GIF, fontInfo FontInfo, subs []subtitles.Subtitle, offset time.Duration) float64 {
	startTime := offset
	title := ""
	var titles []subtitles.Subtitle
	for pos := range img.Image {
		delay := time.Duration(img.Delay[pos]) * time.Second / 100
		var reused bool
		if len(titles) < 1 || titles[len(titles)-1].End <= startTime {
			titles = subtitles.GetSubtitles(subs, startTime, delay)
		} else {
			reused = true
		}
		startTime += delay
		if len(titles) == 0 || reused {
			continue
		}
		for _, t := range titles {
			split := strings.Split(t.Text, "\n")
			for _, line := range split {
				if len(line) > len(title) {
					title = line
				}
			}
		}
	}
	bounds := img.Image[0].Bounds()
	width := bounds.Max.X - bounds.Min.X
	paddedWidth := int(0.9 * float64(width))
	desiredWidth := float64(paddedWidth) * 0.9
	face := truetype.NewFace(fontInfo.Font, &truetype.Options{
		Size:    24,
		Hinting: font.HintingFull,
		DPI:     fontInfo.DPI,
	})
	drawer := &font.Drawer{
		Face: face,
	}
	return 24.0 * desiredWidth / float64(drawer.MeasureString(title).Floor())
}

func genFrame(img *image.Paletted, fontInfo FontInfo, text string, wg *sync.WaitGroup) error {
	defer wg.Done()
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	paddedWidth := int(0.9 * float64(width))
	paddedHeight := int(0.9 * float64(height))
	rawLines := strings.Split(text, "\n")
	var lines []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	face := truetype.NewFace(fontInfo.Font, &truetype.Options{
		Size:    fontInfo.Size,
		Hinting: font.HintingFull,
		DPI:     fontInfo.DPI,
	})
	prevLineOffset := fixed.I(0)
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	for lineNo := range lines {
		drawer.Src = image.NewUniform(color.Black)
		line := lines[len(lines)-(1+lineNo)]
		textWidth := drawer.MeasureString(line)
		textHeight := face.Metrics().Height
		if textWidth+fixed.I(fontInfo.StrokeWeight*2) > fixed.I(paddedWidth) {
			// TODO(paddy): split into multiple lines or resize text?
			fmt.Errorf("error: text width (%d) is greater than the space available (%d)", textWidth, fixed.I(paddedWidth))
		}
		dot := fixed.Point26_6{
			X: fixed.I(width)/2 - textWidth/2,
			Y: fixed.I(paddedHeight) - prevLineOffset,
		}
		for strokeY := -fontInfo.StrokeWeight; strokeY <= fontInfo.StrokeWeight; strokeY++ {
			for strokeX := -fontInfo.StrokeWeight; strokeX <= fontInfo.StrokeWeight; strokeX++ {
				if strokeX*strokeX+strokeY*strokeY >= fontInfo.StrokeWeight*fontInfo.StrokeWeight {
					continue // this rounds the corners, apparently
				}
				strokeDot := fixed.Point26_6{
					X: dot.X + fixed.I(strokeX),
					Y: dot.Y + fixed.I(strokeY),
				}
				drawer.Dot = strokeDot
				drawer.DrawString(line)
			}
		}
		drawer.Src = image.NewUniform(color.White)
		drawer.Dot = dot
		drawer.DrawString(line)
		prevLineOffset = (fixed.I(paddedHeight) - dot.Y) + textHeight + fixed.I(fontInfo.StrokeWeight)
	}
	return nil
}
