package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"

	"github.com/generaltso/vibrant"
)

var (
	input_stdin      bool
	output_json      bool
	output_css       bool
	output_compress  bool
	output_lowercase bool
	output_rgb       bool
)

func usage() {
	println("usage: vibrant [options] file")
	println("       cat image.jpg | vibrant -i [options]")
	println()
	println("options:")
	flag.PrintDefaults()
	os.Exit(2)
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func main() {
	flag.Usage = usage

	flag.BoolVar(&input_stdin, "i", false, "Read image data from stdin")
	flag.BoolVar(&output_compress, "compress", false, "Strip whitespace from output.")
	flag.BoolVar(&output_css, "css", false, "Output results in CSS.")
	flag.BoolVar(&output_json, "json", false, "Output results in JSON.")
	flag.BoolVar(&output_lowercase, "lowercase", true, "Use lowercase only for all output.")
	flag.BoolVar(&output_rgb, "rgb", false, "Output RGB instead of HTML hex, e.g. #ffffff.")

	flag.Parse()

	var img image.Image
	var err error

	if input_stdin {
		img, _, err = image.Decode(os.Stdin)
	} else {
		filename := flag.Arg(0)
		if filename == "" {
			usage()
		}

		f, err := os.Open(filename)
		checkErr(err)

		img, _, err = image.Decode(f)
		f.Close()
	}
	checkErr(err)

	palette, err := vibrant.NewPaletteFromImage(img)
	checkErr(err)

	switch {
	case output_json && output_css:
		usage()
	case output_json:
		print_json(palette)
	case output_css:
		print_css(palette)
	default:
		print_plain(palette)
	}
}

type swatch struct {
	Color string
	Text  string
}

func print_json(palette vibrant.Palette) {
	out := map[string]interface{}{}
	for name, sw := range palette.ExtractAwesome() {
		if output_rgb {
			r, g, b := sw.Color.RGB()
			out[name] = map[string]int{"r": r, "g": g, "b": b}
		} else {
			out[name] = swatch{sw.Color.RGBHex(), sw.Color.TitleTextColor().RGBHex()}
		}
	}
	var b []byte
	var err error
	if output_compress {
		b, err = json.Marshal(out)
	} else {
		b, err = json.MarshalIndent(out, "", "  ")
	}
	checkErr(err)

	str := string(b)
	if output_lowercase {
		str = strings.ToLower(str)
	}
	fmt.Println(str)
}

func rgb(r ...int) string {
	return fmt.Sprintf("rgb(%d,%d,%d)", r[0], r[1], r[2])
}

func print_css(palette vibrant.Palette) {
	sp := " "
	lf := "\n"
	tb := "  "
	sc := ";"
	if output_compress {
		sp = ""
		lf = ""
		tb = ""
		sc = ""
	}
	for name, sw := range palette.ExtractAwesome() {
		var bgcolor string
		var fgcolor string
		if output_rgb {
			bgcolor = rgb(sw.Color.RGB())
			fgcolor = rgb(sw.Color.TitleTextColor().RGB())
		} else {
			bgcolor = sw.Color.RGBHex()
			fgcolor = sw.Color.TitleTextColor().RGBHex()
		}
		if output_lowercase {
			name = strings.ToLower(name)
			bgcolor = strings.ToLower(bgcolor)
			fgcolor = strings.ToLower(fgcolor)
		}
		if output_compress && !output_rgb {
			bgcolor = shorthex(bgcolor)
			fgcolor = shorthex(fgcolor)
		}
		fmt.Printf(".%s%s{%s", name, sp, lf)
		fmt.Printf("%sbackground-color:%s%s;%s", tb, sp, bgcolor, lf)
		fmt.Printf("%scolor:%s%s%s%s}%s", tb, sp, fgcolor, sc, lf, lf)
	}
}

func shorthex(hex string) string {
	x := []byte(hex)
	if x[1] == x[2] && x[3] == x[4] && x[5] == x[6] {
		return "#" + string(x[1]) + string(x[3]) + string(x[5])
	}
	return hex
}

func print_plain(palette vibrant.Palette) {
	for name, sw := range palette.ExtractAwesome() {
		fmt.Printf("% 12s: %s, population: %d\n", name, sw.Color.RGBHex(), sw.Population)
	}
}
