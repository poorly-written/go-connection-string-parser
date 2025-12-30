package connection_string

import (
	"maps"
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
	delimiter    *rune
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

var delimitedStringChecks = map[string]dataProvider{
	"delimited - empty input string": {
		input:    "",
		expected: &connection{},
	},
	"delimited - having no user and password": {
		input: "host=127.0.0.1 port=5432 db=db",
		expected: &connection{
			Host:        "127.0.0.1",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "db",
		},
	},
	"delimited - having only user & no password": {
		input: "user=alice host=awesome.redis.server port=6380 db=0",
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    nil,
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - having user and empty password": {
		input: "user=alice password= host=awesome.redis.server port=6380 db=0",
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr(""),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - having delimiter in input string": {
		input: `user=alice "password=pass word" host=awesome.redis.server port=6380 db=0`,
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr("pass word"),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - having equal operator in input string": {
		input: "user=alice password=pass=word host=awesome.redis.server port=6380 db=0",
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr("pass=word"),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - using different delimiter": {
		input:     "user=alice;password=password;host=awesome.redis.server;port=6380;db=0",
		delimiter: toPtr(';'),
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr("password"),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - trims space around the key": {
		input:     " user=alice; password =password; host =awesome.redis.server;port =6380;db =0",
		delimiter: toPtr(';'),
		expected: &connection{
			Username:    toPtr("alice"),
			Password:    toPtr("password"),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - does not trim space around the value": {
		input:     " user= alice ; password = password ; host =awesome.redis.server;port =6380;db =0",
		delimiter: toPtr(';'),
		expected: &connection{
			Username:    toPtr(" alice "),
			Password:    toPtr(" password "),
			Host:        "awesome.redis.server",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"delimited - skips empty strings caused by delimiter": {
		// additional space before port
		input: "host=example.com  port=5432",
		expected: &connection{
			Host:        "example.com",
			Port:        "5432",
			NumericPort: 5432,
			// Database:    "",
		},
	},
	"delimited - with properties": {
		input: "user=user password= host=example.com port=5432 db=users sslmode=prefer search_path=public charset=utf8",
		expected: &connection{
			Username:    toPtr("user"),
			Password:    toPtr(""),
			Host:        "example.com",
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
	"delimited - wrong delimiter causes error": {
		input:        "user=user password=password host=example.com port=5432 db=users",
		delimiter:    toPtr(rune(0)),
		expectsError: true,
		expected: &connection{
			Username:    toPtr("user"),
			Password:    toPtr(""),
			Host:        "example.com",
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
}

func TestRootLevelParseFunction(t *testing.T) {
	checks := make(map[string]dataProvider)
	maps.Insert(checks, maps.All(urlChecks))
	maps.Insert(checks, maps.All(delimitedStringChecks))

	for name, testCase := range checks {
		t.Run(name, func(t *testing.T) {
			var conn *connection
			var err error
			if testCase.delimiter != nil {
				conn, err = Parse(testCase.input, *testCase.delimiter)
			} else {
				conn, err = Parse(testCase.input)
			}

			if testCase.expectsError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			assert.Equal(t, testCase.expected, conn)
		})
	}
}
