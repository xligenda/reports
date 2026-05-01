package repo

type QueryOptions struct {
	OrderBy OrderBy
	Limit   int
	Offset  int
}

type Filter struct {
	Field string
	// =, !=, >, <, >=, <=, LIKE, IN, NOT IN, OR, IS NULL, IS NOT NULL
	Operator Operator
	Value    any
	// "field": "data->>'key'" for JSON fields
	JSONPath string
}

type Operator string

const (
	Equals        Operator = "="
	NotEquals     Operator = "!="
	GreaterThan   Operator = ">"
	LessThan      Operator = "<"
	GreaterEquals Operator = ">="
	LessEquals    Operator = "<="
	Like          Operator = "LIKE"
	In            Operator = "IN"
	NotIn         Operator = "NOT IN"
	Or            Operator = "OR"
	IsNull        Operator = "IS NULL"
	IsNotNull     Operator = "IS NOT NULL"
	Raw           Operator = "RAW"
)

func NewFilter(field string, operator Operator, value any) Filter {
	return Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

func NewORFilter(conditions ...any) Filter {
	var conditionMaps []map[string]any

	for _, cond := range conditions {
		switch v := cond.(type) {
		case Filter:
			condMap := map[string]any{
				"field":    v.Field,
				"operator": v.Operator,
				"value":    v.Value,
			}
			if v.JSONPath != "" {
				condMap["json_path"] = v.JSONPath
			}
			conditionMaps = append(conditionMaps, condMap)

		case map[string]any:
			conditionMaps = append(conditionMaps, v)
		}
	}

	return Filter{
		Field:    "custom_or",
		Operator: "OR",
		Value: map[string]any{
			"conditions": conditionMaps,
		},
	}
}

func NewRawFilter(sql string) Filter {
	return Filter{
		Field:    "raw_condition",
		Operator: "RAW",
		Value:    sql,
	}
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

type OrderByItem struct {
	Column    string
	Expr      string
	Args      []any
	Direction string
}

type OrderBy []OrderByItem

func (opts *QueryOptions) WithOrderBy(orderBy OrderBy) *QueryOptions {
	opts.OrderBy = orderBy
	return opts
}

func (opts *QueryOptions) WithLimit(limit int) *QueryOptions {
	opts.Limit = limit
	return opts
}

func (opts *QueryOptions) WithOffset(offset int) *QueryOptions {
	opts.Offset = offset
	return opts
}
