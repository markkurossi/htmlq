//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/markkurossi/tabulate"
)

var (
	_ Source = &CSV{}
	_ Source = &HTML{}
	_ Column = StringColumn("")
	_ Column = StringsColumn([]string{})
	_ Column = NullColumn{}
)

// NewSource defines a constructor for data sources.
type NewSource func(in io.ReadCloser, filter string, columns []ColumnSelector) (
	Source, error)

// New creates a new data source for the URL.
func New(url, filter string, columns []ColumnSelector) (Source, error) {
	input, format, err := openInput(url)
	if err != nil {
		return nil, err
	}

	n, ok := formats[format]
	if !ok {
		return nil, fmt.Errorf("unknown data format '%s'", format)
	}
	return n(input, filter, columns)
}

func openInput(input string) (io.ReadCloser, Format, error) {
	var resolver Resolver

	u, err := url.Parse(input)
	if err != nil {
		resolver.ResolvePath(input)
	} else {
		resolver.ResolvePath(u.Path)
	}
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		resp, err := http.Get(input)
		if err != nil {
			return nil, 0, err
		}
		if resp.StatusCode != http.StatusOK {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			return nil, 0, fmt.Errorf("HTTP URL '%s' not found", input)
		}

		resolver.ResolveMediaType(resp.Header.Get("Content-Type"))

		format, err := resolver.Format()
		return resp.Body, format, err
	}
	if err == nil && u.Scheme == "data" {
		idx := strings.IndexByte(input, ',')
		if idx < 0 {
			return nil, 0, fmt.Errorf("malformed data URI: %s", input)
		}
		data := input[idx+1:]
		contentType := input[5:idx]
		var encoding string

		idx = strings.IndexByte(contentType, ';')
		if idx >= 0 {
			encoding = contentType[idx+1:]
			contentType = contentType[:idx]
		}

		var decoded []byte

		// Decode data.
		switch encoding {
		case "base64":
			decoded, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				return nil, 0, err
			}
		case "":
			decoded = []byte(data)
		default:
			return nil, 0, fmt.Errorf("unknown data URI encoding: %s", encoding)
		}

		// Resolve format.
		resolver.ResolveMediaType(contentType)

		format, err := resolver.Format()

		return &memory{
			in: bytes.NewReader(decoded),
		}, format, err
	}

	f, err := os.Open(input)
	if err != nil {
		return nil, 0, err
	}

	format, err := resolver.Format()

	return f, format, err
}

type memory struct {
	in io.Reader
}

func (m *memory) Read(p []byte) (n int, err error) {
	return m.in.Read(p)
}

func (m *memory) Close() error {
	return nil
}

// ColumnType specifies column types.
type ColumnType int

// Column types.
const (
	ColumnBool ColumnType = iota
	ColumnInt
	ColumnFloat
	ColumnString
)

// Literal values.
const (
	True  = "true"
	False = "false"
)

var columnTypes = map[ColumnType]string{
	ColumnBool:   "bool",
	ColumnInt:    "int",
	ColumnFloat:  "float",
	ColumnString: "string",
}

func (t ColumnType) String() string {
	name, ok := columnTypes[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{columnType %d}", t)
}

// Align returns the type specific column alignment type.
func (t ColumnType) Align() tabulate.Align {
	if t == ColumnString {
		return tabulate.ML
	}
	return tabulate.MR
}

// ColumnSelector implements data column selector.
type ColumnSelector struct {
	Name Reference
	As   string
	Type ColumnType
}

// IsPublic reports if the column is public and should be included in
// the result set.
func (col ColumnSelector) IsPublic() bool {
	runes := []rune(col.String())
	return len(runes) > 0 && unicode.IsUpper(runes[0])
}

// ResolveType resolves the column type based on the argument column
// value. This function must be called once for each value and it will
// resolve the most specific column type that is able to represent all
// values
func (col *ColumnSelector) ResolveType(val string) {
	// Skip empty values.
	if len(val) == 0 {
		return
	}
	for {
		switch col.Type {
		case ColumnBool:
			if val == True || val == False {
				return
			}
			col.Type = ColumnInt

		case ColumnInt:
			_, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				return
			}
			col.Type = ColumnFloat

		case ColumnFloat:
			_, err := strconv.ParseFloat(val, 64)
			if err == nil {
				return
			}
			col.Type = ColumnString

		case ColumnString:
			return
		}
	}
}

