/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package evalengine

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/stackql/stackql-parser/go/test/utils"

	"github.com/stretchr/testify/require"

	"github.com/stackql/stackql-parser/go/sqltypes"

	querypb "github.com/stackql/stackql-parser/go/vt/proto/query"
	vtrpcpb "github.com/stackql/stackql-parser/go/vt/proto/vtrpc"
	"github.com/stackql/stackql-parser/go/vt/vterrors"
)

func TestArithmetics(t *testing.T) {
	type tcase struct {
		v1, v2, out sqltypes.Value
		err         string
	}

	tests := []struct {
		operator string
		f        func(a, b sqltypes.Value) (sqltypes.Value, error)
		cases    []tcase
	}{{
		operator: "-",
		f:        Subtract,
		cases: []tcase{{
			// All Nulls
			v1:  sqltypes.NULL,
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// First value null.
			v1:  sqltypes.NewInt32(1),
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// Second value null.
			v1:  sqltypes.NULL,
			v2:  sqltypes.NewInt32(1),
			out: sqltypes.NULL,
		}, {
			// case with negative value
			v1:  sqltypes.NewInt64(-1),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewInt64(1),
		}, {
			// testing for int64 overflow with min negative value
			v1:  sqltypes.NewInt64(math.MinInt64),
			v2:  sqltypes.NewInt64(1),
			err: "BIGINT value is out of range in -9223372036854775808 - 1",
		}, {
			v1:  sqltypes.NewUint64(4),
			v2:  sqltypes.NewInt64(5),
			err: "BIGINT UNSIGNED value is out of range in 4 - 5",
		}, {
			// testing uint - int
			v1:  sqltypes.NewUint64(7),
			v2:  sqltypes.NewInt64(5),
			out: sqltypes.NewUint64(2),
		}, {
			v1:  sqltypes.NewUint64(math.MaxUint64),
			v2:  sqltypes.NewInt64(0),
			out: sqltypes.NewUint64(math.MaxUint64),
		}, {
			// testing for int64 overflow
			v1:  sqltypes.NewInt64(math.MinInt64),
			v2:  sqltypes.NewUint64(0),
			err: "BIGINT UNSIGNED value is out of range in -9223372036854775808 - 0",
		}, {
			v1:  sqltypes.TestValue(querypb.Type_VARCHAR, "c"),
			v2:  sqltypes.NewInt64(1),
			out: sqltypes.NewInt64(-1),
		}, {
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.TestValue(querypb.Type_VARCHAR, "c"),
			out: sqltypes.NewUint64(1),
		}, {
			// testing for error for parsing float value to uint64
			v1:  sqltypes.TestValue(querypb.Type_UINT64, "1.2"),
			v2:  sqltypes.NewInt64(2),
			err: "strconv.ParseUint: parsing \"1.2\": invalid syntax",
		}, {
			// testing for error for parsing float value to uint64
			v1:  sqltypes.NewUint64(2),
			v2:  sqltypes.TestValue(querypb.Type_UINT64, "1.2"),
			err: "strconv.ParseUint: parsing \"1.2\": invalid syntax",
		}, {
			// uint64 - uint64
			v1:  sqltypes.NewUint64(8),
			v2:  sqltypes.NewUint64(4),
			out: sqltypes.NewUint64(4),
		}, {
			// testing for float subtraction: float - int
			v1:  sqltypes.NewFloat64(1.2),
			v2:  sqltypes.NewInt64(2),
			out: sqltypes.NewFloat64(-0.8),
		}, {
			// testing for float subtraction: float - uint
			v1:  sqltypes.NewFloat64(1.2),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewFloat64(-0.8),
		}, {
			v1:  sqltypes.NewInt64(-1),
			v2:  sqltypes.NewUint64(2),
			err: "BIGINT UNSIGNED value is out of range in -1 - 2",
		}, {
			v1:  sqltypes.NewInt64(2),
			v2:  sqltypes.NewUint64(1),
			out: sqltypes.NewUint64(1),
		}, {
			// testing int64 - float64 method
			v1:  sqltypes.NewInt64(-2),
			v2:  sqltypes.NewFloat64(1.0),
			out: sqltypes.NewFloat64(-3.0),
		}, {
			// testing uint64 - float64 method
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.NewFloat64(-2.0),
			out: sqltypes.NewFloat64(3.0),
		}, {
			// testing uint - int to return uintplusint
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewUint64(3),
		}, {
			// testing for float - float
			v1:  sqltypes.NewFloat64(1.2),
			v2:  sqltypes.NewFloat64(3.2),
			out: sqltypes.NewFloat64(-2),
		}, {
			// testing uint - uint if v2 > v1
			v1:  sqltypes.NewUint64(2),
			v2:  sqltypes.NewUint64(4),
			err: "BIGINT UNSIGNED value is out of range in 2 - 4",
		}, {
			// testing uint - (- int)
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewUint64(3),
		}},
	}, {
		operator: "+",
		f:        Add,
		cases: []tcase{{
			// All Nulls
			v1:  sqltypes.NULL,
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// First value null.
			v1:  sqltypes.NewInt32(1),
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// Second value null.
			v1:  sqltypes.NULL,
			v2:  sqltypes.NewInt32(1),
			out: sqltypes.NULL,
		}, {
			// case with negatives
			v1:  sqltypes.NewInt64(-1),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewInt64(-3),
		}, {
			// testing for overflow int64, result will be unsigned int
			v1:  sqltypes.NewInt64(math.MaxInt64),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewUint64(9223372036854775809),
		}, {
			v1:  sqltypes.NewInt64(-2),
			v2:  sqltypes.NewUint64(1),
			err: "BIGINT UNSIGNED value is out of range in 1 + -2",
		}, {
			v1:  sqltypes.NewInt64(math.MaxInt64),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewInt64(9223372036854775805),
		}, {
			// Normal case
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewUint64(3),
		}, {
			// testing for overflow uint64
			v1:  sqltypes.NewUint64(math.MaxUint64),
			v2:  sqltypes.NewUint64(2),
			err: "BIGINT UNSIGNED value is out of range in 18446744073709551615 + 2",
		}, {
			// int64 underflow
			v1:  sqltypes.NewInt64(math.MinInt64),
			v2:  sqltypes.NewInt64(-2),
			err: "BIGINT value is out of range in -9223372036854775808 + -2",
		}, {
			// checking int64 max value can be returned
			v1:  sqltypes.NewInt64(math.MaxInt64),
			v2:  sqltypes.NewUint64(0),
			out: sqltypes.NewUint64(9223372036854775807),
		}, {
			// testing whether uint64 max value can be returned
			v1:  sqltypes.NewUint64(math.MaxUint64),
			v2:  sqltypes.NewInt64(0),
			out: sqltypes.NewUint64(math.MaxUint64),
		}, {
			v1:  sqltypes.NewUint64(math.MaxInt64),
			v2:  sqltypes.NewInt64(1),
			out: sqltypes.NewUint64(9223372036854775808),
		}, {
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.TestValue(querypb.Type_VARCHAR, "c"),
			out: sqltypes.NewUint64(1),
		}, {
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.TestValue(querypb.Type_VARCHAR, "1.2"),
			out: sqltypes.NewFloat64(2.2),
		}, {
			v1:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
			v2:  sqltypes.NewInt64(2),
			err: "strconv.ParseInt: parsing \"1.2\": invalid syntax",
		}, {
			v1:  sqltypes.NewInt64(2),
			v2:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
			err: "strconv.ParseInt: parsing \"1.2\": invalid syntax",
		}, {
			// testing for uint64 overflow with max uint64 + int value
			v1:  sqltypes.NewUint64(math.MaxUint64),
			v2:  sqltypes.NewInt64(2),
			err: "BIGINT UNSIGNED value is out of range in 18446744073709551615 + 2",
		}},
	}, {
		operator: "/",
		f:        Divide,
		cases: []tcase{{
			//All Nulls
			v1:  sqltypes.NULL,
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// First value null.
			v1:  sqltypes.NULL,
			v2:  sqltypes.NewInt32(1),
			out: sqltypes.NULL,
		}, {
			// Second value null.
			v1:  sqltypes.NewInt32(1),
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// Second arg 0
			v1:  sqltypes.NewInt32(5),
			v2:  sqltypes.NewInt32(0),
			out: sqltypes.NULL,
		}, {
			// Both arguments zero
			v1:  sqltypes.NewInt32(0),
			v2:  sqltypes.NewInt32(0),
			out: sqltypes.NULL,
		}, {
			// case with negative value
			v1:  sqltypes.NewInt64(-1),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewFloat64(0.5000),
		}, {
			// float64 division by zero
			v1:  sqltypes.NewFloat64(2),
			v2:  sqltypes.NewFloat64(0),
			out: sqltypes.NULL,
		}, {
			// Lower bound for int64
			v1:  sqltypes.NewInt64(math.MinInt64),
			v2:  sqltypes.NewInt64(1),
			out: sqltypes.NewFloat64(math.MinInt64),
		}, {
			// upper bound for uint64
			v1:  sqltypes.NewUint64(math.MaxUint64),
			v2:  sqltypes.NewUint64(1),
			out: sqltypes.NewFloat64(math.MaxUint64),
		}, {
			// testing for error in types
			v1:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
			v2:  sqltypes.NewInt64(2),
			err: "strconv.ParseInt: parsing \"1.2\": invalid syntax",
		}, {
			// testing for error in types
			v1:  sqltypes.NewInt64(2),
			v2:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
			err: "strconv.ParseInt: parsing \"1.2\": invalid syntax",
		}, {
			// testing for uint/int
			v1:  sqltypes.NewUint64(4),
			v2:  sqltypes.NewInt64(5),
			out: sqltypes.NewFloat64(0.8),
		}, {
			// testing for uint/uint
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewFloat64(0.5),
		}, {
			// testing for float64/int64
			v1:  sqltypes.TestValue(querypb.Type_FLOAT64, "1.2"),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewFloat64(-0.6),
		}, {
			// testing for float64/uint64
			v1:  sqltypes.TestValue(querypb.Type_FLOAT64, "1.2"),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewFloat64(0.6),
		}, {
			// testing for overflow of float64
			v1:  sqltypes.NewFloat64(math.MaxFloat64),
			v2:  sqltypes.NewFloat64(0.5),
			err: "BIGINT is out of range in 1.7976931348623157e+308 / 0.5",
		}},
	}, {
		operator: "*",
		f:        Multiply,
		cases: []tcase{{
			//All Nulls
			v1:  sqltypes.NULL,
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// First value null.
			v1:  sqltypes.NewInt32(1),
			v2:  sqltypes.NULL,
			out: sqltypes.NULL,
		}, {
			// Second value null.
			v1:  sqltypes.NULL,
			v2:  sqltypes.NewInt32(1),
			out: sqltypes.NULL,
		}, {
			// case with negative value
			v1:  sqltypes.NewInt64(-1),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewInt64(2),
		}, {
			// testing for int64 overflow with min negative value
			v1:  sqltypes.NewInt64(math.MinInt64),
			v2:  sqltypes.NewInt64(1),
			out: sqltypes.NewInt64(math.MinInt64),
		}, {
			// testing for error in types
			v1:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
			v2:  sqltypes.NewInt64(2),
			err: "strconv.ParseInt: parsing \"1.2\": invalid syntax",
		}, {
			// testing for error in types
			v1:  sqltypes.NewInt64(2),
			v2:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
			err: "strconv.ParseInt: parsing \"1.2\": invalid syntax",
		}, {
			// testing for uint*int
			v1:  sqltypes.NewUint64(4),
			v2:  sqltypes.NewInt64(5),
			out: sqltypes.NewUint64(20),
		}, {
			// testing for uint*uint
			v1:  sqltypes.NewUint64(1),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewUint64(2),
		}, {
			// testing for float64*int64
			v1:  sqltypes.TestValue(querypb.Type_FLOAT64, "1.2"),
			v2:  sqltypes.NewInt64(-2),
			out: sqltypes.NewFloat64(-2.4),
		}, {
			// testing for float64*uint64
			v1:  sqltypes.TestValue(querypb.Type_FLOAT64, "1.2"),
			v2:  sqltypes.NewUint64(2),
			out: sqltypes.NewFloat64(2.4),
		}, {
			// testing for overflow of int64
			v1:  sqltypes.NewInt64(math.MaxInt64),
			v2:  sqltypes.NewInt64(2),
			err: "BIGINT value is out of range in 9223372036854775807 * 2",
		}, {
			// testing for underflow of uint64*max.uint64
			v1:  sqltypes.NewInt64(2),
			v2:  sqltypes.NewUint64(math.MaxUint64),
			err: "BIGINT UNSIGNED value is out of range in 18446744073709551615 * 2",
		}, {
			v1:  sqltypes.NewUint64(math.MaxUint64),
			v2:  sqltypes.NewUint64(1),
			out: sqltypes.NewUint64(math.MaxUint64),
		}, {
			//Checking whether maxInt value can be passed as uint value
			v1:  sqltypes.NewUint64(math.MaxInt64),
			v2:  sqltypes.NewInt64(3),
			err: "BIGINT UNSIGNED value is out of range in 9223372036854775807 * 3",
		}},
	}}

	for _, test := range tests {
		t.Run(test.operator, func(t *testing.T) {
			for _, tcase := range test.cases {
				name := fmt.Sprintf("%s%s%s", tcase.v1.String(), test.operator, tcase.v2.String())
				t.Run(name, func(t *testing.T) {
					got, err := test.f(tcase.v1, tcase.v2)
					if tcase.err == "" {
						require.NoError(t, err)
						require.Equal(t, tcase.out, got)
					} else {
						require.EqualError(t, err, tcase.err)
					}
				})
			}
		})
	}
}

