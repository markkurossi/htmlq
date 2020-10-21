//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

func TestCSVCorrect(t *testing.T) {
	name := "test.csv"
	source, err := New(name, "", []types.ColumnSelector{
		{
			Name: types.Reference{
				Column: "0",
			},
			As: "Share",
		},
		{
			Name: types.Reference{
				Column: "1",
			},
			As: "Count",
		},
	})
	if err != nil {
		t.Fatalf("NewCSV failed: %s", err)
	}
	rows, err := source.Get()
	if err != nil {
		t.Fatalf("csv.Get() failed: %s", err)
	}
	if len(rows) != 3 {
		t.Errorf("%s: unexpected number of rows", name)
	}
	if len(rows[0]) != 2 {
		t.Errorf("%s: unexpected number of columns", name)
	}
	tab := types.Tabulate(source, tabulate.Unicode)
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col.String())
		}
	}
	tab.Print(os.Stdout)
}

func TestCSVOptions(t *testing.T) {
	source, err := New("test_options.csv", "skip=1 comma=;  comment=# ",
		[]types.ColumnSelector{
			{
				Name: types.Reference{
					Column: "0",
				},
				As: "Year",
			},
			{
				Name: types.Reference{
					Column: "1",
				},
				As: "Value",
			},
			{
				Name: types.Reference{
					Column: "2",
				},
				As: "Delta",
			},
		})
	if err != nil {
		t.Fatalf("NewCSV failed: %s", err)
	}
	rows, err := source.Get()
	if err != nil {
		t.Fatalf("csv.Get() failed: %s", err)
	}
	tab := types.Tabulate(source, tabulate.Unicode)
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col.String())
		}
	}
	tab.Print(os.Stdout)
}
