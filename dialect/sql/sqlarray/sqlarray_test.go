package sqlarray

import (
	"strings"
	"testing"

	"entgo.io/ent/dialect/sql"
)

func TestArray(t *testing.T) {
	b := &sql.Builder{}
	Array("SELECT 1, 2, 3")(b)
	if got, want := b.String(), "ARRAY(SELECT 1, 2, 3)"; got != want {
		t.Errorf("Array() = %q, want %q", got, want)
	}
}

func TestArrayOf(t *testing.T) {
	subquery := sql.Select("id").From(sql.Table("users"))
	result := ArrayOf(subquery)
	
	b := &sql.Builder{}
	b.Join(result)
	
	if got := b.String(); got != "ARRAY(SELECT `id` FROM `users`)" {
		t.Errorf("ArrayOf() = %q, want %q", got, "ARRAY(SELECT `id` FROM `users`)")
	}
}

func TestLiteral(t *testing.T) {
	tests := []struct {
		name   string
		values []any
		want   string
	}{
		{
			name:   "integers",
			values: []any{1, 2, 3},
			want:   "[?, ?, ?]",
		},
		{
			name:   "strings",
			values: []any{"a", "b", "c"},
			want:   "[?, ?, ?]",
		},
		{
			name:   "mixed types",
			values: []any{1, "hello", true},
			want:   "[?, ?, ?]",
		},
		{
			name:   "single value",
			values: []any{42},
			want:   "[?]",
		},
		{
			name:   "empty",
			values: []any{},
			want:   "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Literal(tt.values...)
			b := &sql.Builder{}
			b.Join(result)
			
			if got := b.String(); got != tt.want {
				t.Errorf("Literal(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

func TestArrayAgg(t *testing.T) {
	tests := []struct {
		name    string
		expr    sql.Querier
		orderBy []sql.Querier
		want    string
	}{
		{
			name: "without order by",
			expr: sql.Raw("column"),
			want: "ARRAY_AGG(column)",
		},
		{
			name:    "with single order by",
			expr:    sql.Raw("column"),
			orderBy: []sql.Querier{sql.Raw("column ASC")},
			want:    "ARRAY_AGG(column ORDER BY column ASC)",
		},
		{
			name:    "with multiple order by",
			expr:    sql.Raw("name"),
			orderBy: []sql.Querier{sql.Raw("created_at DESC"), sql.Raw("id ASC")},
			want:    "ARRAY_AGG(name ORDER BY created_at DESC, id ASC)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrayAgg(tt.expr, tt.orderBy...)
			b := &sql.Builder{}
			b.Join(result)
			
			if got := b.String(); got != tt.want {
				t.Errorf("ArrayAgg() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestElementAccess(t *testing.T) {
	array := sql.Raw("my_array")
	
	tests := []struct {
		name   string
		fn     func(sql.Querier, int) sql.Querier
		index  int
		want   string
	}{
		{
			name:  "Offset",
			fn:    Offset,
			index: 0,
			want:  "my_array[OFFSET(0)]",
		},
		{
			name:  "SafeOffset",
			fn:    SafeOffset,
			index: 2,
			want:  "my_array[SAFE_OFFSET(2)]",
		},
		{
			name:  "Ordinal",
			fn:    Ordinal,
			index: 1,
			want:  "my_array[ORDINAL(1)]",
		},
		{
			name:  "SafeOrdinal",
			fn:    SafeOrdinal,
			index: 3,
			want:  "my_array[SAFE_ORDINAL(3)]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(array, tt.index)
			b := &sql.Builder{}
			b.Join(result)
			
			if got := b.String(); got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestManipulationFunctions(t *testing.T) {
	array := sql.Raw("my_array")
	array2 := sql.Raw("other_array")
	
	tests := []struct {
		name string
		fn   func() sql.Querier
		want string
	}{
		{
			name: "Length",
			fn:   func() sql.Querier { return Length(array) },
			want: "ARRAY_LENGTH(my_array)",
		},
		{
			name: "Concat",
			fn:   func() sql.Querier { return Concat(array, array2) },
			want: "ARRAY_CONCAT(my_array, other_array)",
		},
		{
			name: "ToString",
			fn:   func() sql.Querier { return ToString(array, ",") },
			want: "ARRAY_TO_STRING(my_array, ?)",
		},
		{
			name: "ToStringWithNull",
			fn:   func() sql.Querier { return ToStringWithNull(array, ",", "NULL") },
			want: "ARRAY_TO_STRING(my_array, ?, ?)",
		},
		{
			name: "Reverse",
			fn:   func() sql.Querier { return Reverse(array) },
			want: "ARRAY_REVERSE(my_array)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			b := &sql.Builder{}
			b.Join(result)
			
			if got := b.String(); got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestPredicateFunctions(t *testing.T) {
	array := sql.Raw("my_array")
	
	tests := []struct {
		name string
		fn   func() *sql.Predicate
		want string
	}{
		{
			name: "Contains",
			fn:   func() *sql.Predicate { return Contains(array, 5) },
			want: "ARRAY_INCLUDES(my_array, ?)", // Default dialect (Spanner)
		},
		{
			name: "NotContains",
			fn:   func() *sql.Predicate { return NotContains(array, 5) },
			want: "NOT ARRAY_INCLUDES(my_array, ?)", // Default dialect (Spanner)
		},
		{
			name: "IsEmpty",
			fn:   func() *sql.Predicate { return IsEmpty(array) },
			want: "ARRAY_LENGTH(my_array) = 0",
		},
		{
			name: "IsNotEmpty",
			fn:   func() *sql.Predicate { return IsNotEmpty(array) },
			want: "ARRAY_LENGTH(my_array) > 0",
		},
		{
			name: "LengthEQ",
			fn:   func() *sql.Predicate { return LengthEQ(array, 3) },
			want: "ARRAY_LENGTH(my_array) = 3",
		},
		{
			name: "LengthGT",
			fn:   func() *sql.Predicate { return LengthGT(array, 2) },
			want: "ARRAY_LENGTH(my_array) > 2",
		},
		{
			name: "LengthLT",
			fn:   func() *sql.Predicate { return LengthLT(array, 5) },
			want: "ARRAY_LENGTH(my_array) < 5",
		},
		{
			name: "LengthGTE",
			fn:   func() *sql.Predicate { return LengthGTE(array, 1) },
			want: "ARRAY_LENGTH(my_array) >= 1",
		},
		{
			name: "LengthLTE",
			fn:   func() *sql.Predicate { return LengthLTE(array, 10) },
			want: "ARRAY_LENGTH(my_array) <= 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := tt.fn()
			b := &sql.Builder{}
			b.Join(predicate)
			
			if got := b.String(); got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestUnnestBuilder(t *testing.T) {
	array := sql.Raw("my_array")
	
	tests := []struct {
		name    string
		builder func() *UnnestBuilder
		want    string
	}{
		{
			name:    "basic unnest",
			builder: func() *UnnestBuilder { return Unnest(array) },
			want:    "UNNEST(my_array)",
		},
		{
			name:    "unnest with alias",
			builder: func() *UnnestBuilder { return Unnest(array).As("t") },
			want:    "UNNEST(my_array) AS t",
		},
		{
			name:    "unnest with offset",
			builder: func() *UnnestBuilder { return Unnest(array).WithOffset() },
			want:    "UNNEST(my_array) WITH OFFSET",
		},
		{
			name:    "unnest with custom offset alias",
			builder: func() *UnnestBuilder { return Unnest(array).WithOffsetAs("pos") },
			want:    "UNNEST(my_array) WITH OFFSET AS pos",
		},
		{
			name:    "unnest with alias and offset",
			builder: func() *UnnestBuilder { return Unnest(array).As("t").WithOffsetAs("pos") },
			want:    "UNNEST(my_array) AS t WITH OFFSET AS pos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builder()
			
			if got := builder.Query(); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
	
	// Test dialect-specific offset behavior
	t.Run("dialect differences", func(t *testing.T) {
		offsetTests := []struct {
			name        string
			builder     func() *UnnestBuilder
			wantSpanner string
			wantPg      string
		}{
			{
				name:        "with offset default name",
				builder:     func() *UnnestBuilder { return Unnest(array).WithOffset() },
				wantSpanner: "UNNEST(my_array) WITH OFFSET",
				wantPg:      "UNNEST(my_array) WITH ORDINALITY",
			},
			{
				name:        "with offset custom name",
				builder:     func() *UnnestBuilder { return Unnest(array).WithOffsetAs("pos") },
				wantSpanner: "UNNEST(my_array) WITH OFFSET AS pos",
				wantPg:      "UNNEST(my_array) WITH ORDINALITY AS pos",
			},
		}
		
		for _, tt := range offsetTests {
			t.Run(tt.name, func(t *testing.T) {
				// Test Spanner dialect
				builder := tt.builder()
				builder.SetDialect("spanner")
				if got := builder.Query(); got != tt.wantSpanner {
					t.Errorf("Spanner %s = %q, want %q", tt.name, got, tt.wantSpanner)
				}
				
				// Test PostgreSQL dialect
				builder = tt.builder()
				builder.SetDialect("postgres")
				if got := builder.Query(); got != tt.wantPg {
					t.Errorf("PostgreSQL %s = %q, want %q", tt.name, got, tt.wantPg)
				}
			})
		}
	})
}

func TestIntegrationExample(t *testing.T) {
	// Example: Find users whose hobbies include "reading"
	b := &sql.Builder{}
	b.WriteString("SELECT id, name FROM users WHERE ")
	b.Join(Contains(sql.Raw("hobbies"), "reading"))
	
	expectedSQL := "SELECT id, name FROM users WHERE ARRAY_INCLUDES(hobbies, ?)" // Default dialect (Spanner)
	if got := b.String(); got != expectedSQL {
		t.Errorf("Integration example = %q, want %q", got, expectedSQL)
	}
}

func TestComplexExample(t *testing.T) {
	// Example: Get arrays with length > 2 and concatenate them
	tagsArray := sql.Raw("tags")
	categoriesArray := sql.Raw("categories")
	
	concatenated := Concat(tagsArray, categoriesArray)
	lengthCheck := LengthGT(concatenated, 2)
	
	b := &sql.Builder{}
	b.Join(lengthCheck)
	
	// This should produce a complex query with array operations
	result := b.String()
	expectedContains := []string{"ARRAY_CONCAT", "ARRAY_LENGTH"}
	for _, expected := range expectedContains {
		if !strings.Contains(result, expected) {
			t.Errorf("Complex example should contain %q, got: %q", expected, result)
		}
	}
}

// PostgreSQL-specific tests

func TestPgDialectSpecificBehavior(t *testing.T) {
	tests := []struct {
		name        string
		dialect     string
		fn          func() sql.Querier
		wantSpanner string
		wantPg      string
	}{
		{
			name:        "Literal",
			fn:          func() sql.Querier { return Literal(1, 2, 3) },
			wantSpanner: "[?, ?, ?]",
			wantPg:      "'{$1,$2,$3}'",
		},
		{
			name:        "Index",
			fn:          func() sql.Querier { return Index(sql.Raw("arr"), 1) },
			wantSpanner: "arr[OFFSET(1)]",
			wantPg:      "arr[1]",
		},
		{
			name:        "Length",
			fn:          func() sql.Querier { return Length(sql.Raw("arr")) },
			wantSpanner: "ARRAY_LENGTH(arr)",
			wantPg:      "ARRAY_LENGTH(arr, 1)",
		},
		{
			name:        "Concat",
			fn:          func() sql.Querier { return Concat(sql.Raw("arr1"), sql.Raw("arr2")) },
			wantSpanner: "ARRAY_CONCAT(arr1, arr2)",
			wantPg:      "arr1 || arr2",
		},
		{
			name:        "ToString",
			fn:          func() sql.Querier { return ToString(sql.Raw("arr"), ",") },
			wantSpanner: "ARRAY_TO_STRING(arr, ?)",
			wantPg:      "ARRAY_TO_STRING(arr, $1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Spanner dialect
			b := &sql.Builder{}
			b.SetDialect("spanner")
			b.Join(tt.fn())
			if got := b.String(); got != tt.wantSpanner {
				t.Errorf("Spanner %s = %q, want %q", tt.name, got, tt.wantSpanner)
			}

			// Test PostgreSQL dialect
			b = &sql.Builder{}
			b.SetDialect("postgres")
			b.Join(tt.fn())
			if got := b.String(); got != tt.wantPg {
				t.Errorf("PostgreSQL %s = %q, want %q", tt.name, got, tt.wantPg)
			}
		})
	}
}

func TestPgPredicateDialectBehavior(t *testing.T) {
	tests := []struct {
		name        string
		fn          func() *sql.Predicate
		wantSpanner string
		wantPg      string
	}{
		{
			name:        "Contains",
			fn:          func() *sql.Predicate { return Contains(sql.Raw("arr"), 5) },
			wantSpanner: "ARRAY_INCLUDES(arr, ?)",
			wantPg:      "$1 = ANY(arr)",
		},
		{
			name:        "NotContains",
			fn:          func() *sql.Predicate { return NotContains(sql.Raw("arr"), 5) },
			wantSpanner: "NOT ARRAY_INCLUDES(arr, ?)",
			wantPg:      "$1 <> ALL(arr)",
		},
		{
			name:        "IsEmpty",
			fn:          func() *sql.Predicate { return IsEmpty(sql.Raw("arr")) },
			wantSpanner: "ARRAY_LENGTH(arr) = 0",
			wantPg:      "(array_length(arr, 1) = 0 OR arr IS NULL)",
		},
		{
			name:        "IsNotEmpty",
			fn:          func() *sql.Predicate { return IsNotEmpty(sql.Raw("arr")) },
			wantSpanner: "ARRAY_LENGTH(arr) > 0",
			wantPg:      "array_length(arr, 1) > 0",
		},
		{
			name:        "LengthEQ",
			fn:          func() *sql.Predicate { return LengthEQ(sql.Raw("arr"), 3) },
			wantSpanner: "ARRAY_LENGTH(arr) = 3",
			wantPg:      "array_length(arr, 1) = 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Spanner dialect
			b := &sql.Builder{}
			b.SetDialect("spanner")
			b.Join(tt.fn())
			if got := b.String(); got != tt.wantSpanner {
				t.Errorf("Spanner %s = %q, want %q", tt.name, got, tt.wantSpanner)
			}

			// Test PostgreSQL dialect
			b = &sql.Builder{}
			b.SetDialect("postgres")
			b.Join(tt.fn())
			if got := b.String(); got != tt.wantPg {
				t.Errorf("PostgreSQL %s = %q, want %q", tt.name, got, tt.wantPg)
			}
		})
	}
}

func TestArrayConstruction(t *testing.T) {
	tests := []struct {
		name   string
		fn     func() sql.Querier
		want   string
	}{
		{
			name: "ArrayConstructor",
			fn:   func() sql.Querier { return ArrayConstructor(1, 2, 3) },
			want: "[?, ?, ?]", // Default dialect (Spanner)
		},
		{
			name: "ArrayOf",
			fn:   func() sql.Querier { return ArrayOf(sql.Select("id").From(sql.Table("users"))) },
			want: "ARRAY(SELECT `id` FROM `users`)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &sql.Builder{}
			b.Join(tt.fn())
			if got := b.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestArraySlicing(t *testing.T) {
	array := sql.Raw("my_array")
	
	tests := []struct {
		name string
		fn   func() sql.Querier
		want string
	}{
		{
			name: "Slice",
			fn:   func() sql.Querier { return Slice(array, 2, 4) },
			want: "ARRAY_SLICE(my_array, 2, 4)", // Default dialect (Spanner)
		},
		{
			name: "First",
			fn:   func() sql.Querier { return First(array) },
			want: "ARRAY_FIRST(my_array)", // Default dialect (Spanner)
		},
		{
			name: "Last",
			fn:   func() sql.Querier { return Last(array) },
			want: "ARRAY_LAST(my_array)", // Default dialect (Spanner)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &sql.Builder{}
			b.Join(tt.fn())
			if got := b.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestArrayOperators(t *testing.T) {
	array1 := sql.Raw("array1")
	array2 := sql.Raw("array2")
	
	tests := []struct {
		name string
		fn   func() *sql.Predicate
		want string
	}{
		{
			name: "Overlap",
			fn:   func() *sql.Predicate { return Overlap(array1, array2) },
			want: "ARRAY_INCLUDES_ANY(array1, array2)", // Default dialect (Spanner)
		},
		{
			name: "ContainsArray",
			fn:   func() *sql.Predicate { return ContainsArray(array1, array2) },
			want: "ARRAY_INCLUDES_ALL(array1, array2)", // Default dialect (Spanner)
		},
		{
			name: "ContainedBy",
			fn:   func() *sql.Predicate { return ContainedBy(array1, array2) },
			want: "ARRAY_INCLUDES_ALL(array2, array1)", // Default dialect (Spanner)
		},
		{
			name: "ArrayEqual",
			fn:   func() *sql.Predicate { return ArrayEqual(array1, array2) },
			want: "array1 = array2",
		},
		{
			name: "ArrayNotEqual",
			fn:   func() *sql.Predicate { return ArrayNotEqual(array1, array2) },
			want: "array1 <> array2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &sql.Builder{}
			b.Join(tt.fn())
			if got := b.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestArrayFunctions(t *testing.T) {
	array := sql.Raw("my_array")
	
	tests := []struct {
		name string
		fn   func() sql.Querier
		want string
	}{
		{
			name: "Position",
			fn:   func() sql.Querier { return Position(array, "value") },
			want: "(SELECT offset FROM UNNEST(my_array) AS elem WITH OFFSET WHERE elem = ? LIMIT 1)", // Default dialect (Spanner)
		},
		{
			name: "Prepend",
			fn:   func() sql.Querier { return Prepend("value", array) },
			want: "ARRAY_CONCAT([?], my_array)", // Default dialect (Spanner)
		},
		{
			name: "Append",
			fn:   func() sql.Querier { return Append(array, "value") },
			want: "ARRAY_CONCAT(my_array, [?])", // Default dialect (Spanner)
		},
		{
			name: "Remove",
			fn:   func() sql.Querier { return Remove(array, "value") },
			want: "ARRAY_FILTER(my_array, elem -> elem != ?)", // Default dialect (Spanner)
		},
		{
			name: "StringToArray",
			fn:   func() sql.Querier { return StringToArray("a,b,c", ",") },
			want: "SPLIT(?, ?)", // Default dialect (Spanner)
		},
		{
			name: "Transform",
			fn:   func() sql.Querier { return Transform(array, "elem", "elem * 2") },
			want: "ARRAY_TRANSFORM(my_array, elem -> elem * 2)", // Default dialect (Spanner)
		},
		{
			name: "Filter",
			fn:   func() sql.Querier { return Filter(array, "elem", "elem > 5") },
			want: "ARRAY_FILTER(my_array, elem -> elem > 5)", // Default dialect (Spanner)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &sql.Builder{}
			b.Join(tt.fn())
			if got := b.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestArrayIntegrationExample(t *testing.T) {
	// Example: Cross-dialect query using array operators
	b := &sql.Builder{}
	b.SetDialect("postgres")
	
	// Find users whose hobbies overlap with a given set
	givenHobbies := ArrayConstructor("reading", "swimming")
	overlap := Overlap(sql.Raw("hobbies"), givenHobbies)
	
	b.WriteString("SELECT id, name FROM users WHERE ")
	b.Join(overlap)
	
	expected := "SELECT id, name FROM users WHERE hobbies && ARRAY[$1,$2]"
	if got := b.String(); got != expected {
		t.Errorf("PostgreSQL integration example = %q, want %q", got, expected)
	}
}

func TestNewArrayPredicates(t *testing.T) {
	array := sql.Raw("my_array")
	
	tests := []struct {
		name string
		fn   func() *sql.Predicate
		want string
	}{
		{
			name: "IncludesAny",
			fn:   func() *sql.Predicate { return IncludesAny(array, 1, 2, 3) },
			want: "my_array && ARRAY[$1,$2,$3]",
		},
		{
			name: "IncludesAll",
			fn:   func() *sql.Predicate { return IncludesAll(array, 1, 2) },
			want: "my_array @> ARRAY[$1,$2]",
		},
		{
			name: "IsDistinct",
			fn:   func() *sql.Predicate { return IsDistinct(array) },
			want: "my_array = (SELECT array_agg(DISTINCT elem) FROM unnest(my_array) AS elem)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &sql.Builder{}
			b.SetDialect("postgres")
			b.Join(tt.fn())
			if got := b.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