func TestNullsafeAdd(t *testing.T) {
	tcases := []struct {
		v1, v2 sqltypes.Value
		out    sqltypes.Value
		err    error
	}{{
		// All nulls.
		v1:  sqltypes.NULL,
		v2:  sqltypes.NULL,
		out: sqltypes.NewInt64(0),
	}, {
		// First value null.
		v1:  sqltypes.NewInt32(1),
		v2:  sqltypes.NULL,
		out: sqltypes.NewInt64(1),
	}, {
		// Second value null.
		v1:  sqltypes.NULL,
		v2:  sqltypes.NewInt32(1),
		out: sqltypes.NewInt64(1),
	}, {
		// Normal case.
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewInt64(2),
		out: sqltypes.NewInt64(3),
	}, {
		// Make sure underlying error is returned for LHS.
		v1:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		v2:  sqltypes.NewInt64(2),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}, {
		// Make sure underlying error is returned for RHS.
		v1:  sqltypes.NewInt64(2),
		v2:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}, {
		// Make sure underlying error is returned while adding.
		v1:  sqltypes.NewInt64(-1),
		v2:  sqltypes.NewUint64(2),
		out: sqltypes.NewInt64(-9223372036854775808),
	}, {
		// Make sure underlying error is returned while converting.
		v1:  sqltypes.NewFloat64(1),
		v2:  sqltypes.NewFloat64(2),
		out: sqltypes.NewInt64(3),
	}}
	for _, tcase := range tcases {
		got := NullsafeAdd(tcase.v1, tcase.v2, querypb.Type_INT64)

		if !reflect.DeepEqual(got, tcase.out) {
			t.Errorf("NullsafeAdd(%v, %v): %v, want %v", printValue(tcase.v1), printValue(tcase.v2), printValue(got), printValue(tcase.out))
		}
	}
}

