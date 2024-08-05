package bencode

type Type int

const (
	TypeInt Type = iota
	TypeString
	TypeList
	TypeDict
)

func (t Type) String() string {
	switch t {
	case TypeInt:
		return "int"
	case TypeString:
		return "string"
	case TypeList:
		return "list"
	case TypeDict:
		return "dictionary"
	}
	return "unknown"
}

func (t Type) Prefix() byte {
	switch t {
	case TypeInt:
		return 'i'
	case TypeString:
		return 's'
	case TypeList:
		return 'l'
	case TypeDict:
		return 'd'
	}
	return 'u'
}

func (t Type) Suffix() byte {
	switch t {
	case TypeInt:
		return 'e'
	case TypeString:
		return ':'
	case TypeList:
		return 'e'
	case TypeDict:
		return 'e'
	}
	return 'u'
}
