package svg2kicadlib

import (
	"fmt"

	mt "github.com/rustyoz/Mtransform"
	"github.com/rustyoz/gokicadlib"
)

type Rect struct {
	Id        string `xml:"id,attr"`
	Width     string `xml:"width,attr"`
	height    string `xml:"height,attr"`
	Transform string `xml:"transform,attr"`

	transform mt.Transform
	group     *Group
}

func (r *Rect) ToKicad() (ses gokicadlib.SExpression) {
	if len(r.Transform) > 0 {
		t, err := parseTransform(r.Transform)
		if err == nil {

			r.transform = t
		} else {
			fmt.Println(r.Transform)
		}
	}

	var l gokicadlib.Line
	return &l
}
