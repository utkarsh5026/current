package bencode

import (
	"fmt"
	"strconv"
)

type Decoder struct {
	input string
	index int
}

func NewDecoder(input string) *Decoder {
	return &Decoder{input: input}
}

func (d *Decoder) move(steps int) {
	d.index += steps
}

func (d *Decoder) moveNext() {
	d.move(1)
}

func (d *Decoder) peek() byte {
	if d.index < len(d.input) {
		return d.input[d.index]
	}
	return 0
}

func (d *Decoder) hasMore() bool {
	return d.index < len(d.input)
}

func (d *Decoder) Decode() (interface{}, error) {
	if !d.hasMore() {
		return nil, fmt.Errorf("unexpected end of input")
	}

	switch d.peek() {
	case TypeInt.Prefix():
		return d.decodeInt()
	case TypeList.Prefix():
		return d.decodeList()
	default:
		return d.decodeString()
	}
}

// decodeInt decodes an integer from the bencoded input.
// The integer is expected to be prefixed with 'i' and suffixed with 'e'.
// For example, the bencoded integer "i123e" will be decoded to 123.
//
// Returns the decoded integer and an error if the format is invalid.
func (d *Decoder) decodeInt() (int, error) {
	if d.peek() != TypeInt.Prefix() {
		return 0, MissingPrefix(TypeInt)
	}

	d.moveNext() // consume 'i'
	start := d.index
	for d.hasMore() && d.peek() != 'e' {
		d.moveNext()
	}

	if !d.hasMore() {
		return 0, MissingSuffix(TypeInt)
	}

	numStr := d.input[start:d.index]
	d.moveNext() // consume 'e'
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, InvalidFormat(TypeInt)
	}
	return num, nil
}

func (d *Decoder) decodeString() (string, error) {
	start := d.index
	for d.hasMore() && d.peek() != ':' {
		d.moveNext()
	}

	if !d.hasMore() {
		return "", MissingSuffix(TypeString)
	}

	length := d.input[start:d.index]
	d.moveNext() // consume ':'

	l, err := strconv.Atoi(length)
	if err != nil {
		return "", err
	}

	if d.index+l > len(d.input) {
		return "", StringOutOfBounds
	}

	str := d.input[d.index : d.index+l]
	d.move(l)
	return str, nil
}

func (d *Decoder) decodeList() ([]interface{}, error) {
	if d.peek() != TypeList.Prefix() {
		return nil, MissingPrefix(TypeList)
	}

	d.moveNext() // consume 'l'

	var list []interface{}
	for d.hasMore() && d.peek() != TypeList.Suffix() {
		item, err := d.Decode()
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	if !d.hasMore() {
		return nil, MissingSuffix(TypeList)
	}

	d.moveNext()
	return list, nil
}
