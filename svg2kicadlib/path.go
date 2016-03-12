package svg2kicadlib

import (
	"fmt"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
	"github.com/rustyoz/gokicadlib"
)

type Path struct {
	Id          string `xml:"id,attr"`
	D           string `xml:"d,attr"`
	Style       string `xml:"style,attr"`
	properties  map[string]string
	strokeWidth float64

	group *Group
}

type PathDParser struct {
	p              *Path
	lex            lexer
	x, y           float64
	currentcommand int
	tokbuf         [4]item
	peekcount      int
	sexpressions   chan interface{}
	lasttuple      Tuple
	transform      mt.Transform
}

func NewPathDParse() *PathDParser {
	pdp := &PathDParser{}
	pdp.transform = mt.Identity()
	return pdp
}

func (p Path) ParseD() chan interface{} {
	pdp := NewPathDParse()
	pdp.p = &p
	fmt.Println(p.group.transform)
	pdp.transform.MultiplyWith(*p.group.transform)
	fmt.Println(pdp.transform)
	pdp.sexpressions = make(chan interface{})
	l, _ := lex(fmt.Sprint(p.Id), p.D)
	pdp.lex = *l
	go func() {
		defer close(pdp.sexpressions)
		for {
			i := pdp.lex.nextItem()
			switch {
			case i.typ == itemError:
				return
			case i.typ == itemEOS:
				return
			case i.typ == itemLetter:
				parseCommand(pdp, l, i)
			default:
			}
		}
	}()
	return pdp.sexpressions
}

func (p *Path) ToKicad() (ses []gokicadlib.SExpression) {
	p.parseStyle()
	c := p.ParseD()
	for s := range c {
		ses = append(ses, s.(gokicadlib.SExpression))
	}
	return ses
}

func parseCommand(pdp *PathDParser, l *lexer, i item) error {
	var err error
	switch i.val {
	case "M":
		err = parseMoveToAbs(pdp)
	case "m":
		err = parseMoveToRel(pdp)
	case "c":
		parseCurveToRel(pdp)
	case "C":
		parseCurveToAbs(pdp)
	case "L":
		err = parseLineToAbs(pdp)
	case "l":
		err = parseLineToRel(pdp)
	case "H":
		err = parseHLineToAbs(pdp)
	case "h":
		err = parseHLineToRel(pdp)
	}
	//	fmt.Println(err)
	return err

}

func parseMoveToAbs(pdp *PathDParser) error {
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToAbs Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	var tuples []Tuple
	consumeWhiteSpace(&pdp.lex)
	for pdp.lex.peekItem().typ == itemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(&pdp.lex)
	}
	for _, nt := range tuples {

		var l gokicadlib.Line
		l.Width = pdp.p.strokeWidth
		l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
		pdp.x = nt[0]
		pdp.y = nt[1]
		l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}

	return nil

}

func parseLineToAbs(pdp *PathDParser) error {
	var tuples []Tuple
	consumeWhiteSpace(&pdp.lex)
	for pdp.lex.peekItem().typ == itemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing LineToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(&pdp.lex)
	}
	for _, nt := range tuples {
		var l gokicadlib.Line
		l.Width = pdp.p.strokeWidth

		l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
		pdp.x = nt[0]
		pdp.y = nt[1]
		l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}

	return nil

}

func parseMoveToRel(pdp *PathDParser) error {
	//	fmt.Println("parsemovetorel")
	consumeWhiteSpace(&pdp.lex)
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToRel Expected First Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	var tuples []Tuple
	consumeWhiteSpace(&pdp.lex)
	for pdp.lex.peekItem().typ == itemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(&pdp.lex)
	}
	for _, nt := range tuples {

		var l gokicadlib.Line
		l.Width = pdp.p.strokeWidth
		l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
		pdp.x += nt[0]
		pdp.y += nt[1]
		l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}

	return nil
}

