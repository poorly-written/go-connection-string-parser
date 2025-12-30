package connection_string

import "strings"

type config struct{}

var defaultMapping = map[string]string{
	"user_id":  "user",
	"user":     "user",
	"password": "password",
	"pass":     "password",
}

var defaultDelimiter = ';'

type parser struct {
	mapping   map[string]string
	delimiter rune
}

func (p *parser) Delimiter(delimiter rune) *parser {
	p.delimiter = delimiter

	return p
}

func (p *parser) OverwriteMapping(mapping map[string]string) *parser {
	p.mapping = mapping

	return p
}

func (p *parser) AdditionalMapping(mapping map[string]string) *parser {
	for k, v := range mapping {
		p.mapping[k] = v
	}

	return p
}

func (p *parser) FromUrl(input string) (*config, error) {
	return nil, nil
}

func (p *parser) FromDelimitedString(input string) (*config, error) {
	return nil, nil
}

func (p *parser) Parse(input string) (*config, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return &config{}, nil
	}

	if strings.Contains(input, "://") {
		return p.FromUrl(input)
	}

	return p.FromDelimitedString(input)
}

func NewParser() *parser {
	return &parser{
		mapping:   defaultMapping,
		delimiter: defaultDelimiter,
	}
}

func Parse(input string, delimiters ...rune) (*config, error) {
	p := NewParser()

	if len(delimiters) > 0 {
		p = p.Delimiter(delimiters[0])
	}

	return p.Parse(input)
}
