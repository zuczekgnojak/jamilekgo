package jamilek

import "testing"

func TestTrue(t *testing.T) {
	if false {
		t.Fatalf("dupa")
	}
}

func makeChar(r rune) Char {
	return Char{r, 0, 0}
}

func TestFixedChars(t *testing.T) {
	c := makeChar(':')
	if !c.isColon() { t.Fatalf("colon") }

	c = makeChar(' ')
	if !c.isSpace() { t.Fatalf("space") }

	c = makeChar('\n')
	if !c.isNewline() { t.Fatalf("newline") }

	c = makeChar('"')
	if !c.isQuotation() { t.Fatalf("quotation") }

	c = makeChar('t')
	if !c.isT() { t.Fatalf("t") }

	c = makeChar('f')
	if !c.isF() { t.Fatalf("f") }

	c = makeChar('\\')
	if !c.isEscape() { t.Fatalf("escape") }

	c = makeChar('-')
	if !c.isMinus() { t.Fatalf("minus") }

	c = makeChar('0')
	if !c.isZero() { t.Fatalf("zero") }

	c = makeChar('.')
	if !c.isDot() { t.Fatalf("dot") }

	c = makeChar('E')
	if !c.isExp() { t.Fatalf("exp") }
}


func TestDigits(t *testing.T) {
	digits := []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for _, digit := range digits {
		char := makeChar(digit)
		if !char.isDigit() {
			t.Fatalf("%c error", digit)
		}
	}

	nondigits := []rune{'d', 'b', '#', '\n', '\\', '!', 'h'}
	for _, nondigit := range nondigits {
		char := makeChar(nondigit)
		if char.isDigit() {
			t.Fatalf("%c is not digit", nondigit)
		}
	}
}
