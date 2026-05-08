package parser

import (
	"math"
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
			Type:        toPtr("postgres"),
			Host:        "127.0.0.1",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "db",
		},
	},
	"url - having only user & no password": {
		input: "redis://alice@awesome.redis.server:6380/0",
		expected: &connection{
			Type:        toPtr("redis"),
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
			Type:        toPtr("redis"),
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
			Type:        toPtr("postgres"),
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
			Type:        toPtr("postgres"),
			Host:        "127.0.0.1",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "users",
			Properties: map[string][]string{
				"sslmode":     {"prefer"},
				"search_path": {"public"},
				"charset":     {"utf8"},
			},
		},
	},
	"url - with multi scheme": {
		input: "postgresql+asyncpg://postgres:dina@example.com:5432/db",
		expected: &connection{
			Type:        toPtr("postgresql+asyncpg"),
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
	"url - repeated query parameter preserves all values in order": {
		input: "postgres://example.com:5432/db?opt=a&opt=b&opt=c",
		expected: &connection{
			Type:        toPtr("postgres"),
			Host:        "example.com",
			Port:        "5432",
			NumericPort: 5432,
			Database:    "db",
			Properties: map[string][]string{
				"opt": {"a", "b", "c"},
			},
		},
	},
	"url - mongodb readPreferenceTags repeated retains order": {
		input: "mongodb://host.example/?readPreference=secondary&readPreferenceTags=dc:east,rack:1&readPreferenceTags=dc:east&readPreferenceTags=",
		expected: &connection{
			Type: toPtr("mongodb"),
			Host: "host.example",
			Properties: map[string][]string{
				"readPreference":     {"secondary"},
				"readPreferenceTags": {"dc:east,rack:1", "dc:east", ""},
			},
		},
	},
	"url - empty userinfo (bare @) yields empty username pointer": {
		input: "redis://@host.example:6380/0",
		expected: &connection{
			Type:        toPtr("redis"),
			Username:    toPtr(""),
			Host:        "host.example",
			Port:        "6380",
			NumericPort: 6380,
			Database:    "0",
		},
	},
	"url - malformed escape returns parse error": {
		input:        "postgres://%zz/db",
		expectsError: true,
	},
	"url - root-only path leaves Database empty": {
		input: "postgres://example.com:5432/",
		expected: &connection{
			Type:        toPtr("postgres"),
			Host:        "example.com",
			Port:        "5432",
			NumericPort: 5432,
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
			Properties: map[string][]string{
				"sslmode":     {"prefer"},
				"search_path": {"public"},
				"charset":     {"utf8"},
			},
		},
	},
	"delimited - using type key": {
		input: "type=postgres host=example.com port=5432",
		expected: &connection{
			Type:        toPtr("postgres"),
			Host:        "example.com",
			Port:        "5432",
			NumericPort: 5432,
		},
	},
	"delimited - using scheme alias for type": {
		input: "scheme=mysql host=example.com port=3306",
		expected: &connection{
			Type:        toPtr("mysql"),
			Host:        "example.com",
			Port:        "3306",
			NumericPort: 3306,
		},
	},
	"delimited - repeated property key keeps both values in order": {
		input: "host=example.com tag=a tag=b tag=c",
		expected: &connection{
			Host: "example.com",
			Properties: map[string][]string{
				"tag": {"a", "b", "c"},
			},
		},
	},
	"delimited - using pass alias for password": {
		input: "user=alice pass=secret host=example.com",
		expected: &connection{
			Username: toPtr("alice"),
			Password: toPtr("secret"),
			Host:     "example.com",
		},
	},
	"delimited - using dbname alias for database": {
		input: "host=example.com dbname=users",
		expected: &connection{
			Host:     "example.com",
			Database: "users",
		},
	},
	"delimited - non-numeric port leaves NumericPort as zero": {
		input: "host=example.com port=abc",
		expected: &connection{
			Host: "example.com",
			Port: "abc",
		},
	},
	"delimited - bare key without equals goes to properties": {
		input: "host=example.com flag",
		expected: &connection{
			Host: "example.com",
			Properties: map[string][]string{
				"flag": {""},
			},
		},
	},
	"delimited - wrong delimiter causes error": {
		input:        "user=user password=password host=example.com port=5432 db=users",
		delimiter:    toPtr(rune(0)),
		expectsError: true,
	},
}

func TestRootLevelParseFunction(t *testing.T) {
	checks := make(map[string]dataProvider)
	for k, v := range urlChecks {
		checks[k] = v
	}

	for k, v := range delimitedStringChecks {
		checks[k] = v
	}

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

func TestConnectionStructHasUsernameAndPasswordMethod(t *testing.T) {
	conn1 := &connection{
		Host:     "example.com",
		Username: toPtr("alice"),
		Password: toPtr("password"),
	}

	assert.True(t, conn1.HasUsername())
	assert.True(t, conn1.HasPassword())

	conn2 := &connection{
		Host: "example.com",
	}

	assert.False(t, conn2.HasUsername())
	assert.False(t, conn2.HasPassword())
}

func TestConnectionStructHasPropertyMethod(t *testing.T) {
	conn := &connection{
		Host: "example.com",
		Properties: map[string][]string{
			"sslmode":     {"prefer"},
			"search_path": {"public"},
			"charset":     {"utf8"},
		},
	}

	assert.True(t, conn.HasProperty())
	assert.True(t, conn.HasProperty("sslmode"))
	assert.False(t, conn.HasProperty("not_in_property"))
}

func TestConnectionMethods(t *testing.T) {
	conn := &connection{
		Host: "example.com",
		Properties: map[string][]string{
			"sslmode":     {"prefer"},
			"search_path": {"public"},
			"charset":     {"utf8"},
		},
	}

	assert.Equal(t, "example.com", conn.Address(), "Address without port")

	// add port
	conn.Port = "5432"
	assert.Equalf(t, "example.com:5432", conn.Address(), "Address after adding port")

	assert.True(t, conn.HasProperty())
	assert.True(t, conn.GetProperty("sslmode") == "prefer", `get property should return "prefer"`)
	assert.True(t, conn.GetProperty("schema") == "", `get property should return empty string if property is not set`)
	assert.True(t, conn.GetProperty("schema", "public") == "public", `get property should return the default value (2nd parameter) if property is not set`)
}

func TestConnectionGetProperties(t *testing.T) {
	conn := &connection{
		Properties: map[string][]string{
			"readPreferenceTags": {"dc:east,rack:1", "dc:east", ""},
			"sslmode":            {"prefer"},
		},
	}

	assert.Equal(t, []string{"dc:east,rack:1", "dc:east", ""}, conn.GetProperties("readPreferenceTags"))
	assert.Equal(t, []string{"prefer"}, conn.GetProperties("sslmode"))
	assert.Nil(t, conn.GetProperties("missing"))

	// GetProperty on a multi-valued key returns the first value.
	assert.Equal(t, "dc:east,rack:1", conn.GetProperty("readPreferenceTags"))

	// GetProperty falls back when the slice exists but is empty.
	empty := &connection{Properties: map[string][]string{"k": {}}}
	assert.Equal(t, "", empty.GetProperty("k"))
	assert.Equal(t, "fallback", empty.GetProperty("k", "fallback"))
}

func TestNewParser(t *testing.T) {
	urlConn := &connection{
		Type:        toPtr("postgres"),
		Username:    toPtr("alice"),
		Password:    toPtr("bob"),
		Host:        "example.com",
		Port:        "5432",
		NumericPort: 5432,
		Database:    "users",
		Properties: map[string][]string{
			"sslmode":     {"prefer"},
			"search_path": {"public"},
			"charset":     {"utf8"},
		},
	}

	delimitedConn := *urlConn
	delimitedConn.Type = nil

	urlParser, urlParserErr := NewParser().FromUrl("postgres://alice:bob@example.com:5432/users?sslmode=prefer&search_path=public&charset=utf8")
	delimitedParser, delimitedParserErr := NewParser().Delimiter(';').FromPair("user=alice;password=bob;host=example.com;port=5432;db=users;sslmode=prefer;search_path=public;charset=utf8")

	assert.NoError(t, urlParserErr)
	assert.NoError(t, delimitedParserErr)

	assert.Equal(t, urlConn, urlParser)
	assert.Equal(t, &delimitedConn, delimitedParser)
}

func TestConnectionIsFor(t *testing.T) {
	t.Run("returns false when Type is nil", func(t *testing.T) {
		c := &connection{}
		assert.False(t, c.IsFor("postgres"))
		assert.False(t, c.IsFor("postgres", true))
	})

	t.Run("case-insensitive match by default", func(t *testing.T) {
		c := &connection{Type: toPtr("postgres")}
		assert.True(t, c.IsFor("postgres"))
		assert.True(t, c.IsFor("POSTGRES"))
		assert.True(t, c.IsFor("Postgres"))
	})

	t.Run("case-insensitive when sensitive flag is false", func(t *testing.T) {
		c := &connection{Type: toPtr("postgres")}
		assert.True(t, c.IsFor("POSTGRES", false))
	})

	t.Run("case-sensitive when sensitive flag is true", func(t *testing.T) {
		c := &connection{Type: toPtr("postgres")}
		assert.True(t, c.IsFor("postgres", true))
		assert.False(t, c.IsFor("Postgres", true))
		assert.False(t, c.IsFor("POSTGRES", true))
	})

	t.Run("returns false when types differ", func(t *testing.T) {
		c := &connection{Type: toPtr("postgres")}
		assert.False(t, c.IsFor("mysql"))
		assert.False(t, c.IsFor("mysql", true))
	})

	t.Run("only first sensitive arg is honored", func(t *testing.T) {
		c := &connection{Type: toPtr("postgres")}
		// extra args are ignored
		assert.True(t, c.IsFor("postgres", true, false))
		assert.True(t, c.IsFor("POSTGRES", false, true))
	})
}

func TestMakeSureNewConnectionFails(t *testing.T) {
	// To coverage 💯
	_, err := newConnection(map[string]interface{}{"x": math.NaN()})

	assert.Error(t, err)
}