func TestNullsafeCompare(t *testing.T) {
	tcases := []struct {
		v1, v2 sqltypes.Value
		out    int
		err    error
	}{{
		// All nulls.
		v1:  sqltypes.NULL,
		v2:  sqltypes.NULL,
		out: 0,
	}, {
		// LHS null.
		v1:  sqltypes.NULL,
		v2:  sqltypes.NewInt64(1),
		out: -1,
	}, {
		// RHS null.
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NULL,
		out: 1,
	}, {
		// LHS Text
		v1:  sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		v2:  sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		err: vterrors.New(vtrpcpb.Code_UNKNOWN, "types are not comparable: VARCHAR vs VARCHAR"),
	}, {
		// Make sure underlying error is returned for LHS.
		v1:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		v2:  sqltypes.NewInt64(2),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}, {
		// Make sure underlying error is returned for RHS.
		v1:  sqltypes.NewInt64(2),
		v2:  sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}, {
		// Numeric equal.
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewUint64(1),
		out: 0,
	}, {
		// Numeric unequal.
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewUint64(2),
		out: -1,
	}, {
		// Non-numeric equal
		v1:  sqltypes.TestValue(querypb.Type_VARBINARY, "abcd"),
		v2:  sqltypes.TestValue(querypb.Type_BINARY, "abcd"),
		out: 0,
	}, {
		// Non-numeric unequal
		v1:  sqltypes.TestValue(querypb.Type_VARBINARY, "abcd"),
		v2:  sqltypes.TestValue(querypb.Type_BINARY, "bcde"),
		out: -1,
	}, {
		// Date/Time types
		v1:  sqltypes.TestValue(querypb.Type_DATETIME, "1000-01-01 00:00:00"),
		v2:  sqltypes.TestValue(querypb.Type_BINARY, "1000-01-01 00:00:00"),
		out: 0,
	}, {
		// Date/Time types
		v1:  sqltypes.TestValue(querypb.Type_DATETIME, "2000-01-01 00:00:00"),
		v2:  sqltypes.TestValue(querypb.Type_BINARY, "1000-01-01 00:00:00"),
		out: 1,
	}, {
		// Date/Time types
		v1:  sqltypes.TestValue(querypb.Type_DATETIME, "1000-01-01 00:00:00"),
		v2:  sqltypes.TestValue(querypb.Type_BINARY, "2000-01-01 00:00:00"),
		out: -1,
	}}
	for _, tcase := range tcases {
		got, err := NullsafeCompare(tcase.v1, tcase.v2)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("NullsafeCompare(%v, %v) error: %v, want %v", printValue(tcase.v1), printValue(tcase.v2), vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err != nil {
			continue
		}

		if got != tcase.out {
			t.Errorf("NullsafeCompare(%v, %v): %v, want %v", printValue(tcase.v1), printValue(tcase.v2), got, tcase.out)
		}
	}
}

func TestCast(t *testing.T) {
	tcases := []struct {
		typ querypb.Type
		v   sqltypes.Value
		out sqltypes.Value
		err error
	}{{
		typ: querypb.Type_VARCHAR,
		v:   sqltypes.NULL,
		out: sqltypes.NULL,
	}, {
		typ: querypb.Type_VARCHAR,
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "exact types"),
		out: sqltypes.TestValue(querypb.Type_VARCHAR, "exact types"),
	}, {
		typ: querypb.Type_INT64,
		v:   sqltypes.TestValue(querypb.Type_INT32, "32"),
		out: sqltypes.TestValue(querypb.Type_INT64, "32"),
	}, {
		typ: querypb.Type_INT24,
		v:   sqltypes.TestValue(querypb.Type_UINT64, "64"),
		out: sqltypes.TestValue(querypb.Type_INT24, "64"),
	}, {
		typ: querypb.Type_INT24,
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "bad int"),
		err: vterrors.New(vtrpcpb.Code_UNKNOWN, `strconv.ParseInt: parsing "bad int": invalid syntax`),
	}, {
		typ: querypb.Type_UINT64,
		v:   sqltypes.TestValue(querypb.Type_UINT32, "32"),
		out: sqltypes.TestValue(querypb.Type_UINT64, "32"),
	}, {
		typ: querypb.Type_UINT24,
		v:   sqltypes.TestValue(querypb.Type_INT64, "64"),
		out: sqltypes.TestValue(querypb.Type_UINT24, "64"),
	}, {
		typ: querypb.Type_UINT24,
		v:   sqltypes.TestValue(querypb.Type_INT64, "-1"),
		err: vterrors.New(vtrpcpb.Code_UNKNOWN, `strconv.ParseUint: parsing "-1": invalid syntax`),
	}, {
		typ: querypb.Type_FLOAT64,
		v:   sqltypes.TestValue(querypb.Type_INT64, "64"),
		out: sqltypes.TestValue(querypb.Type_FLOAT64, "64"),
	}, {
		typ: querypb.Type_FLOAT32,
		v:   sqltypes.TestValue(querypb.Type_FLOAT64, "64"),
		out: sqltypes.TestValue(querypb.Type_FLOAT32, "64"),
	}, {
		typ: querypb.Type_FLOAT32,
		v:   sqltypes.TestValue(querypb.Type_DECIMAL, "1.24"),
		out: sqltypes.TestValue(querypb.Type_FLOAT32, "1.24"),
	}, {
		typ: querypb.Type_FLOAT64,
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "1.25"),
		out: sqltypes.TestValue(querypb.Type_FLOAT64, "1.25"),
	}, {
		typ: querypb.Type_FLOAT64,
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "bad float"),
		err: vterrors.New(vtrpcpb.Code_UNKNOWN, `strconv.ParseFloat: parsing "bad float": invalid syntax`),
	}, {
		typ: querypb.Type_VARCHAR,
		v:   sqltypes.TestValue(querypb.Type_INT64, "64"),
		out: sqltypes.TestValue(querypb.Type_VARCHAR, "64"),
	}, {
		typ: querypb.Type_VARBINARY,
		v:   sqltypes.TestValue(querypb.Type_FLOAT64, "64"),
		out: sqltypes.TestValue(querypb.Type_VARBINARY, "64"),
	}, {
		typ: querypb.Type_VARBINARY,
		v:   sqltypes.TestValue(querypb.Type_DECIMAL, "1.24"),
		out: sqltypes.TestValue(querypb.Type_VARBINARY, "1.24"),
	}, {
		typ: querypb.Type_VARBINARY,
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "1.25"),
		out: sqltypes.TestValue(querypb.Type_VARBINARY, "1.25"),
	}, {
		typ: querypb.Type_VARCHAR,
		v:   sqltypes.TestValue(querypb.Type_VARBINARY, "valid string"),
		out: sqltypes.TestValue(querypb.Type_VARCHAR, "valid string"),
	}, {
		typ: querypb.Type_VARCHAR,
		v:   sqltypes.TestValue(sqltypes.Expression, "bad string"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "EXPRESSION(bad string) cannot be cast to VARCHAR"),
	}}
	for _, tcase := range tcases {
		got, err := Cast(tcase.v, tcase.typ)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("Cast(%v) error: %v, want %v", tcase.v, vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err != nil {
			continue
		}

		if !reflect.DeepEqual(got, tcase.out) {
			t.Errorf("Cast(%v): %v, want %v", tcase.v, got, tcase.out)
		}
	}
}

