package svg2kicadlib

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemError:
		return "Error"
	case itemComma:
		return "Comma"
	case itemParan:
		return "Parentheses"
	case itemLetter:
		return fmt.Sprint("Letter ", i.val)
	case itemNumber:
		return fmt.Sprint("Number ", i.val)
	case itemWSP:
		return "WSP"
	default:
		return i.val
	}
}

type lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan item
}

type itemType int

const (
	itemError itemType = iota
	itemDot
	itemEOS
	itemLetter
	itemWord
	itemNumber
	itemComma
	itemFlag
	itemWSP
	itemParan
)

type stateFn func(*lexer) stateFn

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l, l.items
}

const eof = -1

func (l *lexer) run() {
	for state := lexD; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) nextItem() item {
	item := <-l.items
	return item
}

func lexNumber(l *lexer) stateFn {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	l.emit(itemNumber)
	return lexD
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) emit(t itemType) {
	i := item{t, l.input[l.start:l.pos]}
	//fmt.Println(i)
	l.items <- i
	l.start = l.pos
}

func lexWord(l *lexer) stateFn {
	l.acceptRun("abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ")
	l.emit(itemWord)
	return lexD
}

func lexComma(l *lexer) stateFn {
	l.accept(",")
	l.emit(itemComma)
	return lexD
}

func isWSP(r rune) bool {
	return r == ' ' || r == '\t'
}

func lexWSP(l *lexer) stateFn {
	l.accept(" \t\r\n\f")
	l.emit(itemWSP)
	return lexD
}

func lexD(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			l.emit(itemEOS)
			return nil
		case isWSP(r):
			return lexWSP
		case unicode.IsLetter(r):
			return lexWord
		case r == '-' || r == '+':
			return lexNumber
		case unicode.IsNumber(r):
			return lexNumber
		case r == ',':
			return lexComma
		case r == '(' || r == ')':
			return lexParan
		}
	}
	return nil
}

func lexParan(l *lexer) stateFn {
	l.accept("()")
	l.emit(itemParan)
	return lexD
}
