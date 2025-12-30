package connection_string

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*type expectation[T any] func(actual T) bool

func extractValue(i reflect.Value) reflect.Value {
	if i.Type().Kind() == reflect.Ptr {
		return i.Elem()
	}

	return i
}

func shouldBeNull(actual any) bool {
	return reflect.ValueOf(actual).IsNil()
}

func shouldBeEmpty(actual any) bool {
	return reflect.ValueOf(actual).IsZero()
}

func exactly[T any](expected T) expectation[T] {
	return func(actual T) bool {
		spew.Dump(expected, actual)
		av := reflect.ValueOf(actual)
		if av.Comparable() {
			ev := reflect.ValueOf(expected)
			return av.Equal(ev)
		}

		return false
	}
}*/

func toPtr[T any](v T) *T {
	return &v
}

type dataProvider struct {
	input        string
	expectsError bool
	expected     *connection
}

var urlChecks = map[string]dataProvider{
	"url - empty input string": {
		input:    "",
		expected: &connection{},
	},
	"url - having no user and password": {
		input: "postgres://127.0.0.1:5432/db",
		expected: &connection{
			Host:        "127.0.0.1",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "db",
		},
	},
	"url - having only user & no password": {
		input: "redis://alice@awesome.redis.server:6380/0",
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    nil,
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"url - having user and empty password": {
		input: "redis://alice:@awesome.redis.server:6380/0",
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr(""),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"url - not encoded user password": {
		input:        "postgres://user:abc{DEf1=ghi@example.com:5432/db",
		expectsError: true,
	},
	"url - percent encoded user password": {
		input: "postgres://user:abc%7BDEf1=ghi@example.com:5432/db",
		expected: &connection{
			Username:    toPtr("user"),
			Password:    toPtr("abc{DEf1=ghi"),
			Host:        "example.com",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "db",
		},
	},
	"url - having parameters": {
		input: "postgres://127.0.0.1:5432/users?sslmode=prefer&search_path=public&charset=utf8",
		expected: &connection{
			Host:        "127.0.0.1",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "users",
			Properties: map[string]string{
				"sslmode":     "prefer",
				"search_path": "public",
				"charset":     "utf8",
			},
		},
	},
	"url - with multi scheme": {
		input: "postgresql+asyncpg://postgres:dina@example.com:5432/db",
		expected: &connection{
			Username:    toPtr("postgres"),
			Password:    toPtr("dina"),
			Host:        "example.com",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "db",
		},
	},
	"url - with no scheme": {
		input: "//alice:@awesome.redis.server:6380/0",
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr(""),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
}

func TestRootLevelParser(t *testing.T) {
	for name, testCase := range urlChecks {
		t.Run(name, func(t *testing.T) {
			conn, err := Parse(testCase.input)

			if testCase.expectsError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			assert.Equal(t, testCase.expected, conn)
		})
	}
}
