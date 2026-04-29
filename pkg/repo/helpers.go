package repo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func (r *GenericRepository[I, T]) buildSelectQuery(filters []Filter, opts *QueryOptions) (string, []any) {
	query := fmt.Sprintf("SELECT * FROM %s", pq.QuoteIdentifier(r.tableName))

	whereClause, args := r.buildWhereClause(filters)
	query += whereClause

	if opts != nil {
		if len(opts.OrderBy) > 0 {
			var orderParts []string
			argIndex := len(args) + 1
			for _, ob := range opts.OrderBy {
				if ob.Expr != "" {
					var b strings.Builder
					for _, ch := range ob.Expr {
						if ch == '?' {
							fmt.Fprintf(&b, "$%d", argIndex)
							argIndex++
						} else {
							b.WriteRune(ch)
						}
					}
					part := b.String()
					if ob.Direction != "" {
						part = part + " " + ob.Direction
					}
					orderParts = append(orderParts, part)
					if len(ob.Args) > 0 {
						args = append(args, ob.Args...)
					}
				} else if ob.Column != "" {
					part := pq.QuoteIdentifier(ob.Column)
					if ob.Direction != "" {
						part = part + " " + ob.Direction
					}
					orderParts = append(orderParts, part)
				}
			}
			if len(orderParts) > 0 {
				query += " ORDER BY " + strings.Join(orderParts, ", ")
			}
		}
		if opts.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		}
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	return query, args
}

func (r *GenericRepository[I, T]) buildWhereClause(filters []Filter) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []any
	argIndex := 1

	for _, filter := range filters {
		condition, arg := r.buildCondition(filter, argIndex)
		conditions = append(conditions, condition)

		if arg != nil {
			switch v := arg.(type) {
			case []any:
				// multiple
				args = append(args, v...)
				argIndex += len(v)
			default:
				// single
				args = append(args, v)
				argIndex++
			}
		}
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")
	return whereClause, args
}
func (r *GenericRepository[I, T]) buildCondition(filter Filter, argIndex int) (string, any) {
	field := pq.QuoteIdentifier(filter.Field)
	if filter.JSONPath != "" {
		field = filter.JSONPath
	}

	switch strings.ToUpper(filter.Operator) {
	case "OR":
		if orData, ok := filter.Value.(map[string]any); ok {
			if conditions, exists := orData["conditions"].([]map[string]any); exists {
				var orParts []string
				var orArgs []any
				currentArgIndex := argIndex

				for _, condition := range conditions {
					fieldName, _ := condition["field"].(string)
					operator, _ := condition["operator"].(string)
					value := condition["value"]
					jsonPath, _ := condition["json_path"].(string)

					var condField string
					if jsonPath != "" {
						condField = jsonPath
					} else {
						condField = pq.QuoteIdentifier(fieldName)
					}

					orParts = append(orParts, fmt.Sprintf("%s %s $%d", condField, operator, currentArgIndex))
					orArgs = append(orArgs, value)
					currentArgIndex++
				}

				condition := fmt.Sprintf("(%s)", strings.Join(orParts, " OR "))
				return condition, orArgs
			}
		}

	case "RAW":
		if rawSQL, ok := filter.Value.(string); ok {
			return rawSQL, nil
		}
		if filter.Value == nil && filter.JSONPath != "" {
			return filter.JSONPath, nil
		}

	case "IN":
		if values, ok := filter.Value.([]any); ok {
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			}
			condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
			return condition, values
		}

	case "NOT IN":
		if values, ok := filter.Value.([]any); ok {
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			}
			condition := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
			return condition, values
		}

	case "IS NULL":
		return fmt.Sprintf("%s IS NULL", field), nil

	case "IS NOT NULL":
		return fmt.Sprintf("%s IS NOT NULL", field), nil

	case "LIKE", "ILIKE":
		condition := fmt.Sprintf("%s %s $%d", field, filter.Operator, argIndex)
		return condition, filter.Value

	default:
		condition := fmt.Sprintf("%s %s $%d", field, filter.Operator, argIndex)
		return condition, filter.Value
	}

	condition := fmt.Sprintf("%s = $%d", field, argIndex)
	return condition, filter.Value
}

func (r *GenericRepository[I, T]) buildInsertData(entity T) ([]string, []any, []string) {
	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	var fields []string
	var values []any
	var placeholders []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		// get field name and options from db tag
		rawTag := field.Tag.Get("db")
		if rawTag == "" || rawTag == "-" {
			continue
		}

		var hasOmitempty bool
		rawInsertTag := field.Tag.Get("insert")
		if rawInsertTag == "omitempty" {
			hasOmitempty = true
		}

		fieldValue := r.getFieldValue(value)

		if value.Kind() == reflect.Pointer && value.IsNil() {
			continue
		}

		if hasOmitempty {
			if value.IsZero() {
				continue
			}
		}

		fields = append(fields, pq.QuoteIdentifier(rawTag))
		values = append(values, fieldValue)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)))
	}

	return fields, values, placeholders
}

