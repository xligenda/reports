package repo

import (
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// MockEntity for testing
type MockEntity struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (m MockEntity) GetID() int {
	return m.ID
}

// TestNewRepository tests the repository constructor
func TestNewRepository(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
	assert.Equal(t, "test_table", repo.tableName)
}

// TestNewFilter tests filter creation with various scenarios
func TestNewFilter(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		operator Operator
		value    any
	}{
		{"Equality", "name", Equals, "John"},
		{"Greater than", "age", ">", 18},
		{"LIKE", "email", "LIKE", "%@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.field, tt.operator, tt.value)
			assert.Equal(t, tt.field, filter.Field)
			assert.Equal(t, tt.operator, filter.Operator)
			assert.Equal(t, tt.value, filter.Value)
		})
	}
}

// TestNewORFilter tests OR filter creation
func TestNewORFilter(t *testing.T) {
	t.Run("With Filter structs", func(t *testing.T) {
		f1 := NewFilter("status", Equals, "active")
		f2 := NewFilter("priority", Equals, "high")
		orFilter := NewORFilter(f1, f2)

		assert.Equal(t, "custom_or", orFilter.Field)
		assert.Equal(t, "OR", orFilter.Operator)
		assert.NotNil(t, orFilter.Value)
	})

	t.Run("With map conditions", func(t *testing.T) {
		cond := map[string]any{"field": "status", "operator": Equals, "value": "active"}
		orFilter := NewORFilter(cond)

		assert.Equal(t, "custom_or", orFilter.Field)
		assert.Equal(t, "OR", orFilter.Operator)
	})

	t.Run("Empty OR filter", func(t *testing.T) {
		orFilter := NewORFilter()
		assert.Equal(t, "OR", orFilter.Operator)
	})
}

// TestNewRawFilter tests raw SQL filter creation
func TestNewRawFilter(t *testing.T) {
	sql := "age > 18 AND status = 'active'"
	filter := NewRawFilter(sql)

	assert.Equal(t, "raw_condition", filter.Field)
	assert.Equal(t, "RAW", filter.Operator)
	assert.Equal(t, sql, filter.Value)
}

// TestQueryOptions tests QueryOptions builder
func TestQueryOptions(t *testing.T) {
	opts := NewQueryOptions().
		WithLimit(10).
		WithOffset(5).
		WithOrderBy(OrderBy{{Column: "id", Direction: "ASC"}})

	assert.Equal(t, 10, opts.Limit)
	assert.Equal(t, 5, opts.Offset)
	assert.Len(t, opts.OrderBy, 1)
}

// TestJSONBFilter tests JSONB filter helper
func TestJSONBFilter(t *testing.T) {
	filter := JSONBFilter("metadata", "status", Equals, "active")

	assert.Equal(t, "metadata", filter.Field)
	assert.Equal(t, "metadata->>'status'", filter.JSONPath)
	assert.Equal(t, Equals, filter.Operator)
	assert.Equal(t, "active", filter.Value)
}

// TestJSONBExistsFilter tests JSONB exists filter
func TestJSONBExistsFilter(t *testing.T) {
	filter := JSONBExistsFilter("data", "key_name")

	assert.Equal(t, "data", filter.Field)
	assert.Equal(t, "RAW", filter.Operator)
	assert.Contains(t, filter.Value.(string), "?")
}

// TestJSONBContainsFilter tests JSONB contains filter
func TestJSONBContainsFilter(t *testing.T) {
	filter := JSONBContainsFilter("metadata", `{"type": "premium"}`)

	assert.Equal(t, "metadata", filter.Field)
	assert.Equal(t, "RAW", filter.Operator)
	assert.Contains(t, filter.JSONPath, "@>")
}

// TestBuildCondition tests condition building for different operators
func TestBuildCondition(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	tests := []struct {
		name   string
		filter Filter
		expect string
	}{
		{
			name:   "Equality",
			filter: Filter{Field: "name", Operator: Equals, Value: "John"},
			expect: `"name" = $1`,
		},
		{
			name:   "IS NULL",
			filter: Filter{Field: "deleted_at", Operator: IsNull},
			expect: `"deleted_at" IS NULL`,
		},
		{
			name:   "LIKE",
			filter: Filter{Field: "name", Operator: Like, Value: "%test%"},
			expect: `"name" LIKE $1`,
		},
		{
			name:   "IN operator",
			filter: Filter{Field: "status", Operator: In, Value: []any{"a", "b"}},
			expect: `"status" IN ($1, $2)`,
		},
		{
			name:   "NOT IN operator",
			filter: Filter{Field: "status", Operator: NotIn, Value: []any{"x"}},
			expect: `"status" NOT IN ($1)`,
		},
		{
			name:   "RAW operator",
			filter: Filter{Operator: Raw, Value: "custom sql"},
			expect: "custom sql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, _ := repo.buildCondition(tt.filter, 1)
			assert.Contains(t, cond, strings.Split(tt.expect, " ")[0])
		})
	}
}

