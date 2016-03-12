package svg2kicadlib

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	mt "github.com/rustyoz/Mtransform"
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
	Transform   string
	transform   *mt.Transform // row, column
	group       *Group
	svg         *Svg
}

type Svg struct {
	Title       string  `xml:"title"`
	Groups      []Group `xml:"g"`
	ses         []gokicadlib.SExpression
	KicadOutput string
	Name        string
	transform   *mt.Transform
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
		case "transform":
			g.Transform = attr.Value
			t, err := parseTransform(g.Transform)
			if err != nil {
				fmt.Println(err)
			}
			g.transform = &t
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
				elementStruct = &Group{group: g, svg: g.svg, transform: mt.NewTransform()}
			case "rect":
				elementStruct = &Rect{group: g}
			case "path":
				elementStruct = &Path{group: g, strokeWidth: 1}

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

func ParseSvg(str string, name string, scale float64) (*Svg, error) {
	var svg Svg
	svg.Name = name
	svg.transform = mt.NewTransform()
	if scale > 0 {
		svg.transform.Scale(scale, scale)
	}
	if scale < 0 {
		svg.transform.Scale(1.0/-scale, 1.0/-scale)
	}
	fmt.Println(svg.transform)
	err := xml.Unmarshal([]byte(str), &svg)
	if err != nil {
		return nil, fmt.Errorf("ParseSvg Error: %v\n", err)
	}
	for i := range svg.Groups {
		svg.Groups[i].svg = &svg
		if svg.Groups[i].transform == nil {
			svg.Groups[i].transform = mt.NewTransform()
		}
	}

	for _, g := range svg.Groups {
		g.ToKicad()
	}
	return &svg, nil
}

type Tuple [2]float64

func (svg *Svg) ToKicadModule() gokicadlib.Module {
	var m gokicadlib.Module
	m.SExpressions = svg.ses
	fmt.Println("Number of kicad graphical elements: ", len(svg.ses))
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

func (g *Group) ToKicad() {
	if g.group != nil {
		g.transform.MultiplyWith(*g.group.transform)
	} else {
		g.transform.MultiplyWith(*g.svg.transform)
	}
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