func TestToUint64(t *testing.T) {
	tcases := []struct {
		v   sqltypes.Value
		out uint64
		err error
	}{{
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "could not parse value: 'abcd'"),
	}, {
		v:   sqltypes.NewInt64(-1),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "negative number cannot be converted to unsigned: -1"),
	}, {
		v:   sqltypes.NewInt64(1),
		out: 1,
	}, {
		v:   sqltypes.NewUint64(1),
		out: 1,
	}}
	for _, tcase := range tcases {
		got, err := ToUint64(tcase.v)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("ToUint64(%v) error: %v, want %v", tcase.v, vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err != nil {
			continue
		}

		if got != tcase.out {
			t.Errorf("ToUint64(%v): %v, want %v", tcase.v, got, tcase.out)
		}
	}
}

func TestToInt64(t *testing.T) {
	tcases := []struct {
		v   sqltypes.Value
		out int64
		err error
	}{{
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "could not parse value: 'abcd'"),
	}, {
		v:   sqltypes.NewUint64(18446744073709551615),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "unsigned number overflows int64 value: 18446744073709551615"),
	}, {
		v:   sqltypes.NewInt64(1),
		out: 1,
	}, {
		v:   sqltypes.NewUint64(1),
		out: 1,
	}}
	for _, tcase := range tcases {
		got, err := ToInt64(tcase.v)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("ToInt64(%v) error: %v, want %v", tcase.v, vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err != nil {
			continue
		}

		if got != tcase.out {
			t.Errorf("ToInt64(%v): %v, want %v", tcase.v, got, tcase.out)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tcases := []struct {
		v   sqltypes.Value
		out float64
		err error
	}{{
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		out: 0,
	}, {
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "1.2"),
		out: 1.2,
	}, {
		v:   sqltypes.NewInt64(1),
		out: 1,
	}, {
		v:   sqltypes.NewUint64(1),
		out: 1,
	}, {
		v:   sqltypes.NewFloat64(1.2),
		out: 1.2,
	}, {
		v:   sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}}
	for _, tcase := range tcases {
		t.Run(tcase.v.String(), func(t *testing.T) {
			got, err := ToFloat64(tcase.v)
			if tcase.err != nil {
				require.EqualError(t, err, tcase.err.Error())
			} else {
				require.Equal(t, tcase.out, got)
			}
		})
	}
}

func TestToNative(t *testing.T) {
	testcases := []struct {
		in  sqltypes.Value
		out interface{}
	}{{
		in:  sqltypes.NULL,
		out: nil,
	}, {
		in:  sqltypes.TestValue(querypb.Type_INT8, "1"),
		out: int64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_INT16, "1"),
		out: int64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_INT24, "1"),
		out: int64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_INT32, "1"),
		out: int64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_INT64, "1"),
		out: int64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_UINT8, "1"),
		out: uint64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_UINT16, "1"),
		out: uint64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_UINT24, "1"),
		out: uint64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_UINT32, "1"),
		out: uint64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_UINT64, "1"),
		out: uint64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_FLOAT32, "1"),
		out: float64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_FLOAT64, "1"),
		out: float64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_TIMESTAMP, "2012-02-24 23:19:43"),
		out: []byte("2012-02-24 23:19:43"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_DATE, "2012-02-24"),
		out: []byte("2012-02-24"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_TIME, "23:19:43"),
		out: []byte("23:19:43"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_DATETIME, "2012-02-24 23:19:43"),
		out: []byte("2012-02-24 23:19:43"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_YEAR, "1"),
		out: uint64(1),
	}, {
		in:  sqltypes.TestValue(querypb.Type_DECIMAL, "1"),
		out: []byte("1"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_TEXT, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_BLOB, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_VARCHAR, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_VARBINARY, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_CHAR, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_BINARY, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_BIT, "1"),
		out: []byte("1"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_ENUM, "a"),
		out: []byte("a"),
	}, {
		in:  sqltypes.TestValue(querypb.Type_SET, "a"),
		out: []byte("a"),
	}}
	for _, tcase := range testcases {
		v, err := ToNative(tcase.in)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(v, tcase.out) {
			t.Errorf("%v.ToNative = %#v, want %#v", tcase.in, v, tcase.out)
		}
	}

	// Test Expression failure.
	_, err := ToNative(sqltypes.TestValue(querypb.Type_EXPRESSION, "aa"))
	want := vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "EXPRESSION(aa) cannot be converted to a go type")
	if !vterrors.Equals(err, want) {
		t.Errorf("ToNative(EXPRESSION): %v, want %v", vterrors.Print(err), vterrors.Print(want))
	}
}