func (r *GenericRepository[I, T]) buildUpdateData(entity T) ([]string, []any) {
	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	// Handle pointer types
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	var fields []string
	var values []any

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		rawTag := field.Tag.Get("db")
		if rawTag == "" || rawTag == "-" {
			continue
		}

		tagParts := strings.Split(rawTag, ",")
		fieldName := strings.TrimSpace(tagParts[0])
		var hasOmitempty bool
		for _, opt := range tagParts[1:] {
			if strings.TrimSpace(opt) == "omitempty" {
				hasOmitempty = true
				break
			}
		}

		if fieldName == "" || fieldName == "id" {
			continue
		}

		fieldValue := r.getFieldValue(value)

		if value.Kind() == reflect.Pointer && value.IsNil() {
			continue
		}

		if hasOmitempty {
			if value.IsZero() {
				continue
			}
		}

		fields = append(fields, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(fieldName), len(values)+1))
		values = append(values, fieldValue)
	}

	return fields, values
}

func (r *GenericRepository[I, T]) getFieldValue(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return r.getFieldValue(value.Elem())
	}

	switch value.Type() {
	case reflect.TypeFor[time.Time]():
		return value.Interface().(time.Time)
	case reflect.TypeFor[*time.Time]():
		if value.IsZero() {
			return nil
		}
		return value.Interface().(time.Time)
	}

	switch value.Kind() {
	case reflect.Slice, reflect.Map, reflect.Struct:
		if value.Type() == reflect.TypeFor[time.Time]() {
			return value.Interface()
		}

		jsonData, err := json.Marshal(value.Interface())
		if err != nil {
			return nil
		}
		return jsonData
	}

	return value.Interface()
}

// scanRowWithJSON scans a row and unmarshals JSONB fields automatically
func scanRowWithJSON[T any](rows *sqlx.Rows) (*T, error) {
	values := make(map[string]any)
	if err := rows.MapScan(values); err != nil {
		return nil, fmt.Errorf("failed to map scan: %w", err)
	}

	var entity T
	entityValue := reflect.ValueOf(&entity).Elem()
	entityType := entityValue.Type()

	for i := 0; i < entityValue.NumField(); i++ {
		field := entityType.Field(i)
		fieldValue := entityValue.Field(i)

		if !field.IsExported() {
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		value, exists := values[dbTag]
		if !exists || value == nil {
			continue
		}

		// Handle JSONB fields ([]byte from postgres)
		if bytes, ok := value.([]byte); ok {
			fieldType := fieldValue.Type()
			fieldKind := fieldValue.Kind()

			// Check if this is a complex type that needs JSON unmarshaling
			// (struct, slice, array, map, or pointer to these)
			needsJSONUnmarshal := fieldKind == reflect.Struct ||
				fieldKind == reflect.Slice ||
				fieldKind == reflect.Array ||
				fieldKind == reflect.Map ||
				fieldKind == reflect.Pointer

			// Exclude time.Time from JSON unmarshaling
			if needsJSONUnmarshal && fieldType.String() != "time.Time" {
				if err := json.Unmarshal(bytes, fieldValue.Addr().Interface()); err != nil {
					return nil, fmt.Errorf("failed to unmarshal JSON for field %s: %w", field.Name, err)
				}
				continue
			}
		}

		// Handle regular fields
		if err := setFieldValue(fieldValue, value); err != nil {
			return nil, fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return &entity, nil
}

// setFieldValue sets a reflect.Value with proper type conversion
func setFieldValue(field reflect.Value, value any) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)

	// pointer fields
	if field.Kind() == reflect.Pointer {
		if val.IsValid() && !val.IsZero() {
			newPtr := reflect.New(field.Type().Elem())
			if val.Type().AssignableTo(field.Type().Elem()) {
				newPtr.Elem().Set(val)
			} else if val.Type().ConvertibleTo(field.Type().Elem()) {
				newPtr.Elem().Set(val.Convert(field.Type().Elem()))
			} else {
				return fmt.Errorf("cannot convert %v to %v", val.Type(), field.Type().Elem())
			}
			field.Set(newPtr)
		}
		return nil
	}

	// direct assignment
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
	} else if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
	} else {
		return fmt.Errorf("cannot convert %v to %v", val.Type(), field.Type())
	}

	return nil
}

func JSONBFilter(field string, key string, operator string, value any) Filter {
	return Filter{
		Field:    field,
		JSONPath: fmt.Sprintf("%s->>'%s'", field, key),
		Operator: operator,
		Value:    value,
	}
}

func JSONBExistsFilter(field string, key string) Filter {
	return Filter{
		Field:    field,
		Operator: "RAW",
		Value:    fmt.Sprintf("%s ? '%s'", field, key),
	}
}

func JSONBContainsFilter(field string, jsonValue string) Filter {
	return Filter{
		Field:    field,
		JSONPath: fmt.Sprintf("%s @> '%s'::jsonb", field, jsonValue),
		Operator: "RAW",
		Value:    nil,
	}
}
