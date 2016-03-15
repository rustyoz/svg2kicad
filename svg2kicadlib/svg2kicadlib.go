package svg2kicadlib

import (
	"fmt"
	"strings"

	"github.com/rustyoz/gokicadlib"
	"github.com/rustyoz/svg"
)

func SplitD(c rune) bool {
	return strings.Contains(" ,", string(c))
}

func PathToKicad(p *svg.Path) (ses []gokicadlib.SExpression) {

	segments := p.Parse()

	for s := range segments {
		if s.Closed {
			var poly gokicadlib.Polygon
			for _, p := range s.Points {
				poly.Points = append(poly.Points, gokicadlib.Point{p[0], p[1]})
			}
			poly.Layer = gokicadlib.F_SilkS
			poly.Width = s.Width
			ses = append(ses, &poly)

		} else {
			for i := range s.Points[1:] {
				var l gokicadlib.Line
				l.Layer = gokicadlib.F_SilkS
				l.Width = s.Width
				l.Origin.X = s.Points[i][0]
				l.Origin.Y = s.Points[i][1]
				l.End.X = s.Points[i+1][0]
				l.End.Y = s.Points[i+1][1]
				ses = append(ses, &l)
			}
		}
	}
	return ses
}

func SvgToKicadModule(svg *svg.Svg) gokicadlib.Module {
	var m gokicadlib.Module
	for _, g := range svg.Groups {
		m.SExpressions = append(m.SExpressions, GroupToKicad(&g)...)
	}
	fmt.Println("Number of kicad graphical elements: ", len(m.SExpressions))
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

func GroupToKicad(g *svg.Group) []gokicadlib.SExpression {
	var ses []gokicadlib.SExpression
	if g.Parent != nil {
		g.Transform.MultiplyWith(*g.Parent.Transform)
	} else {
		g.Transform.MultiplyWith(*g.Owner.Transform)
	}
	for _, elem := range g.Elements {
		switch elem.(type) {
		case *svg.Path:
			ses = append(ses, PathToKicad(elem.(*svg.Path))...)
		case *svg.Group:

			ses = append(ses, GroupToKicad(elem.(*svg.Group))...)
		default:
		}
	}
	return ses
}