var mustMatch = utils.MustMatchFn(
	[]interface{}{ // types with unexported fields
		evalResult{},
	},
	[]string{}, // ignored fields
)

func TestNewNumeric(t *testing.T) {
	tcases := []struct {
		v   sqltypes.Value
		out evalResult
		err error
	}{{
		v:   sqltypes.NewInt64(1),
		out: evalResult{typ: querypb.Type_INT64, ival: 1},
	}, {
		v:   sqltypes.NewUint64(1),
		out: evalResult{typ: querypb.Type_UINT64, uval: 1},
	}, {
		v:   sqltypes.NewFloat64(1),
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 1},
	}, {
		// For non-number type, Int64 is the default.
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "1"),
		out: evalResult{typ: querypb.Type_INT64, ival: 1},
	}, {
		// If Int64 can't work, we use Float64.
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "1.2"),
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 1.2},
	}, {
		// Only valid Int64 allowed if type is Int64.
		v:   sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}, {
		// Only valid Uint64 allowed if type is Uint64.
		v:   sqltypes.TestValue(querypb.Type_UINT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseUint: parsing \"1.2\": invalid syntax"),
	}, {
		// Only valid Float64 allowed if type is Float64.
		v:   sqltypes.TestValue(querypb.Type_FLOAT64, "abcd"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseFloat: parsing \"abcd\": invalid syntax"),
	}, {
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 0},
	}}
	for _, tcase := range tcases {
		got, err := newEvalResult(tcase.v)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("newEvalResult(%s) error: %v, want %v", printValue(tcase.v), vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err == nil {
			continue
		}

		mustMatch(t, tcase.out, got, "newEvalResult")
	}
}

