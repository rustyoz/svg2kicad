package svg2kicadlib

import (
	"fmt"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
)

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

func parseTuple(l *lexer) (Tuple, error) {
	t := Tuple{}

	consumeWhiteSpace(l)

	ni := l.nextItem()
	if ni.typ == itemNumber {
		n, ok := strconv.ParseFloat(ni.val, 64)
		if ok != nil {
			return t, fmt.Errorf("Error passing number %s", ok)
		}
		t[0] = n
	} else {
		return t, fmt.Errorf("Error passing Tuple expected Number got %v", ni)
	}

	if l.peekItem().typ == itemWSP || l.peekItem().typ == itemComma {
		l.nextItem()
	}
	ni = l.nextItem()
	if ni.typ == itemNumber {
		n, ok := strconv.ParseFloat(ni.val, 64)
		if ok != nil {
			return t, fmt.Errorf("Error passing Number %s", ok)
		}
		t[1] = n
	} else {
		return t, fmt.Errorf("Error passing Tuple expected Number got: %v", ni)
	}

	return t, nil
}

func parseTransform(tstring string) (mt.Transform, error) {
	var tm mt.Transform
	lexer, _ := lex("tlexer", tstring)
	for {
		i := lexer.nextItem()
		switch i.typ {
		case itemEOS:
			break
		case itemWord:
			switch i.val {
			case "matrix":
				err := parseMatrix(lexer, &tm)
				return tm, err
				// case "scale":
				// case "rotate":

			}
		}
	}
}

func parseMatrix(l *lexer, t *mt.Transform) error {
	i := l.nextItem()
	if i.typ != itemParan {
		return fmt.Errorf("Error Parsing Transform Matrix: Expected Opening Parantheses")
	}
	var ncount int
	for {
		if ncount > 0 {
			for l.peekItem().typ == itemComma || l.peekItem().typ == itemWSP {
				l.nextItem()
			}
		}
		if l.peekItem().typ != itemNumber {
			return fmt.Errorf("Error Parsing Transform Matrix: Expected Number got %v", l.peekItem().String())
		}
		n, err := parseNumber(l.nextItem())
		if err != nil {
			return err
		}
		t[ncount%2][ncount/3] = n
		ncount++
		if ncount > 5 {
			i = l.peekItem()
			if i.typ != itemParan {
				return fmt.Errorf("Error Parsing Transform Matrix: Expected Closing Parantheses")
			}
			l.nextItem() // consume Parantheses
			return nil
		}
	}
}
