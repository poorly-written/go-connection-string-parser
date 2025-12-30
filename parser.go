package connection_string

import "strings"

type config struct{}

var defaultDelimiter = ' '

type parser struct {
	delimiter rune
}

func (p *parser) Delimiter(delimiter rune) *parser {
	p.delimiter = delimiter

	return p
}

func (p *parser) FromUrl(input string) (*config, error) {
	return nil, nil
}

func (p *parser) FromPair(input string) (*config, error) {
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

	return p.FromPair(input)
}

func NewParser() *parser {
	return &parser{
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