func parseLineToRel(pdp *PathDParser) error {

	var tuples []Tuple
	consumeWhiteSpace(&pdp.lex)
	for pdp.lex.peekItem().typ == itemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing LineToRel\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(&pdp.lex)
	}
	for _, nt := range tuples {

		var l gokicadlib.Line
		l.Width = pdp.p.strokeWidth

		l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
		pdp.x += nt[0]
		pdp.y += nt[1]
		l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}

	return nil
}

func parseHLineToAbs(pdp *PathDParser) error {
	consumeWhiteSpace(&pdp.lex)
	var n float64
	var err error
	if pdp.lex.peekItem().typ != itemNumber {
		n, err = parseNumber(pdp.lex.nextItem())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToAbs\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Width = pdp.p.strokeWidth

	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.x = n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil
}

func parseHLineToRel(pdp *PathDParser) error {
	consumeWhiteSpace(&pdp.lex)
	var n float64
	var err error
	if pdp.lex.peekItem().typ != itemNumber {
		n, err = parseNumber(pdp.lex.nextItem())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToRel\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Width = pdp.p.strokeWidth
	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.x += n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil

}

func parseVLineToAbs(pdp *PathDParser) error {
	consumeWhiteSpace(&pdp.lex)
	var n float64
	var err error
	if pdp.lex.peekItem().typ != itemNumber {
		n, err = parseNumber(pdp.lex.nextItem())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToAbs\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Width = pdp.p.strokeWidth
	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.y = n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil
}

func parseVLineToRel(pdp *PathDParser) error {
	consumeWhiteSpace(&pdp.lex)
	var n float64
	var err error
	if pdp.lex.peekItem().typ != itemNumber {
		n, err = parseNumber(pdp.lex.nextItem())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToRel\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Width = pdp.p.strokeWidth
	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.y += n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil

}

func parseCurveToRel(pdp *PathDParser) error {
	var tuples []Tuple
	consumeWhiteSpace(&pdp.lex)
	for pdp.lex.peekItem().typ == itemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(&pdp.lex)
	}
	var cb CubicBezier
	cb.controlpoints[0][0] = pdp.x
	cb.controlpoints[0][1] = pdp.y
	for i, nt := range tuples {
		pdp.x += nt[0]
		pdp.y += nt[1]
		cb.controlpoints[i+1][0] = pdp.x
		cb.controlpoints[i+1][1] = pdp.y
	}
	vertices := cb.RecursiveInterpolate(10, 0)

	for i := 0; i < len(vertices)-1; i++ {
		var l gokicadlib.Line
		l.Width = pdp.p.strokeWidth

		l.Origin.X, l.Origin.Y = pdp.transform.Apply(vertices[i][0], vertices[i][1])
		l.End.X, l.End.Y = pdp.transform.Apply(vertices[i+1][0], vertices[i+1][1])
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}
	return nil
}

func parseCurveToAbs(pdp *PathDParser) error {
	var tuples []Tuple
	consumeWhiteSpace(&pdp.lex)
	for pdp.lex.peekItem().typ == itemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(&pdp.lex)
	}

	var cb CubicBezier
	cb.controlpoints[0][0] = pdp.x
	cb.controlpoints[0][1] = pdp.y
	for i, nt := range tuples {
		pdp.x = nt[0]
		pdp.y = nt[1]
		cb.controlpoints[i+1][0] = pdp.x
		cb.controlpoints[i+1][1] = pdp.y
	}
	vertices := cb.RecursiveInterpolate(10, 0)

	for i := 0; i < len(vertices)-1; i++ {
		var l gokicadlib.Line
		l.Width = pdp.p.strokeWidth

		l.Origin.X, l.Origin.Y = pdp.transform.Apply(vertices[i][0], vertices[i][1])
		l.End.X, l.End.Y = pdp.transform.Apply(vertices[i+1][0], vertices[i+1][1])
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}
	return nil
}

func (p *Path) parseStyle() {
	p.properties = splitStyle(p.Style)
	for key, val := range p.properties {
		switch key {
		case "stroke-width":
			sw, ok := strconv.ParseFloat(val, 64)
			if ok == nil {
				p.strokeWidth = sw
			}

		}
	}
}
