//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"

	"github.com/markkurossi/iql/types"
)

// Function implements a function.
type Function struct {
	Name       string
	Impl       FunctionImpl
	MinArgs    int
	MaxArgs    int
	Idempotent bool
}

// FunctionImpl implements the built-in IQL functions.
type FunctionImpl func(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error)

var builtIns = []Function{
	// Aggregate functions.
	{
		Name:       "AVG",
		Impl:       builtInAvg,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	{
		Name:       "COUNT",
		Impl:       builtInCount,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	{
		Name:       "MAX",
		Impl:       builtInMax,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	{
		Name:       "MIN",
		Impl:       builtInMin,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	{
		Name:       "SUM",
		Impl:       builtInSum,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},

	{
		Name:       "NULLIF",
		Impl:       builtInNullIf,
		MinArgs:    2,
		MaxArgs:    2,
		Idempotent: false,
	},

	// String functions.
	{
		Name:       "LEFT",
		Impl:       builtInLeft,
		MinArgs:    2,
		MaxArgs:    2,
		Idempotent: false,
	},
}

func builtInAvg(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	seen := make(map[types.Type]bool)

	var intSum int64
	var floatSum float64
	var count int

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, columns, nil)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case types.NullValue:

		case types.IntValue:
			add, err := v.Int()
			if err != nil {
				return nil, err

			}
			seen[types.Int] = true
			intSum += add
			count++

		case types.FloatValue:
			add, err := v.Float()
			if err != nil {
				return nil, err
			}
			seen[types.Float] = true
			floatSum += add
			count++

		default:
			return nil, fmt.Errorf("AVG over %T", val)
		}
	}
	if count == 0 || len(seen) != 1 {
		return types.Null, nil
	}
	if seen[types.Float] {
		return types.FloatValue(floatSum / float64(count)), nil
	}
	return types.IntValue(intSum / int64(count)), nil
}

func builtInCount(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	var count int
	for _, countRow := range rows {
		val, err := args[0].Eval(countRow, columns, nil)
		if err != nil {
			return nil, err
		}
		_, ok := val.(types.NullValue)
		if !ok {
			count++
		}
	}
	return types.IntValue(count), nil
}

func builtInMax(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	seen := make(map[types.Type]bool)

	var intMax int64
	var floatMax float64

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, columns, nil)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case types.NullValue:

		case types.IntValue:
			ival, err := v.Int()
			if err != nil {
				return nil, err
			}
			if !seen[types.Int] || ival > intMax {
				intMax = ival
			}
			seen[types.Int] = true

		case types.FloatValue:
			fval, err := v.Float()
			if err != nil {
				return nil, err
			}
			if !seen[types.Float] || fval > floatMax {
				floatMax = fval
			}
			seen[types.Float] = true

		default:
			return nil, fmt.Errorf("MAX over %T", val)
		}
	}
	if seen[types.Float] && seen[types.Int] {
		var r float64
		if float64(intMax) > floatMax {
			r = float64(intMax)
		} else {
			r = floatMax
		}
		return types.FloatValue(r), nil
	} else if seen[types.Float] {
		return types.FloatValue(floatMax), nil
	}
	return types.IntValue(intMax), nil

}

func builtInMin(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	seen := make(map[types.Type]bool)

	var intMin int64
	var floatMin float64

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, columns, nil)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case types.NullValue:

		case types.IntValue:
			ival, err := v.Int()
			if err != nil {
				return nil, err
			}
			if !seen[types.Int] || ival < intMin {
				intMin = ival
			}
			seen[types.Int] = true

		case types.FloatValue:
			fval, err := v.Float()
			if err != nil {
				return nil, err
			}
			if !seen[types.Float] || fval < floatMin {
				floatMin = fval
			}
			seen[types.Float] = true

		default:
			return nil, fmt.Errorf("MIN over %T", val)
		}
	}
	if seen[types.Float] && seen[types.Int] {
		var r float64
		if float64(intMin) < floatMin {
			r = float64(intMin)
		} else {
			r = floatMin
		}
		return types.FloatValue(r), nil
	} else if seen[types.Float] {
		return types.FloatValue(floatMin), nil
	}

	return types.IntValue(intMin), nil
}

func builtInSum(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	seen := make(map[types.Type]bool)

	var intSum int64
	var floatSum float64

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, columns, nil)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case types.NullValue:

		case types.IntValue:
			add, err := v.Int()
			if err != nil {
				return nil, err
			}
			seen[types.Int] = true
			intSum += add

		case types.FloatValue:
			add, err := v.Float()
			if err != nil {
				return nil, err
			}
			seen[types.Float] = true
			floatSum += add

		default:
			return nil, fmt.Errorf("SUM over %T", val)
		}
	}
	if seen[types.Float] && seen[types.Int] {
		return types.FloatValue(floatSum + float64(intSum)), nil
	} else if seen[types.Float] {
		return types.FloatValue(floatSum), nil
	}
	return types.IntValue(intSum), nil
}

func builtInNullIf(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	val, err := args[0].Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	cmp, err := args[1].Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	ok, err := types.Equal(val, cmp)
	if err != nil {
		return nil, err
	}
	if ok {
		return types.Null, nil
	}
	return val, nil
}

func builtInLeft(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	strVal, err := args[0].Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	idxVal, err := args[1].Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()
	idx64, err := idxVal.Int()
	if err != nil {
		return nil, err
	}
	idx := int(idx64)
	if idx > len(str) {
		idx = len(str)
	}
	return types.StringValue(str[:idx]), nil
}

var builtInsByName map[string]*Function

func init() {
	builtInsByName = make(map[string]*Function)
	for idx, bi := range builtIns {
		builtInsByName[bi.Name] = &builtIns[idx]
	}
}

func builtIn(name string) *Function {
	return builtInsByName[name]
}

func (f *Function) String() string {
	return f.Name
}