func TestNewIntegralNumeric(t *testing.T) {
	tcases := []struct {
		v   sqltypes.Value
		out evalResult
		err error
	}{{
		v:   sqltypes.NewInt64(1),
		out: evalResult{typ: querypb.Type_INT64, ival: 1},
	}, {
		v:   sqltypes.NewUint64(1),
		out: evalResult{typ: querypb.Type_UINT64, uval: 1},
	}, {
		v:   sqltypes.NewFloat64(1),
		out: evalResult{typ: querypb.Type_INT64, ival: 1},
	}, {
		// For non-number type, Int64 is the default.
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "1"),
		out: evalResult{typ: querypb.Type_INT64, ival: 1},
	}, {
		// If Int64 can't work, we use Uint64.
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "18446744073709551615"),
		out: evalResult{typ: querypb.Type_UINT64, uval: 18446744073709551615},
	}, {
		// Only valid Int64 allowed if type is Int64.
		v:   sqltypes.TestValue(querypb.Type_INT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseInt: parsing \"1.2\": invalid syntax"),
	}, {
		// Only valid Uint64 allowed if type is Uint64.
		v:   sqltypes.TestValue(querypb.Type_UINT64, "1.2"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "strconv.ParseUint: parsing \"1.2\": invalid syntax"),
	}, {
		v:   sqltypes.TestValue(querypb.Type_VARCHAR, "abcd"),
		err: vterrors.New(vtrpcpb.Code_INVALID_ARGUMENT, "could not parse value: 'abcd'"),
	}}
	for _, tcase := range tcases {
		got, err := newIntegralNumeric(tcase.v)
		if err != nil && !vterrors.Equals(err, tcase.err) {
			t.Errorf("newIntegralNumeric(%s) error: %v, want %v", printValue(tcase.v), vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err == nil {
			continue
		}

		mustMatch(t, tcase.out, got, "newIntegralNumeric")
	}
}

func TestAddNumeric(t *testing.T) {
	tcases := []struct {
		v1, v2 evalResult
		out    evalResult
		err    error
	}{{
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 2},
		out: evalResult{typ: querypb.Type_INT64, ival: 3},
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: evalResult{typ: querypb.Type_UINT64, uval: 3},
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 3},
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: evalResult{typ: querypb.Type_UINT64, uval: 3},
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 3},
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 3},
	}, {
		// Int64 overflow.
		v1:  evalResult{typ: querypb.Type_INT64, ival: 9223372036854775807},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 9223372036854775809},
	}, {
		// Int64 underflow.
		v1:  evalResult{typ: querypb.Type_INT64, ival: -9223372036854775807},
		v2:  evalResult{typ: querypb.Type_INT64, ival: -2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: -9223372036854775809},
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: -1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 18446744073709551617},
	}, {
		// Uint64 overflow.
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 18446744073709551615},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: evalResult{typ: querypb.Type_FLOAT64, fval: 18446744073709551617},
	}}
	for _, tcase := range tcases {
		got := addNumeric(tcase.v1, tcase.v2)

		mustMatch(t, tcase.out, got, "addNumeric")
	}
}

