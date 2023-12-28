package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hinshun/vt10x"
	"github.com/muesli/termenv"
	player "github.com/xakep666/asciinema-player/v3"
)

type Terminal struct {
	vt10x vt10x.Terminal
}

func (t Terminal) Write(p []byte) (n int, err error)      { return t.vt10x.Write(p) }
func (t Terminal) Close() error                           { return nil }
func (t Terminal) Dimensions() (width, height int)        { return t.vt10x.Size() }
func (t Terminal) ToRaw() error                           { return nil }
func (t Terminal) Restore() error                         { return nil }
func (t Terminal) Control(control player.PlaybackControl) {}

func (t Terminal) Cell(x, y int) (glyph vt10x.Glyph, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
		}
	}()
	return t.vt10x.Cell(x, y), nil
}
func (t Terminal) Glyps() [][]vt10x.Glyph {
	t.vt10x.Lock()
	defer t.vt10x.Unlock()

	cols, rows := t.Dimensions()
	glyphs := make([][]vt10x.Glyph, rows)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			g, err := t.Cell(x, y)
			if err != nil {
				break
			}

			glyphs[y] = append(glyphs[y], g)
		}
	}
	return glyphs
}

const (
	attrReverse = 1 << iota
	attrUnderline
	attrBold
	attrGfx
	attrItalic
	attrBlink
	attrWrap
)

func (t Terminal) RawString() string {
	s := ""
	// var bg, fg vt10x.Color
	for _, row := range t.Glyps() {
		for _, col := range row {
			c := termenv.String(string(col.Char))
			if col.BG == vt10x.DefaultBG {
				c = c.Background(termenv.RGBColor("#000000")) // TODO fix default: related to termenv profile?
			} else {
				c = c.Background(t.color(col.BG))
			}

			if col.FG == vt10x.DefaultFG {
				c = c.Foreground(termenv.RGBColor("#FFFFFF")) // TODO fix default: related to termenv profile?
			} else {
				c = c.Foreground(t.color(col.FG))
			}

			if col.Mode&attrReverse == attrReverse {
				c = c.Reverse()
			}

			if col.Mode&attrUnderline == attrUnderline {
				c = c.Underline()
			}
			if col.Mode&attrBold == attrBold {
				c = c.Bold()
			}

			// TODO gfx??

			if col.Mode&attrItalic == attrItalic {
				c = c.Italic()
			}
			if col.Mode&attrBlink == attrBlink {
				c = c.Blink()
			}

			// TODO wrap??

			s += c.String()
		}
		s += "\n\r"
	}
	return strings.TrimSuffix(s, "\n\r")
}

func rgb(j int) (r, g, b int) {
	return (j >> 16) & 0xff, (j >> 8) & 0xff, j & 0xff
}

func (t Terminal) color(j vt10x.Color) termenv.Color {
	if j.ANSI() {
		return termenv.ANSIColor(j)
	}

	if j < 256 {
		return termenv.ANSI256Color(j)
	}

	r, g, b := rgb(int(j))
	return termenv.RGBColor(fmt.Sprintf("#%02x%02x%02x", r, g, b)) // TODO hex color is still wrong
}

func main() {
	file, err := os.Open(os.Args[1]) // TODO check arg length
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	stream, err := player.NewStreamFrameSource(file)
	if err != nil {
		panic(err.Error())
	}

	terminal := Terminal{
		vt10x: vt10x.New(vt10x.WithSize(stream.Header().Width, stream.Header().Height)),
	}

	position, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		panic(err.Error())
	}

	for stream.Next() {
		f := stream.Frame()

		if f.Time > position {
			break
		}

		if f.Type == player.OutputFrame {
			terminal.Write(f.Data)
		}
	}

	// fmt.Println(terminal.vt10x.String())
	fmt.Println(terminal.RawString())
	// fmt.Printf("%#v", "data:text/plain,"+terminal.RawString())
}
