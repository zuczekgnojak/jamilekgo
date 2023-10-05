package jamilek

import (
	"fmt"
	"errors"
	"unicode"
)

type Char struct {
	value rune
	row, column int
}

func (char Char) isLetter() bool {
	v := char.value
	return 0x41 <= v && v<= 0x5a || 0x61 <= v && v <= 0x7a
}

func (char Char) isDigit() bool {
	v := char.value
	return 0x30 <= v && v <= 0x39
}

func (char Char) isUnderscore() bool {
	return char.value == 0x5f
}

func (char Char) isKeyChar() bool {
	return char.isLetter() || char.isDigit() || char.isUnderscore()
}

func (char Char) isColon() bool {
	return char.value == 0x3a
}

func (char Char) isSpace() bool {
	return char.value == 0x20
}

func (char Char) isNewline() bool {
	return char.value == 0x0a
}

func (char Char) isQuotation() bool {
	return char.value == 0x22
}

func (char Char) isT() bool {
	return char.value == 0x74
}

func (char Char) isF() bool {
	return char.value == 0x66
}

func (char Char) isEscape() bool {
	return char.value == 0x5c
}

func (char Char) isPrintable() bool {
	return unicode.IsPrint(char.value)
}

func (char Char) isMinus() bool {
	return char.value == 0x2d
}

func (char Char) isZero() bool {
	return char.value == 0x30
}

func (char Char) isPositiveDigit() bool {
	return 0x31 <= char.value && char.value <= 0x39
}

func (char Char) isDot() bool {
	return char.value == 0x2e
}

func (char Char) isExp() bool {
	return char.value == 0x45
}

func (char Char) isArrayStart() bool {
	return char.value == 0x5b
}

func (char Char) isArrayEnd() bool {
	return char.value == 0x5d
}

func (char Char) isObjectStart() bool {
	return char.value == 0x7b
}

func (char Char) isObjectEnd() bool {
	return char.value == 0x7d
}

func (char Char) fromEscaped() (Char, error) {
	val := char.value
	var r rune = 0x00
	var err error = nil
	if val == 0x6e {
		r = 0x0a
	}
	if val == 0x5c {
		r = val
	}
	if val == 0x74 {
		r = 0x09
	}
	if r == 0x00 {
		err = errors.New("invalid escape code")
	}
	return Char{r, char.row, char.column}, err
}

func (c Char) String() string {
	format := "<non-print>(0x%[1]x):%[2]d:%[3]d"
	if unicode.IsPrint(c.value) {
		format = "'%[1]c'(0x%[1]x):%[2]d:%[3]d"
	}
	return fmt.Sprintf(format, c.value, c.row, c.column)
}
