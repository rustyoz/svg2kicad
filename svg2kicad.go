// Example SVG parser using a combination of xml.Unmarshal and the
// xml.Unmarshaler interface to handle an unknown combination of group
// elements where order is important.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/rustyoz/svg"
	"github.com/rustyoz/svg2kicad/svg2kicadlib"
)

func main() {
	var scale = flag.String("scale", "1.0", "scale svg units to mm")
	var filename = flag.String("f", "", "filename")
	flag.Parse()
	fmt.Println(flag.Args())
	if len(os.Args) < 2 {
		fmt.Println("Usage: svg2kicad filename.svg")
		return
	}
	fmt.Println("Reading:", *filename)
	buf, err := ioutil.ReadFile(*filename)
	if err != nil {
		fmt.Println("ioutil.ReadFile(os.Args[1]) Error: " + err.Error())
		return
	}

	svgStr := string(buf)
	var scalefloat float64
	scalefloat = 1.0
	fmt.Println("scale string", *scale)
	sf, ok := strconv.ParseFloat(*scale, 64)
	if ok == nil {
		fmt.Println("Scale:", sf)
		scalefloat = sf
	} else {
		fmt.Println("Error parsing scale flag ", ok)
	}
	svg, err := svg.ParseSvg(svgStr, *filename, scalefloat)
	if err != nil {
		fmt.Println("svg2kicadlib.ParseSvg(svgStr, filename, scalefloat) Error:" + err.Error())
		return
	}
	m := svg2kicadlib.SvgToKicadModule(svg)
	fmt.Println(m.Tags)
	var outfile *os.File
	outfile, err = os.Create(*filename + ".kicad_mod")
	_, err = outfile.WriteString(m.ToSExp())

	if err != nil {
		log.Println("Error writing file: ", *filename+".kicad_mod", err)
	}
}
