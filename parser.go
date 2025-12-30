package connection_string

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const keyScheme = "scheme"
const keyUsername = "username"
const keyPassword = "password"
const keyHost = "host"
const keyPort = "port"
const keyNumericPort = "numeric_port"
const keyDatabase = "database"
const keyProperties = "properties"

var defaultDelimiter = ' '

type connection struct {
	Username    *string           `json:"username,omitempty"`
	Password    *string           `json:"password,omitempty"`
	Host        string            `json:"host"`
	Port        string            `json:"port"`
	NumericPort int               `json:"numeric_port"`
	Database    string            `json:"database"`
	Properties  map[string]string `json:"properties,omitempty"`
}

func (c *connection) Address() string {
	if c.Port != "" {
		return fmt.Sprintf("%s:%s", c.Host, c.Port)
	}

	return c.Host
}

func (c *connection) HasProperties() bool {
	return len(c.Properties) > 0
}

func newConnection(data map[string]interface{}) (*connection, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	c := &connection{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}

	return c, nil
}

type parser struct {
	delimiter rune
}

func (p *parser) Delimiter(delimiter rune) *parser {
	p.delimiter = delimiter

	return p
}

func (p *parser) FromUrl(input string) (*connection, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	properties := make(map[string]string)

	if u.Scheme != "" {
		data[keyScheme] = u.Scheme
	}

	if u.User != nil {
		data[keyUsername] = u.User.Username()

		if password, ok := u.User.Password(); ok {
			data[keyPassword] = password
		}
	}

	if u.Host != "" {
		data[keyHost] = u.Hostname()
	}

	if port := u.Port(); port != "" {
		data[keyPort] = port
		if numericPort, err := strconv.Atoi(port); err == nil {
			data[keyNumericPort] = numericPort
		}
	}

	if path := strings.TrimLeft(u.Path, "/"); path != "" {
		data[keyDatabase] = path
	}

	if queries := u.Query(); len(queries) > 0 {
		for k, v := range queries {
			properties[k] = v[0]
		}

		data[keyProperties] = properties
	}

	return newConnection(data)
}

func (p *parser) FromPair(input string) (*connection, error) {
	return nil, nil
}

func (p *parser) Parse(input string) (*connection, error) {
	if input == "" {
		return &connection{}, nil
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

func Parse(input string, delimiters ...rune) (*connection, error) {
	p := NewParser()

	if len(delimiters) > 0 {
		p = p.Delimiter(delimiters[0])
	}

	return p.Parse(input)
}
