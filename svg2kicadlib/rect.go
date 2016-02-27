package svg2kicadlib

import (
	"fmt"

	"github.com/rustyoz/gokicadlib"
)

type Rect struct {
	Id        string `xml:"id,attr"`
	Width     string `xml:"width,attr"`
	height    string `xml:"height,attr"`
	Transform string `xml:"transform,attr"`

	transform Transform
	group     *Group
}

func (r *Rect) ToKicad() (ses gokicadlib.SExpression) {
	if len(r.Transform) > 0 {
		t, err := parseTransform(r.Transform)
		if err == nil {
			r.transform = t
		}
	}

	var l gokicadlib.Line
	return &l
}

func parseTransform(tstring string) (Transform, error) {
	var tm Transform
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
			}
		}
	}
}

func parseMatrix(l *lexer, t *Transform) error {
	i := l.nextItem()
	if i.typ != itemParan {
		return fmt.Errorf("Error Parsing Transform Matrix: Expected Parantheses")
	}
	var ncount int
	for {
		i = l.nextItem()
		for i.typ == itemComma || i.typ == itemWSP {
			i = l.nextItem()
		}
		if i.typ != itemNumber {
			return fmt.Errorf("Error Parsing Transform Matrix: Expected Number")
		}
		n, err := parseNumber(i)
		if err != nil {
			return err
		}
		t[ncount%2][ncount/3] = n
	}
}
