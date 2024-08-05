package bencode

import "fmt"

var StringOutOfBounds = fmt.Errorf("string out of bounds")

var InvalidFormat = func(t Type) error {
	format := t.String()
	return fmt.Errorf("invalid format: %s", format)
}

var MissingPrefix = func(t Type) error {
	return fmt.Errorf("missing prefix '%c' for type %s", t.Prefix(), t)
}

var MissingSuffix = func(t Type) error {
	return fmt.Errorf("missing suffix '%c'for type %s", t.Prefix(), t)
}
