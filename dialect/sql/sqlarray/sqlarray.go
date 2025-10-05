package sqlarray

import (
	"errors"
	"strconv"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

// checkArraySupport adds an error if the dialect doesn't support arrays.
func checkArraySupport(b *sql.Builder) {
	switch b.Dialect() {
	case dialect.MySQL, dialect.SQLite:
		b.AddError(errors.New("arrays are not supported by " + b.Dialect()))
	}
}

// Array wraps the expression with the ARRAY function (GoogleSQL).
func Array(e string) func(*sql.Builder) {
	return func(b *sql.Builder) {
		checkArraySupport(b)
		b.WriteString("ARRAY(")
		b.WriteString(e)
		b.WriteString(")")
	}
}

// ArrayOf creates an ARRAY expression from a sql.Querier.
func ArrayOf(q sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.WriteString("ARRAY(")
		b.Join(q)
		b.WriteString(")")
	})
}

// Literal creates an array literal from values.
// For PostgreSQL: Literal(1, 2, 3) produces '{1,2,3}'
// For Spanner: Literal(1, 2, 3) produces [1, 2, 3]
func Literal(values ...any) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("'{")
			for i, v := range values {
				if i > 0 {
					b.WriteString(",")
				}
				// For PostgreSQL, we need to format values properly
				switch v := v.(type) {
				case string:
					b.WriteString(`"` + v + `"`)
				case nil:
					b.WriteString("NULL")
				default:
					b.Arg(v)
				}
			}
			b.WriteString("}'")
		default:
			// Spanner/default syntax
			b.WriteString("[")
			for i, v := range values {
				if i > 0 {
					b.WriteString(", ")
				}
				b.Arg(v)
			}
			b.WriteString("]")
		}
	})
}

// ArrayConstructor creates an array constructor from values.
// For PostgreSQL: ArrayConstructor(1, 2, 3) produces ARRAY[1,2,3]
// For Spanner: ArrayConstructor(1, 2, 3) produces [1, 2, 3]
func ArrayConstructor(values ...any) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("ARRAY[")
			for i, v := range values {
				if i > 0 {
					b.WriteString(",")
				}
				b.Arg(v)
			}
			b.WriteString("]")
		default:
			// Spanner syntax
			b.WriteString("[")
			for i, v := range values {
				if i > 0 {
					b.WriteString(", ")
				}
				b.Arg(v)
			}
			b.WriteString("]")
		}
	})
}

// ArrayAgg creates an ARRAY_AGG expression with optional ordering.
func ArrayAgg(expr sql.Querier, orderBy ...sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.WriteString("ARRAY_AGG(")
		b.Join(expr)
		if len(orderBy) > 0 {
			b.WriteString(" ORDER BY ")
			for i, order := range orderBy {
				if i > 0 {
					b.WriteString(", ")
				}
				b.Join(order)
			}
		}
		b.WriteString(")")
	})
}

// UnnestBuilder is a builder for the UNNEST expression.
type UnnestBuilder struct {
	sql.Builder
	sql.TableView
	expr   sql.Querier
	alias  string
	offset string
}

// Unnest wraps the array expression with the UNNEST operator (GoogleSQL).
func Unnest(array sql.Querier) *UnnestBuilder {
	return &UnnestBuilder{
		expr: array,
	}
}

func (u *UnnestBuilder) As(alias string) *UnnestBuilder {
	u.alias = alias
	return u
}

func (u *UnnestBuilder) WithOffset() *UnnestBuilder {
	u.offset = "offset"
	return u
}

func (u *UnnestBuilder) WithOffsetAs(offset string) *UnnestBuilder {
	u.offset = offset
	return u
}

func (u *UnnestBuilder) Query() string {
	b := u.Builder.Clone()
	checkArraySupport(&b)
	
	// Both PostgreSQL and Spanner use the same UNNEST(array) syntax
	b.WriteString("UNNEST(")
	b.Join(u.expr)
	b.WriteString(")")
	
	if u.alias != "" {
		b.WriteString(" AS ")
		b.WriteString(u.alias)
	}
	
	if u.offset != "" {
		// Only the offset clause differs between dialects
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString(" WITH ORDINALITY")
		default:
			// Spanner/GoogleSQL
			b.WriteString(" WITH OFFSET")
		}
		
		if u.offset != "offset" {
			b.WriteString(" AS ")
			b.WriteString(u.offset)
		}
	}
	
	return b.String()
}