// TestBuildWhereClause tests WHERE clause building
func TestBuildWhereClause(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	t.Run("Empty filters", func(t *testing.T) {
		clause, args := repo.buildWhereClause([]Filter{})
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("Single filter", func(t *testing.T) {
		filters := []Filter{{Field: "name", Operator: Equals, Value: "John"}}
		clause, args := repo.buildWhereClause(filters)
		assert.Contains(t, clause, " WHERE ")
		assert.Equal(t, []any{"John"}, args)
	})

	t.Run("Multiple filters AND", func(t *testing.T) {
		filters := []Filter{
			{Field: "name", Operator: Equals, Value: "John"},
			{Field: "age", Operator: ">", Value: 18},
		}
		clause, args := repo.buildWhereClause(filters)
		assert.Contains(t, clause, " AND ")
		assert.Equal(t, 2, len(args))
	})

	t.Run("With IN operator", func(t *testing.T) {
		filters := []Filter{{Field: "status", Operator: "IN", Value: []any{"a", "b", "c"}}}
		_, args := repo.buildWhereClause(filters)
		assert.Equal(t, 3, len(args))
	})
}

// TestBuildSelectQuery tests SELECT query building
func TestBuildSelectQuery(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	t.Run("Basic select", func(t *testing.T) {
		query, _ := repo.buildSelectQuery([]Filter{}, nil)
		assert.Equal(t, `SELECT * FROM "test_table"`, query)
	})

	t.Run("With filters", func(t *testing.T) {
		filters := []Filter{{Field: "name", Operator: Equals, Value: "John"}}
		query, _ := repo.buildSelectQuery(filters, nil)
		assert.Contains(t, query, "WHERE")
	})

	t.Run("With LIMIT", func(t *testing.T) {
		opts := &QueryOptions{Limit: 10}
		query, _ := repo.buildSelectQuery([]Filter{}, opts)
		assert.Contains(t, query, "LIMIT 10")
	})

	t.Run("With OFFSET", func(t *testing.T) {
		opts := &QueryOptions{Offset: 5}
		query, _ := repo.buildSelectQuery([]Filter{}, opts)
		assert.Contains(t, query, "OFFSET 5")
	})

	t.Run("With ORDER BY", func(t *testing.T) {
		opts := &QueryOptions{
			OrderBy: OrderBy{{Column: "id", Direction: "DESC"}},
		}
		query, _ := repo.buildSelectQuery([]Filter{}, opts)
		assert.Contains(t, query, `"id" DESC`)
	})

	t.Run("Full query", func(t *testing.T) {
		filters := []Filter{{Field: "status", Operator: Equals, Value: "active"}}
		opts := &QueryOptions{
			OrderBy: OrderBy{{Column: "id", Direction: "DESC"}},
			Limit:   20,
			Offset:  10,
		}
		query, _ := repo.buildSelectQuery(filters, opts)
		assert.Contains(t, query, "WHERE")
		assert.Contains(t, query, "ORDER BY")
		assert.Contains(t, query, "LIMIT 20")
		assert.Contains(t, query, "OFFSET 10")
	})
}

// TestBuildInsertData tests insert data building
func TestBuildInsertData(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	entity := MockEntity{ID: 1, Name: "John", Age: 25}
	fields, values, placeholders := repo.buildInsertData(entity)

	assert.NotEmpty(t, fields)
	assert.NotEmpty(t, values)
	assert.NotEmpty(t, placeholders)
	assert.Equal(t, len(fields), len(values))
	assert.Equal(t, len(values), len(placeholders))
}

// TestBuildUpdateData tests update data building
func TestBuildUpdateData(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	entity := MockEntity{ID: 1, Name: "Updated", Age: 26}
	fields, values := repo.buildUpdateData(entity)

	assert.NotEmpty(t, fields)
	assert.NotEmpty(t, values)
	// ID should be excluded from update
	for _, field := range fields {
		assert.NotContains(t, field, "id")
	}
}

// TestFieldEscaping tests that field names are properly escaped
func TestFieldEscaping(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")

	filter := Filter{Field: "user_name", Operator: Equals, Value: "test"}
	cond, _ := repo.buildCondition(filter, 1)

	assert.Contains(t, cond, `"user_name"`)
}

// TestSpecialTableNames tests with special table names
func TestSpecialTableNames(t *testing.T) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "user_accounts")

	query, _ := repo.buildSelectQuery([]Filter{}, nil)
	assert.Contains(t, query, `"user_accounts"`)
}

// Benchmarks

func BenchmarkBuildSelectQuery(b *testing.B) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")
	filters := []Filter{{Field: "name", Operator: Equals, Value: "test"}}
	opts := &QueryOptions{Limit: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.buildSelectQuery(filters, opts)
	}
}

func BenchmarkBuildWhereClause(b *testing.B) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")
	filters := []Filter{
		{Field: "name", Operator: Equals, Value: "John"},
		{Field: "age", Operator: ">", Value: 18},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.buildWhereClause(filters)
	}
}

func BenchmarkBuildInsertData(b *testing.B) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")
	entity := MockEntity{ID: 1, Name: "John", Age: 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.buildInsertData(entity)
	}
}

func BenchmarkBuildUpdateData(b *testing.B) {
	db := &sqlx.DB{}
	repo := NewRepository[int, MockEntity](db, "test_table")
	entity := MockEntity{ID: 1, Name: "John", Age: 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.buildUpdateData(entity)
	}
}
