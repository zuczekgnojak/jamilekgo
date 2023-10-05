package jamilek

import (
	"fmt"
	"io"
	"errors"
)

type ParseError struct {
	message string
	char Char
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s, %s", e.char, e.message)
}

type Parser struct {
	runeReader io.RuneReader
	row, column int
}

func Parse(reader io.RuneReader) (*Node, error) {
	parser := Parser{reader, 0, 0}
	return parser.Parse()
}

func (p *Parser) Parse() (*Node, error) {
	rootNode, err := p.parseRoot()
	return rootNode, err
}

func (c *Parser) Read() (Char, error) {
	value, _, err := c.runeReader.ReadRune()
	char := Char{value, c.row, c.column}
	c.column += 1
	if value == '\n' {
		c.column = 0
		c.row += 1
	}
	return char, err
}

func (p *Parser) parseRoot() (*Node, error) {
	rootObject := make(map[string]*Node)
	previousKey := ""
	for {
		key, err := p.parseKey()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if key < previousKey {
			return nil, errors.New("keys in disorder")
		}
		previousKey = key
		err = p.parseSpace()
		if err != nil {
			return nil, err
		}
		value, err := p.parseValue(0)
		if err != nil {
			return nil, err
		}
		rootObject[key] = value
	}
	node := &Node{Object, rootObject}
	return node, nil
}

func (p *Parser) parseKey() (string, error) {
	result := ""
	for {
		char, err := p.Read()
		if err != nil {
			return "", err
		}
		if char.isKeyChar() {
			result += string(char.value)
			continue
		}
		if char.isColon() && len(result) > 0 {
			return result, nil
		}
		return "", ParseError{"invalid key character", char}
	}
}

func (p *Parser) parseValue(level int) (*Node, error) {
	char, err := p.Read()
	if err != nil {
		return nil, err
	}
	if char.isQuotation() {
		return p.parseString()
	}
	if char.isT() {
		return p.parseTrue()
	}
	if char.isF() {
		return p.parseFalse()
	}
	if char.isMinus() || char.isDigit() {
		return p.parseNumber(char)
	}
	if char.isArrayStart() {
		return p.parseArray(level)
	}
	if char.isObjectStart() {
		return p.parseObject(level)
	}
	return nil, ParseError{"invalid value", char}
}

func (p *Parser) parseObject(level int) (*Node, error) {
	object := make(map[string]*Node)
	previousKey := ""

	err := p.parseNewline()
	if err != nil { return nil, err }

	for {
		isObjectEnd, err := p.parseObjectIndent(level)
		if err != nil { return nil, err }
		if isObjectEnd { break }

		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		if key < previousKey {
			return nil, errors.New("keys in disorder")
		}
		previousKey = key
		err = p.parseSpace()
		if err != nil { return nil, err }
		value, err := p.parseValue(level+1)
		if err != nil { return nil, err }
		object[key] = value
	}

	err = p.parseNewline()
	if err != nil { return nil, err }

	return &Node{Object, object}, nil
}

func (p *Parser) parseObjectIndent(level int) (bool, error) {
	for i := 0; i < level; i++ {
		err := p.parseSpace()
		if err != nil { return false, err }
		err = p.parseSpace()
		if err != nil { return false, err }
	}
	char, err := p.Read()
	if err != nil { return false, err }
	if char.isObjectEnd() { return true, nil }
	if char.isSpace() {
		err = p.parseSpace()
		if err != nil { return false, err }
		return false, nil
	}
	return false, ParseError{ "invalid char", char}
}

func (p *Parser) parseArray(level int) (*Node, error) {
	value := make([]*Node, 0)
	err := p.parseNewline()
	if err != nil { return nil, err }

	for {
		isArrayEnd, err := p.parseArrayIndent(level)
		if err != nil { return nil, err }
		if isArrayEnd { break }

		node, err := p.parseValue(level+1)
		if err != nil { return nil, err }
		value = append(value, node)
	}

	err = p.parseNewline()
	if err != nil { return nil, err }

	return &Node{Array, value}, nil
}

func (p *Parser) parseArrayIndent(level int) (bool, error) {
	for i := 0; i < level; i++ {
		err := p.parseSpace()
		if err != nil { return false, err }
		err = p.parseSpace()
		if err != nil { return false, err }
	}
	char, err := p.Read()
	if err != nil { return false, err }
	if char.isArrayEnd() { return true, nil }
	if char.isSpace() {
		err = p.parseSpace()
		if err != nil { return false, err }
		return false, nil
	}
	return false, ParseError{ "invalid char", char}
}

func (p *Parser) parseNumber(firstChar Char) (*Node, error) {
	if firstChar.isZero() {
		return p.parseZeros()
	}
	if firstChar.isMinus() {
		return p.parseNegativeNumber()
	}
	return p.parsePositiveNumber(firstChar)
}

func (p *Parser) parseZeros() (*Node, error) {
	// one zero has already been read
	char, err := p.Read()
	if err != nil {return nil, err}
	if char.isNewline() {
		return GetIntegerNode("0")
	}
	if !char.isDot() {
		return nil, ParseError{"invalid number", char}
	}
	// there's zero and dot, we need one more zero
	char, err = p.Read()
	if err != nil { return nil, err }
	if !char.isZero() {
		return nil, ParseError{"expecting 0", char}
	}
	// lastly, there should be a nl
	err = p.parseNewline()
	if err != nil { return nil, err}

	return GetFloatNode("0.0")
}