// Element access functions

// Index accesses an array element by index.
// For PostgreSQL: uses 1-based indexing: my_array[1]
// For Spanner: uses 0-based OFFSET: my_array[OFFSET(0)]
func Index(array sql.Querier, index int) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(array)
		switch b.Dialect() {
		case dialect.Postgres:
			// PostgreSQL uses 1-based indexing
			b.WriteString("[")
			b.WriteString(strconv.Itoa(index))
			b.WriteString("]")
		default:
			// Spanner uses OFFSET (0-based)
			b.WriteString("[OFFSET(")
			b.WriteString(strconv.Itoa(index))
			b.WriteString(")]")
		}
	})
}

// Slice accesses an array slice.
// For PostgreSQL: Slice("my_array", 2, 4) produces my_array[2:4]
// For Spanner: Slice("my_array", 2, 4) produces ARRAY_SLICE(my_array, 2, 4)
func Slice(array sql.Querier, start, end int) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Join(array)
			b.WriteString("[")
			b.WriteString(strconv.Itoa(start))
			b.WriteString(":")
			b.WriteString(strconv.Itoa(end))
			b.WriteString("]")
		default:
			// Spanner ARRAY_SLICE
			b.WriteString("ARRAY_SLICE(")
			b.Join(array)
			b.WriteString(", ")
			b.WriteString(strconv.Itoa(start))
			b.WriteString(", ")
			b.WriteString(strconv.Itoa(end))
			b.WriteString(")")
		}
	})
}

// Offset accesses an array element using zero-based indexing (Spanner-specific).
// Example: Offset("my_array", 0) produces my_array[OFFSET(0)]
func Offset(array sql.Querier, index int) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(array)
		b.WriteString("[OFFSET(")
		b.WriteString(strconv.Itoa(index))
		b.WriteString(")]")
	})
}

// SafeOffset accesses an array element using zero-based indexing, returns NULL if out of bounds.
// Example: SafeOffset("my_array", 0) produces my_array[SAFE_OFFSET(0)]
func SafeOffset(array sql.Querier, index int) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(array)
		b.WriteString("[SAFE_OFFSET(")
		b.WriteString(strconv.Itoa(index))
		b.WriteString(")]")
	})
}

// Ordinal accesses an array element using one-based indexing.
// Example: Ordinal("my_array", 1) produces my_array[ORDINAL(1)]
func Ordinal(array sql.Querier, index int) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(array)
		b.WriteString("[ORDINAL(")
		b.WriteString(strconv.Itoa(index))
		b.WriteString(")]")
	})
}

// SafeOrdinal accesses an array element using one-based indexing, returns NULL if out of bounds.
// Example: SafeOrdinal("my_array", 1) produces my_array[SAFE_ORDINAL(1)]
func SafeOrdinal(array sql.Querier, index int) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(array)
		b.WriteString("[SAFE_ORDINAL(")
		b.WriteString(strconv.Itoa(index))
		b.WriteString(")]")
	})
}

// Array predicate and comparison functions

// Overlap checks if two arrays have common elements.
// For PostgreSQL: Overlap(array1, array2) produces array1 && array2
// For Spanner: uses ARRAY_INCLUDES_ANY
func Overlap(left, right sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Join(left)
			b.WriteString(" && ")
			b.Join(right)
		default:
			// For Spanner, use ARRAY_INCLUDES_ANY
			b.WriteString("ARRAY_INCLUDES_ANY(")
			b.Join(left)
			b.WriteString(", ")
			b.Join(right)
			b.WriteString(")")
		}
	})
}

