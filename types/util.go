/*
 * Collection of utility methods to manipulate type data
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package types

import ()

// Determine the 'truthiness' of the data value according to specification
func IsTruthy(val *DataType) bool {
    switch (*val).(type) {
    case UndefinedType, NullType:
        return false
    case BooleanType:
        return (*val).Native().(bool)
    case IntegerType:
        return (*val).Native().(int64) != 0
    case NumberType:
        n := (*val).Native().(float64)
        return n != 0 && n == n // NaN check
    case StringType:
        return len((*val).Native().(string)) > 0
    }

    // Objects are truthy
    return true
}