func TestPrioritize(t *testing.T) {
	ival := evalResult{typ: querypb.Type_INT64, ival: -1}
	uval := evalResult{typ: querypb.Type_UINT64, uval: 1}
	fval := evalResult{typ: querypb.Type_FLOAT64, fval: 1.2}
	textIntval := evalResult{typ: querypb.Type_VARBINARY, bytes: []byte("-1")}
	textFloatval := evalResult{typ: querypb.Type_VARBINARY, bytes: []byte("1.2")}

	tcases := []struct {
		v1, v2     evalResult
		out1, out2 evalResult
	}{{
		v1:   ival,
		v2:   uval,
		out1: uval,
		out2: ival,
	}, {
		v1:   ival,
		v2:   fval,
		out1: fval,
		out2: ival,
	}, {
		v1:   uval,
		v2:   ival,
		out1: uval,
		out2: ival,
	}, {
		v1:   uval,
		v2:   fval,
		out1: fval,
		out2: uval,
	}, {
		v1:   fval,
		v2:   ival,
		out1: fval,
		out2: ival,
	}, {
		v1:   fval,
		v2:   uval,
		out1: fval,
		out2: uval,
	}, {
		v1:   textIntval,
		v2:   ival,
		out1: ival,
		out2: ival,
	}, {
		v1:   ival,
		v2:   textFloatval,
		out1: fval,
		out2: ival,
	}}
	for _, tcase := range tcases {
		t.Run(tcase.v1.Value().String()+" - "+tcase.v2.Value().String(), func(t *testing.T) {
			got1, got2 := makeNumericAndprioritize(tcase.v1, tcase.v2)
			mustMatch(t, tcase.out1, got1, "makeNumericAndprioritize")
			mustMatch(t, tcase.out2, got2, "makeNumericAndprioritize")
		})
	}
}

func TestCastFromNumeric(t *testing.T) {
	tcases := []struct {
		typ querypb.Type
		v   evalResult
		out sqltypes.Value
		err error
	}{{
		typ: querypb.Type_INT64,
		v:   evalResult{typ: querypb.Type_INT64, ival: 1},
		out: sqltypes.NewInt64(1),
	}, {
		typ: querypb.Type_INT64,
		v:   evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: sqltypes.NewInt64(1),
	}, {
		typ: querypb.Type_INT64,
		v:   evalResult{typ: querypb.Type_FLOAT64, fval: 1.2e-16},
		out: sqltypes.NewInt64(0),
	}, {
		typ: querypb.Type_UINT64,
		v:   evalResult{typ: querypb.Type_INT64, ival: 1},
		out: sqltypes.NewUint64(1),
	}, {
		typ: querypb.Type_UINT64,
		v:   evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: sqltypes.NewUint64(1),
	}, {
		typ: querypb.Type_UINT64,
		v:   evalResult{typ: querypb.Type_FLOAT64, fval: 1.2e-16},
		out: sqltypes.NewUint64(0),
	}, {
		typ: querypb.Type_FLOAT64,
		v:   evalResult{typ: querypb.Type_INT64, ival: 1},
		out: sqltypes.TestValue(querypb.Type_FLOAT64, "1"),
	}, {
		typ: querypb.Type_FLOAT64,
		v:   evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: sqltypes.TestValue(querypb.Type_FLOAT64, "1"),
	}, {
		typ: querypb.Type_FLOAT64,
		v:   evalResult{typ: querypb.Type_FLOAT64, fval: 1.2e-16},
		out: sqltypes.TestValue(querypb.Type_FLOAT64, "1.2e-16"),
	}, {
		typ: querypb.Type_DECIMAL,
		v:   evalResult{typ: querypb.Type_INT64, ival: 1},
		out: sqltypes.TestValue(querypb.Type_DECIMAL, "1"),
	}, {
		typ: querypb.Type_DECIMAL,
		v:   evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: sqltypes.TestValue(querypb.Type_DECIMAL, "1"),
	}, {
		// For float, we should not use scientific notation.
		typ: querypb.Type_DECIMAL,
		v:   evalResult{typ: querypb.Type_FLOAT64, fval: 1.2e-16},
		out: sqltypes.TestValue(querypb.Type_DECIMAL, "0.00000000000000012"),
	}}
	for _, tcase := range tcases {
		got := castFromNumeric(tcase.v, tcase.typ)

		if !reflect.DeepEqual(got, tcase.out) {
			t.Errorf("castFromNumeric(%v, %v): %v, want %v", tcase.v, tcase.typ, printValue(got), printValue(tcase.out))
		}
	}
}

func TestCompareNumeric(t *testing.T) {
	tcases := []struct {
		v1, v2 evalResult
		out    int
	}{{
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 2},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 1},
		out: 1,
	}, {
		// Special case.
		v1:  evalResult{typ: querypb.Type_INT64, ival: -1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 2},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_INT64, ival: 2},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		out: 1,
	}, {
		// Special case.
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: -1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		v2:  evalResult{typ: querypb.Type_INT64, ival: 1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		v2:  evalResult{typ: querypb.Type_UINT64, uval: 1},
		out: 1,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		out: 0,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		out: -1,
	}, {
		v1:  evalResult{typ: querypb.Type_FLOAT64, fval: 2},
		v2:  evalResult{typ: querypb.Type_FLOAT64, fval: 1},
		out: 1,
	}}
	for _, tcase := range tcases {
		got := compareNumeric(tcase.v1, tcase.v2)
		if got != tcase.out {
			t.Errorf("equalNumeric(%v, %v): %v, want %v", tcase.v1, tcase.v2, got, tcase.out)
		}
	}
}