func (col ColumnSelector) String() string {
	if len(col.As) > 0 {
		return col.As
	}
	return col.Name.Column
}

// Reference implements a reference to table column.
type Reference struct {
	Source string
	Column string
}

// NewReference creates a new column reference for the argument name.
func NewReference(name string) (Reference, error) {
	// XXX escapes
	parts := strings.Split(name, ".")
	switch len(parts) {
	case 1:
		return Reference{
			Column: parts[0],
		}, nil

	case 2:
		return Reference{
			Source: parts[0],
			Column: parts[1],
		}, nil

	default:
		return Reference{}, fmt.Errorf("invalid column reference '%s'", name)
	}
}

// IsAbsolute tests if the reference is an absolute reference
// i.e. specifying both the data source and column.
func (ref *Reference) IsAbsolute() bool {
	return len(ref.Source) > 0
}

func (ref Reference) String() string {
	// XXX escapes
	if len(ref.Source) > 0 {
		return fmt.Sprintf("%s.%s", ref.Source, ref.Column)
	}
	return ref.Column
}

// Column defines a data column.
type Column interface {
	// Count returns number of column elements.
	Count() int
	// Size returns the column size in characters.
	Size() int
	Bool() (Value, error)
	Int() (Value, error)
	Float() (Value, error)
	String() string
}

// NullColumn implements a null-column.
type NullColumn NullValue

// Count implements the Column.Count().
func (n NullColumn) Count() int {
	return 0
}

// Size implements the Column.Size().
func (n NullColumn) Size() int {
	return 0
}

// Bool implements the Column.Bool().
func (n NullColumn) Bool() (Value, error) {
	return Null, nil
}

// Int implements the Column.Int().
func (n NullColumn) Int() (Value, error) {
	return Null, nil
}

// Float implements the Column.Float().
func (n NullColumn) Float() (Value, error) {
	return Null, nil
}

func (n NullColumn) String() string {
	return "NULL"
}

// StringColumn implements a string column.
type StringColumn string

// Count implements the Column.Count().
func (s StringColumn) Count() int {
	return 1
}

// Size implements the Column.Size().
func (s StringColumn) Size() int {
	return len(s)
}

// Bool implements the Column.Bool().
func (s StringColumn) Bool() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	switch s {
	case True:
		return BoolValue(true), nil
	case False:
		return BoolValue(false), nil
	default:
		return nil, fmt.Errorf("string value '%s' used as bool", s)
	}
}

// Int implements the Column.Int().
func (s StringColumn) Int() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	v, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return nil, err
	}
	return IntValue(v), nil
}

// Float implements the Column.Float().
func (s StringColumn) Float() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	v, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return nil, err
	}
	return FloatValue(v), nil
}

func (s StringColumn) String() string {
	return string(s)
}

// StringsColumn implements a string array column.
type StringsColumn []string

// Count implements the Column.Count().
func (s StringsColumn) Count() int {
	return len(s)
}

// Size implements the Column.Size().
func (s StringsColumn) Size() int {
	var size int
	for _, e := range s {
		if len(e) > size {
			size = len(e)
		}
	}
	return size
}

// Bool implements the Column.Bool().
func (s StringsColumn) Bool() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	return nil, fmt.Errorf("string array used as bool")
}

// Int implements the Column.Int().
func (s StringsColumn) Int() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	return nil, fmt.Errorf("string array used as int")
}

// Float implements the Column.Float().
func (s StringsColumn) Float() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	return nil, fmt.Errorf("string array used as float")
}

func (s StringsColumn) String() string {
	return fmt.Sprintf("%v", []string(s))
}

// Row defines an input data row.
type Row []Column

// Source is an interface that defines data input sources.
type Source interface {
	Columns() []ColumnSelector
	Get() ([]Row, error)
}

// Table creates a tabulation table for the data source.
func Table(source Source, style tabulate.Style) *tabulate.Tabulate {
	tab := tabulate.New(style)
	for _, col := range source.Columns() {
		if col.IsPublic() {
			tab.Header(col.String()).SetAlign(col.Type.Align())
		}
	}
	return tab
}