// ContainsArray checks if left array contains all elements of right array.
// For PostgreSQL: ContainsArray(array1, array2) produces array1 @> array2
// For Spanner: uses ARRAY_INCLUDES_ALL
func ContainsArray(left, right sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Join(left)
			b.WriteString(" @> ")
			b.Join(right)
		default:
			// For Spanner, use ARRAY_INCLUDES_ALL
			b.WriteString("ARRAY_INCLUDES_ALL(")
			b.Join(left)
			b.WriteString(", ")
			b.Join(right)
			b.WriteString(")")
		}
	})
}

// ContainedBy checks if left array is contained by right array.
// For PostgreSQL: ContainedBy(array1, array2) produces array1 <@ array2
// For Spanner: uses ARRAY_INCLUDES_ALL with reversed arguments
func ContainedBy(left, right sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Join(left)
			b.WriteString(" <@ ")
			b.Join(right)
		default:
			// For Spanner, use ARRAY_INCLUDES_ALL with reversed arguments
			b.WriteString("ARRAY_INCLUDES_ALL(")
			b.Join(right)
			b.WriteString(", ")
			b.Join(left)
			b.WriteString(")")
		}
	})
}

// ArrayEqual checks if two arrays are equal.
func ArrayEqual(left, right sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(left)
		b.WriteString(" = ")
		b.Join(right)
	})
}

// ArrayNotEqual checks if two arrays are not equal.
func ArrayNotEqual(left, right sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		b.Join(left)
		b.WriteString(" <> ")
		b.Join(right)
	})
}

// Array manipulation functions

// Length returns the length of an array.
// For PostgreSQL: Length("my_array") produces ARRAY_LENGTH(my_array, 1)
// For Spanner: Length("my_array") produces ARRAY_LENGTH(my_array)
func Length(array sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(", 1)")
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(")")
		}
	})
}

// Concat concatenates multiple arrays into one.
// For PostgreSQL: uses || operator between pairs
// For Spanner: Concat(array1, array2, array3) produces ARRAY_CONCAT(array1, array2, array3)
func Concat(arrays ...sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			if len(arrays) == 0 {
				return
			}
			if len(arrays) == 1 {
				b.Join(arrays[0])
				return
			}
			// For PostgreSQL, use || operator between arrays
			for i, array := range arrays {
				if i > 0 {
					b.WriteString(" || ")
				}
				b.Join(array)
			}
		default:
			b.WriteString("ARRAY_CONCAT(")
			for i, array := range arrays {
				if i > 0 {
					b.WriteString(", ")
				}
				b.Join(array)
			}
			b.WriteString(")")
		}
	})
}

// ToString converts an array to a string with the specified delimiter.
// For PostgreSQL: ToString("my_array", ",") produces ARRAY_TO_STRING(my_array, ',')
// For Spanner: ToString("my_array", ",") produces ARRAY_TO_STRING(my_array, ',')
func ToString(array sql.Querier, delimiter string) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.WriteString("ARRAY_TO_STRING(")
		b.Join(array)
		b.WriteString(", ")
		b.Arg(delimiter)
		b.WriteString(")")
	})
}

// ToStringWithNull converts an array to a string with the specified delimiter and null string.
// For PostgreSQL: ToStringWithNull("my_array", ",", "NULL") produces ARRAY_TO_STRING(my_array, ',', 'NULL')
// For Spanner: ToStringWithNull("my_array", ",", "NULL") produces ARRAY_TO_STRING(my_array, ',', 'NULL')
func ToStringWithNull(array sql.Querier, delimiter, nullString string) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.WriteString("ARRAY_TO_STRING(")
		b.Join(array)
		b.WriteString(", ")
		b.Arg(delimiter)
		b.WriteString(", ")
		b.Arg(nullString)
		b.WriteString(")")
	})
}

// Reverse reverses the order of elements in an array.
// Example: Reverse("my_array") produces ARRAY_REVERSE(my_array)
func Reverse(array sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		b.WriteString("ARRAY_REVERSE(")
		b.Join(array)
		b.WriteString(")")
	})
}

// Additional array functions

// First returns the first element of an array.
// For PostgreSQL: uses array[1]
// For Spanner: uses ARRAY_FIRST(array)
func First(array sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Join(array)
			b.WriteString("[1]")
		default:
			b.WriteString("ARRAY_FIRST(")
			b.Join(array)
			b.WriteString(")")
		}
	})
}

