package bencode

import (
	"fmt"
	"strconv"
)

type Decoder struct {
	input string // bencoded input
	index int    // current index in the input
}

// NewDecoder creates a new Decoder with the specified input.
func NewDecoder(input string) *Decoder {
	return &Decoder{input: input}
}

// move moves the index by the specified number of steps.
func (d *Decoder) move(steps int) {
	d.index += steps
}

// moveNext moves the index to the next character in the input.
func (d *Decoder) moveNext() {
	d.move(1)
}

// peek returns the next character in the input without consuming it.
func (d *Decoder) peek() byte {
	if d.index < len(d.input) {
		return d.input[d.index]
	}
	return 0
}

// hasMore checks if there are more characters to read in the input.
func (d *Decoder) hasMore() bool {
	return d.index < len(d.input)
}

// Decode decodes the next bencoded value from the input.
// It determines the type of the value by peeking at the current character
// and then delegates the decoding to the appropriate method.
//
// Returns the decoded value as an interface{} and an error if the format is invalid.
func (d *Decoder) Decode() (interface{}, error) {
	if !d.hasMore() {
		return nil, fmt.Errorf("unexpected end of input")
	}

	switch d.peek() {
	case TypeInt.Prefix():
		return d.decodeInt()
	case TypeList.Prefix():
		return d.decodeList()
	case TypeDict.Prefix():
		return d.decodeDict()
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

// decodeString decodes a string from the bencoded input.
// The string is expected to be prefixed with its length followed by a colon.
// For example, the bencoded string "4:spam" will be decoded to "spam".
//
// Returns the decoded string and an error if the format is invalid.
func (d *Decoder) decodeString() (string, error) {
	start := d.index
	for d.hasMore() && d.peek() != TypeString.Suffix() {
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

// decodeList decodes a list from the bencoded input.
// The list is expected to be prefixed with 'l' and suffixed with 'e'.
// For example, the bencoded list "li123ee" will be decoded to [123].
// An empty list "le" will be decoded to [].
//
// Returns the decoded list and an error if the format is invalid.
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

// decodeDict decodes a dictionary from the bencoded input.
// The dictionary is expected to be prefixed with 'd' and suffixed with 'e'.
// For example, the bencoded dictionary "d3:foo3:bare" will be decoded to {"foo": "bar"}.
//
// Returns the decoded dictionary and an error if the format is invalid.
func (d *Decoder) decodeDict() (map[string]interface{}, error) {
	if d.peek() != TypeDict.Prefix() {
		return nil, MissingPrefix(TypeDict)
	}

	d.moveNext()

	dict := make(map[string]interface{})
	for d.hasMore() && d.peek() != TypeDict.Suffix() {
		key, err := d.decodeString()
		if err != nil {
			return nil, err
		}

		value, err := d.Decode()
		if err != nil {
			return nil, err
		}

		dict[key] = value
	}

	if !d.hasMore() {
		return nil, MissingSuffix(TypeDict)
	}
	d.moveNext()

	return dict, nil
}
