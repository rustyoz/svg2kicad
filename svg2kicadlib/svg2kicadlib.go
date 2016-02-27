package svg2kicadlib

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/rustyoz/gokicadlib"
)

func SplitD(c rune) bool {
	return strings.Contains(" ,", string(c))
}

type Group struct {
	Groups      []Group `xml:"g"`
	Id          string
	Stroke      string
	StrokeWidth int32
	Fill        string
	FillRule    string
	Elements    []interface{}
	transform   Transform // row, column
	group       *Group
	svg         *Svg
}

type Svg struct {
	Title       string  `xml:"title"`
	Groups      []Group `xml:"g"`
	ses         []gokicadlib.SExpression
	KicadOutput string
	Name        string
}

// Implements encoding.xml.Unmarshaler interface
func (g *Group) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			g.Id = attr.Value
		case "stroke":
			g.Stroke = attr.Value
		case "stroke-width":
			if intValue, err := strconv.ParseInt(attr.Value, 10, 32); err != nil {
				return err
			} else {
				g.StrokeWidth = int32(intValue)
			}
		case "fill":
			g.Fill = attr.Value
		case "fill-rule":
			g.FillRule = attr.Value
		}
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			var elementStruct interface{}

			switch tok.Name.Local {
			case "g":
				elementStruct = &Group{group: g, svg: g.svg}
			case "rect":
				elementStruct = &Rect{group: g}
			case "path":
				elementStruct = &Path{group: g}
			}

			if err = decoder.DecodeElement(elementStruct, &tok); err != nil {
				return fmt.Errorf("Error decoding element of Group\n%s", err)
			} else {
				g.Elements = append(g.Elements, elementStruct)
			}

		case xml.EndElement:
			return nil
		}
	}
}

func ParseSvg(str string, name string) (*Svg, error) {
	var svg Svg
	//	svg.Name = name
	//	fmt.Println(str)
	err := xml.Unmarshal([]byte(str), &svg)
	if err != nil {
		return nil, fmt.Errorf("ParseSvg Error: %v\n", err)
	}
	for i := range svg.Groups {
		fmt.Printf("B  %p \n", &svg.Groups[i])
		svg.Groups[i].svg = &svg

		svg.Groups[i].SetSVGPointer(&svg)

	}

	for _, g := range svg.Groups {
		fmt.Printf("B %p \n", &g)
		g.ToKicad()
	}
	return &svg, nil
}

type Tuple [2]float64

func (svg *Svg) ToKicadModule() gokicadlib.Module {
	var m gokicadlib.Module
	fmt.Println(len(svg.ses))
	m.SExpressions = svg.ses
	m.Layer = gokicadlib.F_SilkS
	m.Reference.Text = "REF**"
	m.Reference.Type = "reference"
	m.Reference.Layer = gokicadlib.F_SilkS
	m.Value.Type = "value"
	m.Value.Text = svg.Name
	m.Tags = []string{svg.Name}
	m.Value.Layer = gokicadlib.F_Fab
	return m
}

type Transform [3][3]float64

func (t *Transform) Apply(x float64, y float64) (float64, float64) {
	var X, Y float64
	X = t[0][0]*x + t[0][1] + t[0][2]
	Y = t[1][0]*y + t[1][1] + t[1][2]
	return X, Y
}

func (t *Transform) Identity() {
	t[0][0] = 1
	t[0][1] = 0
	t[0][2] = 0
	t[1][0] = 0
	t[1][1] = 1
	t[1][2] = 0
	t[2][0] = 0
	t[2][1] = 0
	t[2][2] = 1

}
func (t *Transform) String() string {
	return fmt.Sprintln(t)
}

func (g *Group) ToKicad() {
	for _, elem := range g.Elements {
		switch elem.(type) {
		case *Path:
			g.svg.ses = append(g.svg.ses, elem.(*Path).ToKicad()...)
		case *Rect:
			g.svg.ses = append(g.svg.ses, elem.(*Rect).ToKicad())
		case *Group:
			elem.(*Group).ToKicad()
		default:
		}
	}
}

func (g *Group) SetSVGPointer(svg *Svg) {
	g.svg = svg
	for _, ng := range g.Elements {
		switch ng.(type) {
		case *Group:
			ng.(*Group).svg = g.svg
			ng.(*Group).SetSVGPointer(svg)
		}
	}
}