// Last returns the last element of an array.
// For PostgreSQL: uses array[array_length(array, 1)]
// For Spanner: uses ARRAY_LAST(array)
func Last(array sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Join(array)
			b.WriteString("[array_length(")
			b.Join(array)
			b.WriteString(", 1)]")
		default:
			b.WriteString("ARRAY_LAST(")
			b.Join(array)
			b.WriteString(")")
		}
	})
}

// Position returns the position of the first occurrence of a value in an array.
// For PostgreSQL: Position("my_array", "value") produces ARRAY_POSITION(my_array, 'value')
// For Spanner: uses UNNEST with OFFSET approach (no built-in equivalent)
func Position(array sql.Querier, value any) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("ARRAY_POSITION(")
			b.Join(array)
			b.WriteString(", ")
			b.Arg(value)
			b.WriteString(")")
		default:
			// For Spanner, use UNNEST with OFFSET (corrected logic)
			b.WriteString("(SELECT offset FROM UNNEST(")
			b.Join(array)
			b.WriteString(") AS elem WITH OFFSET WHERE elem = ")
			b.Arg(value)
			b.WriteString(" LIMIT 1)")
		}
	})
}

// Append adds an element to the end of an array.
// For PostgreSQL: Append("my_array", "value") produces ARRAY_APPEND(my_array, 'value')
// For Spanner: uses ARRAY_CONCAT
func Append(array sql.Querier, value any) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("ARRAY_APPEND(")
			b.Join(array)
			b.WriteString(", ")
			b.Arg(value)
			b.WriteString(")")
		default:
			// For Spanner, concatenate with single-element array
			b.WriteString("ARRAY_CONCAT(")
			b.Join(array)
			b.WriteString(", [")
			b.Arg(value)
			b.WriteString("])")
		}
	})
}

// Prepend adds an element to the beginning of an array.
// For PostgreSQL: Prepend("value", "my_array") produces ARRAY_PREPEND('value', my_array)
// For Spanner: uses ARRAY_CONCAT
func Prepend(value any, array sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("ARRAY_PREPEND(")
			b.Arg(value)
			b.WriteString(", ")
			b.Join(array)
			b.WriteString(")")
		default:
			// For Spanner, concatenate single-element array with main array
			b.WriteString("ARRAY_CONCAT([")
			b.Arg(value)
			b.WriteString("], ")
			b.Join(array)
			b.WriteString(")")
		}
	})
}

// Remove removes all occurrences of a value from an array.
// For PostgreSQL: Remove("my_array", "value") produces ARRAY_REMOVE(my_array, 'value')
// For Spanner: uses ARRAY_FILTER
func Remove(array sql.Querier, value any) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("ARRAY_REMOVE(")
			b.Join(array)
			b.WriteString(", ")
			b.Arg(value)
			b.WriteString(")")
		default:
			// For Spanner, use ARRAY_FILTER
			b.WriteString("ARRAY_FILTER(")
			b.Join(array)
			b.WriteString(", elem -> elem != ")
			b.Arg(value)
			b.WriteString(")")
		}
	})
}

// StringToArray converts a delimited string to an array.
// For PostgreSQL: StringToArray("a,b,c", ",") produces STRING_TO_ARRAY('a,b,c', ',')
// For Spanner: uses SPLIT function
func StringToArray(str, delimiter string) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("STRING_TO_ARRAY(")
			b.Arg(str)
			b.WriteString(", ")
			b.Arg(delimiter)
			b.WriteString(")")
		default:
			// For Spanner, use SPLIT function
			b.WriteString("SPLIT(")
			b.Arg(str)
			b.WriteString(", ")
			b.Arg(delimiter)
			b.WriteString(")")
		}
	})
}

// Additional Spanner-specific functions that have no PostgreSQL equivalent

// Transform transforms array elements using a lambda expression (Spanner only).
// Example: Transform("my_array", "elem", "elem * 2") produces ARRAY_TRANSFORM(my_array, elem -> elem * 2)
func Transform(array sql.Querier, var_, expr string) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			// PostgreSQL doesn't have a direct equivalent, would need more complex SQL
			// For now, just return the array unchanged as a fallback
			b.Join(array)
		default:
			b.WriteString("ARRAY_TRANSFORM(")
			b.Join(array)
			b.WriteString(", ")
			b.WriteString(var_)
			b.WriteString(" -> ")
			b.WriteString(expr)
			b.WriteString(")")
		}
	})
}

