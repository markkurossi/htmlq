//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

type IQLTest struct {
	q    string
	v    [][]string
	rest [][][]string
}

var builtInData = `Year,IVal,FVal
1970,100,100.5
1971,200,200.5
1972,300,300.5
1973,400,400.5
1974,500,500.5`

var builtInTests = []IQLTest{
	{
		q: `
select AVG(Year)
from (
      select Year, IVal, FVal from data
     );`,
		v: [][]string{{"1972"}},
	},
	{
		q: `
select COUNT(Year)
from (
      select Year, IVal, FVal from data
     );`,
		v: [][]string{{"5"}},
	},
	{
		q: `
select MAX(Year)
from (
      select Year, IVal, FVal from data
     );`,
		v: [][]string{{"1974"}},
	},
	{
		q: `
select MIN(Year)
from (
      select Year, IVal, FVal from data
     );`,
		v: [][]string{{"1970"}},
	},
	{
		q: `
select SUM(Year)
from (
      select Year, IVal, FVal from data
     );`,
		v: [][]string{{"9860"}},
	},
	{
		q: `
SELECT NULLIF(4, 4);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `
SELECT NULLIF(5, 4);`,
		v: [][]string{{"5"}},
	},
	{
		q: `
SELECT 5 / NULLIF(0.0, 0.0);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `
SELECT 5 / NULLIF(5.0, 0.0);`,
		v: [][]string{{"1"}},
	},

	// CAST tests.
	{
		q: `SELECT CAST(false AS BOOLEAN);`,
		v: [][]string{{"false"}},
	},
	{
		q: `SELECT CAST(false AS VARCHAR);`,
		v: [][]string{{"false"}},
	},
	{
		q: `SELECT CAST(5 AS INTEGER);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5 AS REAL);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5 AS VARCHAR);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5.0 AS INTEGER);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5.0 AS REAL);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5.0 AS VARCHAR);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST('5' AS INTEGER);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST('5' AS REAL);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST('5' AS VARCHAR);`,
		v: [][]string{{"5"}},
	},

	// String functions.
	{
		q: `SELECT CHAR(-1);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT CHAR(0xffffffff);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT CHAR(42);`,
		v: [][]string{{"*"}},
	},
	{
		q: `SELECT CHARINDEX('Reflectors are vital safety' +
                             ' components of your bicycle.',
                             'bicycle');`,
		v: [][]string{{"48"}},
	},
	{
		q: `SELECT CHARINDEX('Reflectors are vital safety' +
                             ' components of your bicycle.',
                             'vital', 5);`,
		v: [][]string{{"16"}},
	},
	{
		q: `SELECT CHARINDEX('Reflectors are vital safety' +
                             ' components of your bicycle.',
                             'bike');`,
		v: [][]string{{"0"}},
	},
	{
		q: `SELECT CONCAT('Happy ', 'Birthday ', 11, '/', '25');`,
		v: [][]string{{"Happy Birthday 11/25"}},
	},
	{
		q: `SELECT CONCAT('Name', NULL, 'Lastname');`,
		v: [][]string{{"NameLastname"}},
	},
	{
		q: `SELECT CONCAT_WS(',', '1 Microsoft Way', NULL, NULL, 'Redmond',
                             'WA', 98052);`,
		v: [][]string{{"1 Microsoft Way,Redmond,WA,98052"}},
	},
	{
		q: `SELECT CONCAT_WS(null, 'a', 'b', 'c');`,
		v: [][]string{{"abc"}},
	},
	{
		q: `SELECT CONCAT_WS('-', null, 'a', null);`,
		v: [][]string{{"a"}},
	},
	{
		q: `SELECT BASE64ENC('foo');`,
		v: [][]string{{"Zm9v"}},
	},
	{
		q: `SELECT BASE64DEC('Zm9v');`,
		v: [][]string{{"foo"}},
	},
	{
		q: `SELECT LASTCHARINDEX('}abcd}def', '}');`,
		v: [][]string{{"6"}},
	},
	{
		q: `SELECT LASTCHARINDEX('}abcd}def', ',');`,
		v: [][]string{{"0"}},
	},
	{
		q: `SELECT LEFT('Hello, world!', 6);`,
		v: [][]string{{"Hello,"}},
	},
	{
		q: `SELECT LEFT('Hello', 6);`,
		v: [][]string{{"Hello"}},
	},
	{
		q: `SELECT LEN('Hello, world!');`,
		v: [][]string{{"13"}},
	},
	{
		q: `SELECT LOWER('Hello, world!');`,
		v: [][]string{{"hello, world!"}},
	},
	{
		q: `SELECT LPAD('ABC', 5, '*');`,
		v: [][]string{{"**ABC"}},
	},
	{
		q: `SELECT LPAD('ABC', 5);`,
		v: [][]string{{"  ABC"}},
	},
	{
		q: `SELECT LPAD('ABCDEF', 5, '*');`,
		v: [][]string{{"ABCDE"}},
	},
	{
		q: `SELECT LTRIM('  Hello, World!  ');`,
		v: [][]string{{"Hello, World!  "}},
	},
	{
		q: `SELECT NCHAR(64);`,
		v: [][]string{{"@"}},
	},
	{
		q: `SELECT REPLICATE('0', 4);`,
		v: [][]string{{"0000"}},
	},
	{
		q: `SELECT REPLICATE('0', -1);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT REVERSE('Ken');`,
		v: [][]string{{"neK"}},
	},
	{
		q: `SELECT REVERSE('Rob');`,
		v: [][]string{{"boR"}},
	},
	{
		q: `SELECT REVERSE(1234);`,
		v: [][]string{{"4321"}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 0);`,
		v: [][]string{{""}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 2);`,
		v: [][]string{{"fg"}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 7);`,
		v: [][]string{{"abcdefg"}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 100000);`,
		v: [][]string{{"abcdefg"}},
	},
	{
		q: `SELECT RTRIM('  Hello, World!  ');`,
		v: [][]string{{"  Hello, World!"}},
	},
	{
		q: `SELECT SPACE(5);`,
		v: [][]string{{"     "}},
	},
	{
		q: `SELECT SPACE(-1);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT STUFF('abcdef', 2, 3, 'ijklmn');`,
		v: [][]string{{"aijklmnef"}},
	},
	{
		q: `SELECT STUFF('abcdef', 0, 3, 'ijklmn');`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT STUFF('abcdef', -1, 3, 'ijklmn');`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT STUFF('abcdef', 7, 0, 'ijklmn');`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT STUFF('abcdef', 2, -1, 'ijklmn');`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `SELECT STUFF('abcdef', 2, 100, 'ijklmn');`,
		v: [][]string{{"aijklmn"}},
	},
	{
		q: `SELECT STUFF('abcdef', 2, 0, 'ijklmn');`,
		v: [][]string{{"aijklmnbcdef"}},
	},
	{
		q: `SELECT STUFF('abcdef', 2, 4, null);`,
		v: [][]string{{"af"}},
	},
	{
		q: `SELECT SUBSTRING('master', 1, 1);`,
		v: [][]string{{"m"}},
	},
	{
		q: `SELECT SUBSTRING('master', 3, 2);`,
		v: [][]string{{"st"}},
	},
	{
		q: `SELECT SUBSTRING('tempdb', 1, 1);`,
		v: [][]string{{"t"}},
	},
	{
		q: `SELECT SUBSTRING('tempdb', 3, 2);`,
		v: [][]string{{"mp"}},
	},
	{
		q: `SELECT SUBSTRING('hello', 0, 2);`,
		v: [][]string{{"he"}},
	},
	{
		q: `SELECT SUBSTRING('hello', -10, 2);`,
		v: [][]string{{"he"}},
	},
	{
		q: `SELECT SUBSTRING('hello', 100, 2);`,
		v: [][]string{{""}},
	},
	{
		q: `SELECT SUBSTRING('hello', 3, 100);`,
		v: [][]string{{"llo"}},
	},
	{
		q: `SELECT TRIM('  Hello, World!  ');`,
		v: [][]string{{"Hello, World!"}},
	},
	{
		q: `DECLARE nstring VARCHAR;
SET nstring = 'Åkergatan 24';
SELECT UNICODE(nstring), NCHAR(UNICODE(nstring));`,
		v: [][]string{{"197", "Å"}},
	},
	{
		q: `SELECT UPPER('Hello, world!');`,
		v: [][]string{{"HELLO, WORLD!"}},
	},

	// Datetime literals.
	{
		q: `SELECT YEAR('2010-04-30T01:01:01.1234567-07:00');`,
		v: [][]string{{"2010"}},
	},
	{
		q: `SELECT YEAR('2007-04-30 13:10:02.0474381');`,
		v: [][]string{{"2007"}},
	},
	{
		q: `SELECT YEAR('2007-04-30 13:10:02.0474381 -07:00');`,
		v: [][]string{{"2007"}},
	},
	{
		q: `SELECT YEAR('2007-04-30');`,
		v: [][]string{{"2007"}},
	},
	{
		q: `SELECT YEAR(0);`,
		// XXX SQL Server return 1900
		v: [][]string{{"1970"}},
	},
	{
		q: `SELECT MONTH('2007-04-30T01:01:01.1234567 -07:00');`,
		v: [][]string{{"4"}},
	},
	{
		q: `SELECT DAY('2015-04-30 01:01:01.1234567');`,
		v: [][]string{{"30"}},
	},

	// Datetime functions.
	{
		q: `SELECT DATEDIFF(year,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(day,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(hour,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(minute,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(second,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(millisecond,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(microsecond,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(nanosecond,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"100"}},
	},
	{
		q: `DECLARE now DATETIME;
SET now = GETDATE();
SELECT DATEDIFF(year, now, now);`,
		v: [][]string{{"0"}},
	},
	{
		q: `SELECT YEAR('2005-12-31 23:59:59.9999999');`,
		v: [][]string{{"2005"}},
	},
}

func TestBuiltIn(t *testing.T) {
	data := fmt.Sprintf("data:text/csv;base64,%s",
		base64.StdEncoding.EncodeToString([]byte(builtInData)))

	for testID, input := range builtInTests {
		name := fmt.Sprintf("Test %d", testID)
		parser := NewParser(bytes.NewReader([]byte(input.q)), name)

		parser.SetString("data", data)

		for {
			q, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("%s: parse failed: %v", name, err)
			}
			verifyResult(t, name, input.q, q, input.v)
		}
	}
}

func verifyResult(t *testing.T, name, source string, q types.Source,
	v [][]string) {
	rows, err := q.Get()
	if err != nil {
		t.Errorf("%s: q.Get failed: %v:\n%s\n", name, err, source)
		return
	}
	if len(rows) != len(v) {
		t.Errorf("%s: got %d rows, expected %d", name, len(rows), len(v))
		printResult(q, rows)
		return
	}
	for rowID, row := range rows {
		if len(row) != len(v[rowID]) {
			t.Fatalf("%s: row %d: got %d columns, expected %d",
				name, rowID, len(row), len(v[rowID]))
			printResult(q, rows)
			continue
		}
		for colID, col := range row {
			result := col.String()
			if result != v[rowID][colID] {
				t.Errorf("%s: %d.%d: got '%s', expected '%s'",
					name, rowID, colID, result, v[rowID][colID])
				printResult(q, rows)
			}
		}
	}
}

func printResult(q types.Source, rows []types.Row) {
	tab, err := types.Tabulate(q, tabulate.Unicode)
	if err != nil {
		return
	}
	tab.Print(os.Stdout)
}
