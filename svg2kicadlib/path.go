package svg2kicadlib

import (
	"fmt"
	"strconv"

	"github.com/rustyoz/gokicadlib"
)

type Path struct {
	Id    string `xml:"id,attr"`
	D     string `xml:"d,attr"`
	group *Group
}

type PathDParser struct {
	lex            lexer
	x, y           float64
	currentcommand int
	tokbuf         [4]item
	peekcount      int
	sexpressions   chan interface{}
	lasttuple      Tuple
	transform      Transform
}

func NewPathDParse() *PathDParser {
	pdp := &PathDParser{}
	pdp.transform.Identity()
	return pdp
}

// peek returns but does not consume the next token.
func (t *PathDParser) peek() item {
	if t.peekcount > 0 {
		return t.tokbuf[t.peekcount-1]
	}
	t.peekcount = 1
	t.tokbuf[0] = t.lex.nextItem()
	return t.tokbuf[0]
}

func (pdp *PathDParser) next() item {
	if pdp.peekcount > 0 {
		pdp.peekcount--
	} else {
		pdp.tokbuf[0] = pdp.lex.nextItem()
	}
	return pdp.tokbuf[pdp.peekcount]
}

func (p Path) ParseD() chan interface{} {
	pdp := NewPathDParse()

	pdp.sexpressions = make(chan interface{})
	l, _ := lex("dlexer", p.D)
	pdp.lex = *l
	go func() {
		defer close(pdp.sexpressions)
		for {
			switch i := pdp.next(); {
			case i.typ == itemEOS:
				return
			case i.typ == itemWSP:
				pdp.lex.next()
			case i.typ == itemWord:
				parseCommand(pdp, l, i)
			}
		}
	}()
	return pdp.sexpressions
}

func (p Path) ToKicad() (ses []gokicadlib.SExpression) {
	c := p.ParseD()
	for s := range c {
		ses = append(ses, s.(gokicadlib.SExpression))
	}
	return ses
}

func parseNumber(i item) (float64, error) {
	var n float64
	var ok error
	if i.typ == itemNumber {
		n, ok = strconv.ParseFloat(i.val, 64)
		if ok != nil {
			return n, fmt.Errorf("Error passing number %s", ok)
		}
	}
	return n, nil
}

func parseCommand(pdp *PathDParser, l *lexer, i item) {
	switch i.val {
	case "M":
		parseMoveToAbs(pdp)
	case "m":
		parseMoveToRel(pdp)
	//	case "c":
	//	parseCurveToRel(pdp)
	//case "C":
	//	parseCurvetoAbs(*lexer)
	case "L":
		parseLineToAbs(pdp)
	case "l":
		parseLineToRel(pdp)
	case "H":
		parseHLineToAbs(pdp)
	case "h":
		parseHLineToRel(pdp)
	}
}

func parseMoveToAbs(pdp *PathDParser) error {
	t, err := parseTuple(pdp)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToAbs Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	var tuples []Tuple
	consumeWhiteSpace(pdp)
	for pdp.peek().typ == itemNumber {
		t, err := parseTuple(pdp)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(pdp)
	}
	for _, nt := range tuples {

		var l gokicadlib.Line
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
	consumeWhiteSpace(pdp)
	for pdp.peek().typ == itemNumber {
		t, err := parseTuple(pdp)
		if err != nil {
			return fmt.Errorf("Error Passing LineToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(pdp)
	}
	for _, nt := range tuples {
		var l gokicadlib.Line
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
	t, err := parseTuple(pdp)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToRel Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	var tuples []Tuple
	consumeWhiteSpace(pdp)
	for pdp.peek().typ == itemNumber {
		t, err := parseTuple(pdp)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(pdp)
	}
	for _, nt := range tuples {

		var l gokicadlib.Line
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
	consumeWhiteSpace(pdp)
	for pdp.peek().typ == itemNumber {
		t, err := parseTuple(pdp)
		if err != nil {
			return fmt.Errorf("Error Passing LineToRel\n%s", err)
		}
		tuples = append(tuples, t)
		consumeWhiteSpace(pdp)
	}
	for _, nt := range tuples {

		var l gokicadlib.Line

		l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
		pdp.x += nt[0]
		pdp.y += nt[1]
		l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
		l.Layer = gokicadlib.F_SilkS
		pdp.sexpressions <- &l
	}

	return nil
}

func parseTuple(pdp *PathDParser) (Tuple, error) {
	t := Tuple{}
	ni := pdp.next()

	for ni.typ == itemWSP {
		ni = pdp.next()
	}

	if ni.typ == itemNumber {
		n, ok := strconv.ParseFloat(ni.val, 64)
		if ok != nil {
			return t, fmt.Errorf("Error passing number %s", ok)
		}
		t[0] = n
	} else {
		return t, fmt.Errorf("Error passing Tuple expected Number")
	}
	ni = pdp.next()

	if ni.typ == itemWSP || ni.typ == itemComma {
		ni = pdp.next()
	}
	if ni.typ == itemNumber {
		n, ok := strconv.ParseFloat(ni.val, 64)
		if ok != nil {
			return t, fmt.Errorf("Error passing Number %s", ok)
		}
		t[1] = n
	} else {
		return t, fmt.Errorf("Error passing Tuple expected Number")
	}

	return t, nil
}

func consumeWhiteSpace(pdp *PathDParser) error {
	for {
		ni := pdp.peek()
		if ni.typ == itemWSP {
			pdp.next()
		} else {
			return nil
		}
	}
	return nil
}

func parseHLineToAbs(pdp *PathDParser) error {
	consumeWhiteSpace(pdp)
	var n float64
	var err error
	if pdp.peek().typ != itemNumber {
		n, err = parseNumber(pdp.next())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToAbs\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.x = n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil
}

func parseHLineToRel(pdp *PathDParser) error {
	consumeWhiteSpace(pdp)
	var n float64
	var err error
	if pdp.peek().typ != itemNumber {
		n, err = parseNumber(pdp.next())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToRel\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.x += n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil

}

func parseVLineToAbs(pdp *PathDParser) error {
	consumeWhiteSpace(pdp)
	var n float64
	var err error
	if pdp.peek().typ != itemNumber {
		n, err = parseNumber(pdp.next())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToAbs\n%s", err)
		}
	}

	var l gokicadlib.Line
	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.y = n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil
}

func parseVLineToRel(pdp *PathDParser) error {
	consumeWhiteSpace(pdp)
	var n float64
	var err error
	if pdp.peek().typ != itemNumber {
		n, err = parseNumber(pdp.next())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToRel\n%s", err)
		}
	}

	var l gokicadlib.Line

	l.Origin.X, l.Origin.Y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.y += n
	l.End.X, l.End.Y = pdp.transform.Apply(pdp.x, pdp.y)
	l.Layer = gokicadlib.F_SilkS
	pdp.sexpressions <- &l

	return nil

}