// Filter filters array elements based on a condition (Spanner only).
// Example: Filter("my_array", "elem", "elem > 5") produces ARRAY_FILTER(my_array, elem -> elem > 5)
func Filter(array sql.Querier, var_, condition string) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			// PostgreSQL doesn't have a direct equivalent
			b.Join(array)
		default:
			b.WriteString("ARRAY_FILTER(")
			b.Join(array)
			b.WriteString(", ")
			b.WriteString(var_)
			b.WriteString(" -> ")
			b.WriteString(condition)
			b.WriteString(")")
		}
	})
}

// IsDistinct checks if all array elements are unique (Spanner only).
// Example: IsDistinct("my_array") produces ARRAY_IS_DISTINCT(my_array)
func IsDistinct(array sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			// PostgreSQL equivalent using array comparison
			b.Join(array)
			b.WriteString(" = (SELECT array_agg(DISTINCT elem) FROM unnest(")
			b.Join(array)
			b.WriteString(") AS elem)")
		default:
			b.WriteString("ARRAY_IS_DISTINCT(")
			b.Join(array)
			b.WriteString(")")
		}
	})
}

// Array predicate functions for filtering and conditions

// Contains checks if an array contains a specific value.
// For PostgreSQL: Contains("my_array", 5) produces 5 = ANY(my_array)
// For Spanner: Contains("my_array", 5) produces ARRAY_INCLUDES(my_array, 5)
func Contains(array sql.Querier, value any) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Arg(value)
			b.WriteString(" = ANY(")
			b.Join(array)
			b.WriteString(")")
		default:
			// For Spanner, use ARRAY_INCLUDES
			b.WriteString("ARRAY_INCLUDES(")
			b.Join(array)
			b.WriteString(", ")
			b.Arg(value)
			b.WriteString(")")
		}
	})
}

// NotContains checks if an array does not contain a specific value.
// For PostgreSQL: NotContains("my_array", 5) produces 5 <> ALL(my_array)
// For Spanner: NotContains("my_array", 5) produces NOT ARRAY_INCLUDES(my_array, 5)
func NotContains(array sql.Querier, value any) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.Arg(value)
			b.WriteString(" <> ALL(")
			b.Join(array)
			b.WriteString(")")
		default:
			// For Spanner, use NOT ARRAY_INCLUDES
			b.WriteString("NOT ARRAY_INCLUDES(")
			b.Join(array)
			b.WriteString(", ")
			b.Arg(value)
			b.WriteString(")")
		}
	})
}

// IsEmpty checks if an array is empty.
// For PostgreSQL: IsEmpty("my_array") produces array_length(my_array, 1) = 0 OR my_array IS NULL
// For Spanner: IsEmpty("my_array") produces ARRAY_LENGTH(my_array) = 0
func IsEmpty(array sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			// PostgreSQL arrays can be NULL, and array_length returns NULL for empty arrays
			b.WriteString("(array_length(")
			b.Join(array)
			b.WriteString(", 1) = 0 OR ")
			b.Join(array)
			b.WriteString(" IS NULL)")
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") = 0")
		}
	})
}

// IsNotEmpty checks if an array is not empty.
// For PostgreSQL: IsNotEmpty("my_array") produces array_length(my_array, 1) > 0
// For Spanner: IsNotEmpty("my_array") produces ARRAY_LENGTH(my_array) > 0
func IsNotEmpty(array sql.Querier) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("array_length(")
			b.Join(array)
			b.WriteString(", 1) > 0")
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") > 0")
		}
	})
}

// LengthEQ checks if an array has a specific length.
// For PostgreSQL: LengthEQ("my_array", 3) produces array_length(my_array, 1) = 3
// For Spanner: LengthEQ("my_array", 3) produces ARRAY_LENGTH(my_array) = 3
func LengthEQ(array sql.Querier, length int) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("array_length(")
			b.Join(array)
			b.WriteString(", 1) = ")
			b.WriteString(strconv.Itoa(length))
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") = ")
			b.WriteString(strconv.Itoa(length))
		}
	})
}

