package formattypes

type FormatType string

const (
	FormatJSON    FormatType = "json"
	FormatConsole FormatType = "console"
)

func (f FormatType) String() string {
	return string(f)
}

func (f FormatType) IsValid() bool {
	switch f {
	case FormatJSON, FormatConsole:
		return true
	default:
		return false
	}
}

func FromString(format string) FormatType {
	ft := FormatType(format)
	if ft.IsValid() {
		return ft
	}
	return FormatJSON
}
