poorly-written/go-connection-string-parser
[![codecov](https://codecov.io/github/poorly-written/go-connection-string-parser/graph/badge.svg?token=BNH1MVJQHP)](https://codecov.io/github/poorly-written/go-connection-string-parser)
[![Go Version](https://img.shields.io/github/go-mod/go-version/poorly-written/go-connection-string-parser)](https://github.com/poorly-written/go-connection-string-parser/blob/main/go.mod)
[![Latest Tag](https://img.shields.io/github/v/tag/poorly-written/go-connection-string-parser)](https://github.com/poorly-written/go-connection-string-parser/tags)
===

`go-connection-string-parser` parses a database connection string into a small Go struct. It supports the two most common
shapes — URL style (`postgres://user:pass@host:5432/db`) and key/value pair style (`host=example.com port=5432`) — with
one tiny API.

# Installation

To install the package, run

> `go get github.com/poorly-written/go-connection-string-parser`

# Documentation

## Quick start

The package-level `Parse` function looks at the input and decides the format for you.

```go
package main

import (
    "fmt"

    parser "github.com/poorly-written/go-connection-string-parser"
)

func main() {
    conn, err := parser.Parse("postgres://alice:secret@example.com:5432/users?sslmode=prefer")
    if err != nil {
        panic(err)
    }

    fmt.Println(*conn.Type)                  // postgres
    fmt.Println(conn.Address())              // example.com:5432
    fmt.Println(conn.GetProperty("sslmode")) // prefer
}
```

`Parse` picks the right branch based on the input:

- An input that contains `://`, or starts with `//`, is parsed as a URL.
- Anything else is parsed as a delimited key/value string. The default delimiter is a space.

You can also pass a custom delimiter as a second argument.

```go
conn, err := parser.Parse("user=alice;password=secret;host=example.com", ';')
```

## Parser

If you want full control, build a parser instance and call its methods directly.

- `parser.NewParser()` returns a new parser with the default delimiter (space).
- `Delimiter(rune)` changes the delimiter. It returns the same parser, so calls can be chained.
- `Parse(string)` parses the input — same auto-detect rules as the package-level `Parse`.
- `FromUrl(string)` parses the input as a URL.
- `FromPair(string)` parses the input as a delimited key/value string.

```go
p := parser.NewParser()
conn, err := p.Parse("redis://alice@cache.local:6379/0")

p := parser.NewParser().Delimiter(';')
conn, err := p.FromPair("user=alice;password=secret;host=example.com")

p := parser.NewParser()
conn, err := p.FromUrl("mysql://root:root@127.0.0.1:3306/app?charset=utf8mb4")
```

### URL form

`FromUrl` is built on Go's standard `net/url` package. It returns the underlying parse error if the input is not a valid
URL — for example, an invalid percent-escape (`%zz`) or a non-numeric port.

The query string is read with `url.Values`, so a key that appears more than once keeps every value, in the order the
keys appeared in the URL.

### Delimited form

`FromPair` reads the input through Go's `encoding/csv` reader with `LazyQuotes` enabled. This means you can:

- wrap a value in double quotes if it contains the delimiter — `"password=pass word"` keeps `pass word` as one value,
- include the equals sign inside a value — `password=pass=word` keeps `pass=word` as the password,
- pick any rune as the delimiter — space, semicolon, comma, pipe, tab, and so on.

Leading and trailing space around a **key** is trimmed. Space around a **value** is **not** trimmed — `password = secret`
keeps the value as `" secret"`.

A key without an `=` sign (for example, a bare flag) is treated as a property with an empty value.

#### Recognised keys

The parser knows a few well-known field names plus their common aliases. Anything else goes into `Properties`.

| Field      | Recognised keys              |
|------------|------------------------------|
| `Type`     | `type`, `scheme`             |
| `Username` | `username`, `user`           |
| `Password` | `password`, `pass`           |
| `Host`     | `host`                       |
| `Port`     | `port`                       |
| `Database` | `database`, `dbname`, `db`   |

## Connection

`Parse`, `FromUrl`, and `FromPair` all return a pointer to a `connection` struct. The type itself is unexported, but its
fields and methods are not — so you can read the fields and call the methods normally. Use `:=` to let Go infer the type.

### Fields

| Field         | Type                  | Notes                                                                      |
|---------------|-----------------------|----------------------------------------------------------------------------|
| `Type`        | `*string`             | Scheme name. `nil` if the input has no scheme.                             |
| `Username`    | `*string`             | `nil` if the input has no user info. Empty pointer if the user is `""`.    |
| `Password`    | `*string`             | `nil` if the input has no password. Empty pointer if the password is `""`. |
| `Host`        | `string`              | Hostname only — no port, no user info.                                     |
| `Port`        | `string`              | Port as written in the input.                                              |
| `NumericPort` | `int`                 | Same as `Port`, parsed to `int`. Stays `0` if `Port` is not a number.      |
| `Database`    | `string`              | Database name. For URLs, this is the path with the leading `/` removed.    |
| `Properties`  | `map[string][]string` | Extra key/value pairs. A single key can hold many values, in input order.  |

`Username` and `Password` are pointers on purpose. There are three states to tell apart:

- `nil`      — the input did not contain a username (or password) at all.
- `&""`      — the input had a username (or password), but it was empty.
- `&"alice"` — the input had a real value.

### Methods

#### `IsFor(t string, sensitive ...bool) bool`

Returns `true` if `Type` matches `t`. The match is case-insensitive by default. Pass `true` as the second argument for an
exact, case-sensitive match.

```go
conn, _ := parser.Parse("postgres://example.com")
conn.IsFor("postgres")       // true
conn.IsFor("POSTGRES")       // true  — case-insensitive by default
conn.IsFor("POSTGRES", true) // false — case-sensitive
```

If `Type` is `nil`, `IsFor` always returns `false`.

#### `Address() string`

Returns `Host:Port` if `Port` is set, otherwise just `Host`.

```go
conn.Address() // "example.com:5432"
```

#### `HasUsername() bool` and `HasPassword() bool`

Return `true` if `Username` (or `Password`) is set. The empty string still counts as set — these methods only check the
pointer, not the value.

#### `HasProperty(key ...string) bool`

- Without arguments, returns `true` if the connection has at least one property.
- With one argument, returns `true` if that key exists.

```go
conn.HasProperty()          // true if any property is set
conn.HasProperty("sslmode") // true if "sslmode" is in Properties
```

#### `GetProperty(key string, defaults ...string) string`

Returns the **first** value stored for `key`. If the key is missing, or its slice is empty, the optional default is
returned. With no default, an empty string is returned.

```go
conn.GetProperty("sslmode")          // "prefer"
conn.GetProperty("schema", "public") // "public" — key missing, default returned
conn.GetProperty("missing")          // ""
```

#### `GetProperties(key string) []string`

Returns every value stored for `key`, in the order they appeared in the input. Returns `nil` if the key is missing.

## Repeated query parameters

Some real-world connection strings allow the same query parameter to appear more than once, and order can matter. The
classic example is MongoDB's `readPreferenceTags`, where each entry is a fallback tag set:

```
mongodb://host/?readPreference=secondary&readPreferenceTags=dc:east,rack:1&readPreferenceTags=dc:east&readPreferenceTags=
```

`Properties` keeps every value, in the order it was seen.

```go
conn, _ := parser.Parse(
    "mongodb://host/?readPreferenceTags=dc:east,rack:1&readPreferenceTags=dc:east&readPreferenceTags=",
)

conn.GetProperty("readPreferenceTags")   // "dc:east,rack:1" — first value
conn.GetProperties("readPreferenceTags") // ["dc:east,rack:1", "dc:east", ""]
```

The same is true for the delimited form. `host=example.com tag=a tag=b tag=c` produces:

```go
conn.GetProperties("tag") // []string{"a", "b", "c"}
```

# Issues?

If you find a bug, a missing feature, or unclear documentation, please open an issue. Pull requests are welcome.