// LengthGT checks if an array length is greater than a specific value.
// For PostgreSQL: LengthGT("my_array", 3) produces array_length(my_array, 1) > 3
// For Spanner: LengthGT("my_array", 3) produces ARRAY_LENGTH(my_array) > 3
func LengthGT(array sql.Querier, length int) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("array_length(")
			b.Join(array)
			b.WriteString(", 1) > ")
			b.WriteString(strconv.Itoa(length))
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") > ")
			b.WriteString(strconv.Itoa(length))
		}
	})
}

// LengthLT checks if an array length is less than a specific value.
// For PostgreSQL: LengthLT("my_array", 3) produces array_length(my_array, 1) < 3
// For Spanner: LengthLT("my_array", 3) produces ARRAY_LENGTH(my_array) < 3
func LengthLT(array sql.Querier, length int) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("array_length(")
			b.Join(array)
			b.WriteString(", 1) < ")
			b.WriteString(strconv.Itoa(length))
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") < ")
			b.WriteString(strconv.Itoa(length))
		}
	})
}

// LengthGTE checks if an array length is greater than or equal to a specific value.
// For PostgreSQL: LengthGTE("my_array", 3) produces array_length(my_array, 1) >= 3
// For Spanner: LengthGTE("my_array", 3) produces ARRAY_LENGTH(my_array) >= 3
func LengthGTE(array sql.Querier, length int) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("array_length(")
			b.Join(array)
			b.WriteString(", 1) >= ")
			b.WriteString(strconv.Itoa(length))
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") >= ")
			b.WriteString(strconv.Itoa(length))
		}
	})
}

// LengthLTE checks if an array length is less than or equal to a specific value.
// For PostgreSQL: LengthLTE("my_array", 3) produces array_length(my_array, 1) <= 3
// For Spanner: LengthLTE("my_array", 3) produces ARRAY_LENGTH(my_array) <= 3
func LengthLTE(array sql.Querier, length int) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			b.WriteString("array_length(")
			b.Join(array)
			b.WriteString(", 1) <= ")
			b.WriteString(strconv.Itoa(length))
		default:
			b.WriteString("ARRAY_LENGTH(")
			b.Join(array)
			b.WriteString(") <= ")
			b.WriteString(strconv.Itoa(length))
		}
	})
}

// IncludesAny checks if an array includes any of the given values (Spanner only).
// For PostgreSQL: uses overlap with constructed array
func IncludesAny(array sql.Querier, values ...any) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			// Use overlap operator with constructed array
			b.Join(array)
			b.WriteString(" && ARRAY[")
			for i, v := range values {
				if i > 0 {
					b.WriteString(",")
				}
				b.Arg(v)
			}
			b.WriteString("]")
		default:
			b.WriteString("ARRAY_INCLUDES_ANY(")
			b.Join(array)
			b.WriteString(", [")
			for i, v := range values {
				if i > 0 {
					b.WriteString(", ")
				}
				b.Arg(v)
			}
			b.WriteString("])")
		}
	})
}

// IncludesAll checks if an array includes all of the given values.
// For PostgreSQL: uses containment operator with constructed array
// For Spanner: uses ARRAY_INCLUDES_ALL
func IncludesAll(array sql.Querier, values ...any) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		checkArraySupport(b)
		switch b.Dialect() {
		case dialect.Postgres:
			// Use @> operator with constructed array
			b.Join(array)
			b.WriteString(" @> ARRAY[")
			for i, v := range values {
				if i > 0 {
					b.WriteString(",")
				}
				b.Arg(v)
			}
			b.WriteString("]")
		default:
			b.WriteString("ARRAY_INCLUDES_ALL(")
			b.Join(array)
			b.WriteString(", [")
			for i, v := range values {
				if i > 0 {
					b.WriteString(", ")
				}
				b.Arg(v)
			}
			b.WriteString("])")
		}
	})
}