func (p *Parser) parseNegativeNumber() (*Node, error) {
	result := "-"
	// only positive digit allowed after -
	char, err := p.readPositiveDigit()
	if err != nil { return nil, err }
	result += string(char.value)

	// we might get a dot or a digit here
	char, err = p.Read()
	if err != nil { return nil, err }

	// it's one digit, negative int
	if char.isNewline() {
		return GetIntegerNode(result)
	}

	// it's multidigit negative int

	if char.isDigit() {
		result += string(char.value)
		return p.parseInt(result)
	}

	// expecting dot from now on
	if !char.isDot() {
		return nil, ParseError{"invalid char in number", char}
	}

	// expecting float
	return p.parseFloat(result)
}

func (p *Parser) parsePositiveNumber(firstChar Char) (*Node, error) {
	result := string(firstChar.value)

	char, err := p.Read()
	if err != nil {return nil, err}

	// singele digit positive int
	if char.isNewline() {
		return GetIntegerNode(result)
	}

	// multidigit int
	if char.isDigit() {
		result += string(char.value)
		return p.parseInt(result)
	}

	// positive float
	if char.isDot() {
		return p.parseFloat(result)
	}

	return nil, ParseError{"expected digit or dot", char}
}

func (p *Parser) parseFloat(result string) (*Node, error) {
	// there is a dot already in result, parsing reminder
	result += "."
	// at least on digit after dot
	char, err := p.readDigit()
	if err != nil {return nil, err}
	result += string(char.value)

	// more then one digit after dot, need to check if ends with 0
	lastChar := Char{}
	for {
		currentChar, err := p.Read()
		if err != nil {return nil, err}

		if lastChar.isZero() && (currentChar.isNewline() || currentChar.isExp()) {
			return nil, ParseError{"invalid float", currentChar}
		}

		if currentChar.isNewline() {
			return GetFloatNode(result)
		}
		if currentChar.isExp() {
			return p.parseExp(result)
		}
		if !currentChar.isDigit() {
			return nil, ParseError{"expecting digit", currentChar}
		}
		lastChar = currentChar
		result += string(currentChar.value)
	}
}

func (p *Parser) parseExp(result string) (*Node, error) {
	result += "E"
	char, err := p.Read()
	if err != nil { return nil, err }
	if !char.isMinus() && !char.isPositiveDigit() {
		return nil, ParseError{"invalid char", char}
	}
	result += string(char.value)

	for {
		char, err = p.Read()
		if err != nil { return nil, err }
		if char.isNewline() { break }
		if !char.isDigit() {
			return nil, ParseError{"invalid char", char}
		}
		result += string(char.value)
	}


	return GetFloatNode(result)
}

func (p *Parser) readDigit() (Char, error) {
	char, err := p.Read()
	if err != nil {
		return char, err
	}
	if !char.isDigit() {
		return char, ParseError{"expecting digit", char}
	}

	return char, nil
}

func (p *Parser) readPositiveDigit() (Char, error) {
	char, err := p.readDigit()
	if err != nil { return char, err }
	if !char.isPositiveDigit() {
		return char, ParseError{"expecting positive digit", char}
	}
	return char, nil
}

func (p *Parser) parseInt(result string) (*Node, error) {
	for {
		char, err := p.Read()
		if err != nil {return nil, err}

		if !char.isDigit() && !char.isNewline() {
			return nil, ParseError{"invalid int character", char}
		}

		if char.isNewline() {
			return GetIntegerNode(result)
		}

		result += string(char.value)
	}
}

func (p *Parser) parseString() (*Node, error) {
	result := ""
	for {
		char, err := p.Read()
		if !char.isPrintable() && !char.isSpace() {
		    return nil, ParseError{"invalid character in string", char}
		}
		if err != nil {
			return nil, err
		}
		if char.isQuotation() {
			err = p.parseNewline()
			if err != nil {
				return nil, err
			}
			return &Node{String, result}, nil
		}
		if char.isEscape() {
			char, err = p.Read()
			if err != nil { return nil, err }
			char, err = char.fromEscaped()
			if err != nil {return nil, err}
		}

		result += string(char.value)
	}
}

func (p *Parser) parseTrue() (*Node, error) {
	runes := []rune{'r', 'u', 'e'}
	for _, r := range runes {
		char, err := p.Read()
		if err != nil {
			return nil, err
		}
		if char.value != r {
			return nil, ParseError{"invalid character while parsing 'true'", char}
		}
	}
	err := p.parseNewline()
	if err != nil {
		return nil, err
	}
	return &Node{Bool, true}, nil
}

func (p *Parser) parseFalse() (*Node, error) {
	runes := []rune{'a', 'l', 's', 'e'}
	for _, r := range runes {
		char, err := p.Read()
		if err != nil {
			return nil, err
		}
		if char.value != r {
			return nil, ParseError{"invalid character while parsing 'false'", char}
		}
	}
	err := p.parseNewline()
	if err != nil {
		return nil, err
	}
	return &Node{Bool, false}, nil
}

func (p *Parser) parseSpace() error {
	char, err := p.Read()
	if err != nil {
		return err
	}
	if char.isSpace() {
		return nil
	}
	return ParseError{"expected space", char}
}

func (p *Parser) parseNewline() error {
	char, err := p.Read()
	if err != nil {
		return err
	}
	if char.isNewline() {
		return nil
	}
	return ParseError{"expected newline", char}
}