func TestMin(t *testing.T) {
	tcases := []struct {
		v1, v2 sqltypes.Value
		min    sqltypes.Value
		err    error
	}{{
		v1:  sqltypes.NULL,
		v2:  sqltypes.NULL,
		min: sqltypes.NULL,
	}, {
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NULL,
		min: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.NULL,
		v2:  sqltypes.NewInt64(1),
		min: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewInt64(2),
		min: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.NewInt64(2),
		v2:  sqltypes.NewInt64(1),
		min: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewInt64(1),
		min: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.TestValue(querypb.Type_VARCHAR, "aa"),
		v2:  sqltypes.TestValue(querypb.Type_VARCHAR, "aa"),
		err: vterrors.New(vtrpcpb.Code_UNKNOWN, "types are not comparable: VARCHAR vs VARCHAR"),
	}}
	for _, tcase := range tcases {
		v, err := Min(tcase.v1, tcase.v2)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("Min error: %v, want %v", vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err != nil {
			continue
		}

		if !reflect.DeepEqual(v, tcase.min) {
			t.Errorf("Min(%v, %v): %v, want %v", tcase.v1, tcase.v2, v, tcase.min)
		}
	}
}

func TestMax(t *testing.T) {
	tcases := []struct {
		v1, v2 sqltypes.Value
		max    sqltypes.Value
		err    error
	}{{
		v1:  sqltypes.NULL,
		v2:  sqltypes.NULL,
		max: sqltypes.NULL,
	}, {
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NULL,
		max: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.NULL,
		v2:  sqltypes.NewInt64(1),
		max: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewInt64(2),
		max: sqltypes.NewInt64(2),
	}, {
		v1:  sqltypes.NewInt64(2),
		v2:  sqltypes.NewInt64(1),
		max: sqltypes.NewInt64(2),
	}, {
		v1:  sqltypes.NewInt64(1),
		v2:  sqltypes.NewInt64(1),
		max: sqltypes.NewInt64(1),
	}, {
		v1:  sqltypes.TestValue(querypb.Type_VARCHAR, "aa"),
		v2:  sqltypes.TestValue(querypb.Type_VARCHAR, "aa"),
		err: vterrors.New(vtrpcpb.Code_UNKNOWN, "types are not comparable: VARCHAR vs VARCHAR"),
	}}
	for _, tcase := range tcases {
		v, err := Max(tcase.v1, tcase.v2)
		if !vterrors.Equals(err, tcase.err) {
			t.Errorf("Max error: %v, want %v", vterrors.Print(err), vterrors.Print(tcase.err))
		}
		if tcase.err != nil {
			continue
		}

		if !reflect.DeepEqual(v, tcase.max) {
			t.Errorf("Max(%v, %v): %v, want %v", tcase.v1, tcase.v2, v, tcase.max)
		}
	}
}

func printValue(v sqltypes.Value) string {
	return fmt.Sprintf("%v:%q", v.Type(), v.ToBytes())
}

// These benchmarks show that using existing ASCII representations
// for numbers is about 6x slower than using native representations.
// However, 229ns is still a negligible time compared to the cost of
// other operations. The additional complexity of introducing native
// types is currently not worth it. So, we'll stay with the existing
// ASCII representation for now. Using interfaces is more expensive
// than native representation of values. This is probably because
// interfaces also allocate memory, and also perform type assertions.
// Actual benchmark is based on NoNative. So, the numbers are similar.
// Date: 6/4/17
// Version: go1.8
// BenchmarkAddActual-8            10000000               263 ns/op
// BenchmarkAddNoNative-8          10000000               228 ns/op
// BenchmarkAddNative-8            50000000                40.0 ns/op
// BenchmarkAddGoInterface-8       30000000                52.4 ns/op
// BenchmarkAddGoNonInterface-8    2000000000               1.00 ns/op
// BenchmarkAddGo-8                2000000000               1.00 ns/op
func BenchmarkAddActual(b *testing.B) {
	v1 := sqltypes.MakeTrusted(querypb.Type_INT64, []byte("1"))
	v2 := sqltypes.MakeTrusted(querypb.Type_INT64, []byte("12"))
	for i := 0; i < b.N; i++ {
		v1 = NullsafeAdd(v1, v2, querypb.Type_INT64)
	}
}

func BenchmarkAddNoNative(b *testing.B) {
	v1 := sqltypes.MakeTrusted(querypb.Type_INT64, []byte("1"))
	v2 := sqltypes.MakeTrusted(querypb.Type_INT64, []byte("12"))
	for i := 0; i < b.N; i++ {
		iv1, _ := ToInt64(v1)
		iv2, _ := ToInt64(v2)
		v1 = sqltypes.MakeTrusted(querypb.Type_INT64, strconv.AppendInt(nil, iv1+iv2, 10))
	}
}

func BenchmarkAddNative(b *testing.B) {
	v1 := makeNativeInt64(1)
	v2 := makeNativeInt64(12)
	for i := 0; i < b.N; i++ {
		iv1 := int64(binary.BigEndian.Uint64(v1.Raw()))
		iv2 := int64(binary.BigEndian.Uint64(v2.Raw()))
		v1 = makeNativeInt64(iv1 + iv2)
	}
}

func makeNativeInt64(v int64) sqltypes.Value {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return sqltypes.MakeTrusted(querypb.Type_INT64, buf)
}

func BenchmarkAddGoInterface(b *testing.B) {
	var v1, v2 interface{}
	v1 = int64(1)
	v2 = int64(2)
	for i := 0; i < b.N; i++ {
		v1 = v1.(int64) + v2.(int64)
	}
}

func BenchmarkAddGoNonInterface(b *testing.B) {
	v1 := evalResult{typ: querypb.Type_INT64, ival: 1}
	v2 := evalResult{typ: querypb.Type_INT64, ival: 12}
	for i := 0; i < b.N; i++ {
		if v1.typ != querypb.Type_INT64 {
			b.Error("type assertion failed")
		}
		if v2.typ != querypb.Type_INT64 {
			b.Error("type assertion failed")
		}
		v1 = evalResult{typ: querypb.Type_INT64, ival: v1.ival + v2.ival}
	}
}

func BenchmarkAddGo(b *testing.B) {
	v1 := int64(1)
	v2 := int64(2)
	for i := 0; i < b.N; i++ {
		v1 += v2
	}
}
