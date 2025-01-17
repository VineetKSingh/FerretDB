// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operators

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// sum represents $sum aggregation operator.
type sum struct {
	expression types.Expression
	number     any
}

// newSum creates a new $sum aggregation operator.
func newSum(accumulation *types.Document) (Accumulator, error) {
	expression := must.NotFail(accumulation.Get("$sum"))
	accumulator := new(sum)

	switch expr := expression.(type) {
	case *types.Array:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupUnaryOperator,
			"The $sum accumulator is a unary accumulator",
			"$sum (accumulator)",
		)
	case float64:
		accumulator.number = expr
	case string:
		var err error
		if accumulator.expression, err = types.NewExpression(expr); err != nil {
			// $sum returns 0 on non-existent field.
			accumulator.number = int32(0)
		}
	case int32, int64:
		accumulator.number = expr
	default:
		accumulator.number = int32(0)
		// $sum returns 0 on non-numeric field
	}

	return accumulator, nil
}

// Accumulate implements Accumulator interface.
func (s *sum) Accumulate(ctx context.Context, groupID any, grouped []*types.Document) (any, error) {
	if s.expression != nil {
		var values []any

		for _, doc := range grouped {
			v := s.expression.Evaluate(doc)
			values = append(values, v)
		}

		return sumNumbers(values...), nil
	}

	switch number := s.number.(type) {
	case float64, int32, int64:
		// Below is equivalent of len(grouped)*number,
		// with conversion handling upon overflow of int32 and int64.
		// For example, { $sum: 1 } is equivalent of $count.
		numbers := make([]any, len(grouped))
		for i := 0; i < len(grouped); i++ {
			numbers[i] = number
		}

		return sumNumbers(numbers...), nil
	}

	// $sum returns 0 on non-existent and non-numeric field.
	return int32(0), nil
}

// check interfaces
var (
	_ Accumulator = (*sum)(nil)
)
