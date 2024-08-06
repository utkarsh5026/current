package bencode

import (
	"fmt"
	"sort"
	"strings"
)

type Encoder struct {
}

func (e *Encoder) Encode(value interface{}) (string, error) {
	return e.encodeValue(value)
}

// encodeValue encodes a given value into a bencoded string based on its type.
// The function supports encoding integers, strings, lists, and dictionaries.
//
// Parameters:
// - value: An interface{} representing the value to be encoded. The value can be of type int, string, []interface{}, or map[string]interface{}.
//
// Returns:
// - A string containing the bencoded representation of the value.
// - An error if the value type is unsupported or if there is an error during encoding.
func (e *Encoder) encodeValue(value interface{}) (string, error) {
	switch valType := value.(type) {
	case int:
		return e.encodeInt(valType), nil
	case string:
		return e.encodeString(valType), nil
	case []interface{}:
		return e.encodeList(valType)
	case map[string]interface{}:
		return e.encodeDict(valType)
	default:
		return "", fmt.Errorf("unsupported type: %T", value)
	}
}

// encodeInt encodes an integer value into a bencoded string.
func (e *Encoder) encodeInt(value int) string {
	return fmt.Sprintf("i%de", value)
}

// encodeString encodes a string value into a bencoded string.
func (e *Encoder) encodeString(s string) string {
	return fmt.Sprintf("%d:%s", len(s), s)
}

// encodeList encodes a list of values into a bencoded string.
// The list is expected to be prefixed with 'l' and suffixed with 'e'.
// For example, the list [123, "spam"] will be encoded to "li123e4:spame".
//
// Parameters:
// - list: A slice of interface{} representing the list to be encoded.
//
// Returns:
// - A string containing the bencoded representation of the list.
// - An error if any of the values in the list cannot be encoded.
func (e *Encoder) encodeList(list []interface{}) (string, error) {
	var builder strings.Builder
	builder.WriteByte(TypeList.Prefix())

	for _, value := range list {
		encoded, err := e.encodeValue(value)

		if err != nil {
			return "", err
		}
		builder.WriteString(encoded)
	}
	builder.WriteByte(TypeList.Suffix())
	return builder.String(), nil
}

// encodeDict encodes a dictionary into a bencoded string.
// The dictionary is expected to be prefixed with 'd' and suffixed with 'e'.
// For example, the dictionary {"foo": "bar"} will be encoded to "d3:foo3:bare".
//
// Parameters:
// - dict: A map[string]interface{} representing the dictionary to be encoded.
//
// Returns:
// - A string containing the bencoded representation of the dictionary.
// - An error if any of the values in the dictionary cannot be encoded.
func (e *Encoder) encodeDict(dict map[string]interface{}) (string, error) {
	var builder strings.Builder
	builder.WriteByte(TypeDict.Prefix())
	keys := make([]string, 0, len(dict))

	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		builder.WriteString(e.encodeString(key))
		encoded, err := e.encodeValue(dict[key])

		if err != nil {
			return "", err
		}
		builder.WriteString(encoded)
	}

	builder.WriteByte(TypeDict.Suffix())
	return builder.String(), nil
}
