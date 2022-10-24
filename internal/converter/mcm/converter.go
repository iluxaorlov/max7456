package mcm

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	MAX7456 = "MAX7456"
	WIDTH   = 12
	HEIGHT  = 18
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) Decode(filePath string) error {
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("can't get absolute path of file: %w", err)
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return fmt.Errorf("can't open file: %w", err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("can't read first line: %w", err)
	}
	if !strings.HasPrefix(line, MAX7456) {
		return fmt.Errorf("invalid file")
	}

	var glyphs [256][256]byte
	for i := 0; i < 256; i++ {
		var glyph [256]byte
		for j := 0; j < 64; j++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return fmt.Errorf("can't read line: %w", err)
			}
			for k := 0; k < 4; k++ {
				if line[k*2] == '1' {
					glyph[(j*4)+k] |= 1 << 1
				} else {
					glyph[(j*4)+k] |= 0 << 0
				}
				if line[(k*2)+1] == '1' {
					glyph[(j*4)+k] |= 1
				} else {
					glyph[(j*4)+k] |= 0
				}
			}
		}

		glyphs[i] = glyph
	}

	directoryPath := filepath.Join(filepath.Dir(absolutePath), c.clearFileName(absolutePath))
	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		return fmt.Errorf("can't create directory: %w", err)
	}
	for i, glyph := range glyphs {
		if err := func() error {
			bitmap := image.NewNRGBA(image.Rect(0, 0, WIDTH, HEIGHT))
			for y := 0; y < HEIGHT; y++ {
				for x := 0; x < WIDTH; x++ {
					pixel := glyph[(y*WIDTH)+x]
					switch pixel {
					case 0b00:
						bitmap.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
					case 0b10:
						bitmap.Set(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
					default:
						bitmap.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
					}
				}
			}

			file, err := os.Create(filepath.Join(directoryPath, fmt.Sprintf("0x%.2X.png", i)))
			if err != nil {
				return fmt.Errorf("can't create file: %w", err)
			}

			defer file.Close()

			if err := png.Encode(file, bitmap); err != nil {
				return fmt.Errorf("can't encode image: %w", err)
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) Encode(directoryPath string) error {
	absolutePath, err := filepath.Abs(directoryPath)
	if err != nil {
		return fmt.Errorf("can't get absolute path of file: %w", err)
	}

	entries, err := os.ReadDir(absolutePath)
	if err != nil {
		return fmt.Errorf("can't read directory: %w", err)
	}

	var glyphs [256][216]byte
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "0x") {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".png") {
			continue
		}

		if err := func() error {
			file, err := os.Open(filepath.Join(absolutePath, entry.Name()))
			if err != nil {
				return fmt.Errorf("can't open file: %w", err)
			}

			defer file.Close()

			img, err := png.Decode(file)
			if err != nil {
				return fmt.Errorf("can't decode image: %w", err)
			}
			if img.Bounds().Max.X != WIDTH && img.Bounds().Max.Y != HEIGHT {
				return fmt.Errorf("image %s is not 12px wide by 18px height", file.Name())
			}

			i, err := strconv.ParseInt(strings.TrimPrefix(c.clearFileName(file.Name()), "0x"), 16, 64)
			if err != nil {
				return fmt.Errorf("can't convert file  to decimal: %w", err)
			}
			if i > 255 {
				return nil
			}

			var glyph [216]byte
			for y := 0; y < HEIGHT; y++ {
				for x := 0; x < WIDTH; x++ {
					switch img.At(x, y) {
					case color.NRGBA{R: 0, G: 0, B: 0, A: 255}:
						glyph[(y*WIDTH)+x] = 0b00
					case color.NRGBA{R: 255, G: 255, B: 255, A: 255}:
						glyph[(y*WIDTH)+x] = 0b10
					default:
						glyph[(y*WIDTH)+x] = 0b01
					}
				}
			}

			glyphs[i] = glyph

			return nil
		}(); err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(absolutePath, fmt.Sprintf("%s.mcm", filepath.Base(absolutePath))))
	if err != nil {
		return fmt.Errorf("can't create file: %w", err)
	}

	defer file.Close()

	if _, err = file.WriteString(MAX7456); err != nil {
		return fmt.Errorf("can't write string to file: %w", err)
	}
	for _, glyph := range glyphs {
		for i := 0; i < 256; i++ {
			if i%4 == 0 {
				if _, err := file.WriteString("\n"); err != nil {
					return fmt.Errorf("can't write string to file: %w", err)
				}
			}

			if i >= len(glyph) {
				if _, err := file.WriteString("01"); err != nil {
					return fmt.Errorf("can't write string to file: %w", err)
				}
				continue
			}

			if _, err := file.WriteString(fmt.Sprintf("%d%d", (glyph[i]>>1)&1, glyph[i]&1)); err != nil {
				return fmt.Errorf("can't write string to file: %w", err)
			}
		}
	}

	return nil
}

func (c *Converter) clearFileName(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}
