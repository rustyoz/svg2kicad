// Example SVG parser using a combination of xml.Unmarshal and the
// xml.Unmarshaler interface to handle an unknown combination of group
// elements where order is important.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/rustyoz/svg2kicad/svg2kicadlib"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Println(wd)
	if len(os.Args) < 2 {
		fmt.Println("Usage: svg2kicad filename.svg")
		return
	}

	buf, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	svgStr := string(buf)

	svg, err := svg2kicadlib.ParseSvg(svgStr, os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	m := svg.ToKicadModule()

	var outfile *os.File
	outfile, err = os.Create(os.Args[1] + ".kicad_mod")
	_, err = outfile.WriteString(m.ToSExp())

	if err != nil {
		log.Println("Error writing file: ", os.Args[1]+".kicad_mod", err)
	}
	//fmt.Println(svg)
}
