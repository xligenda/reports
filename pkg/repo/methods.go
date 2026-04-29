package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Find retrieves multiple records with filters and options.
func (r *GenericRepository[I, T]) Find(ctx context.Context, filters []Filter, opts *QueryOptions) ([]*T, error) {
	query, args := r.buildSelectQuery(filters, opts)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, newDatabaseError("query", err)
	}
	defer rows.Close()

	var result []*T
	for rows.Next() {
		entity, err := scanRowWithJSON[T](rows)
		if err != nil {
			return nil, newDatabaseError("scan row", err)
		}
		result = append(result, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, newDatabaseError("iterating rows", err)
	}

	return result, nil
}

// FindOne retrieves a single record, returning nil if not found.
func (r *GenericRepository[I, T]) FindOne(ctx context.Context, filters []Filter) (*T, error) {
	opts := &QueryOptions{Limit: 1}
	query, args := r.buildSelectQuery(filters, opts)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, newDatabaseError("query", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	entity, err := scanRowWithJSON[T](rows)
	if err != nil {
		return nil, newDatabaseError("scan row", err)
	}

	return entity, nil
}

// FindByID retrieves a record by ID, returning ErrNotFound if it does not exist.
func (r *GenericRepository[I, T]) FindByID(ctx context.Context, id I) (*T, error) {
	filters := []Filter{{Field: "id", Operator: "=", Value: id}}

	entity, err := r.FindOne(ctx, filters)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, newNotFoundError(fmt.Sprintf("id %v", id))
	}

	return entity, nil
}

// Create inserts a new record and returns the persisted entity.
func (r *GenericRepository[I, T]) Create(ctx context.Context, entity T) (*T, error) {
	fields, values, placeholders := r.buildInsertData(entity)

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	rows, err := r.db.QueryxContext(ctx, query, values...)
	if err != nil {
		return nil, newDatabaseError("create", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, newDatabaseError("create", ErrNoRows)
	}

	created, err := scanRowWithJSON[T](rows)
	if err != nil {
		return nil, newDatabaseError("scan row", err)
	}

	return created, nil
}

// Update updates a record by ID and returns the updated entity.
func (r *GenericRepository[I, T]) Update(ctx context.Context, id I, entity T) (*T, error) {
	fields, values := r.buildUpdateData(entity)
	if len(fields) == 0 {
		return nil, newValidationError("no fields to update")
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		strings.Join(fields, ", "),
		len(values)+1,
	)
	values = append(values, id)

	rows, err := r.db.QueryxContext(ctx, query, values...)
	if err != nil {
		return nil, newDatabaseError("update", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, newNotFoundError(fmt.Sprintf("id %v", id))
	}

	updated, err := scanRowWithJSON[T](rows)
	if err != nil {
		return nil, newDatabaseError("scan row", err)
	}

	return updated, nil
}

// Upsert inserts or updates a record based on the given conflict columns.
func (r *GenericRepository[I, T]) Upsert(ctx context.Context, entity T, conflictColumns []string) (*T, error) {
	fields, values, placeholders := r.buildInsertData(entity)

	if len(conflictColumns) == 0 {
		conflictColumns = []string{"id"}
	}

	var updateParts []string
	for _, field := range fields {
		isConflictCol := false
		for _, cc := range conflictColumns {
			if field == pq.QuoteIdentifier(cc) {
				isConflictCol = true
				break
			}
		}
		if !isConflictCol {
			updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", field, field))
		}
	}

	if len(updateParts) == 0 {
		return nil, newValidationError("all fields are conflict columns; nothing to update")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(conflictColumns, ", "),
		strings.Join(updateParts, ", "),
	)

	rows, err := r.db.QueryxContext(ctx, query, values...)
	if err != nil {
		return nil, newDatabaseError("upsert", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, newDatabaseError("upsert", ErrNoRows)
	}

	upserted, err := scanRowWithJSON[T](rows)
	if err != nil {
		return nil, newDatabaseError("scan row", err)
	}

	return upserted, nil
}

// Delete removes a record by ID, returning ErrNotFound if it does not exist.
func (r *GenericRepository[I, T]) Delete(ctx context.Context, id I) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", pq.QuoteIdentifier(r.tableName))

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return newDatabaseError("delete", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return newDatabaseError("delete rows affected", err)
	}
	if affected == 0 {
		return newNotFoundError(fmt.Sprintf("id %v", id))
	}

	return nil
}

// DeleteMany removes all records matching the given filters and returns the number deleted.
func (r *GenericRepository[I, T]) DeleteMany(ctx context.Context, filters []Filter) (int64, error) {
	whereClause, args := r.buildWhereClause(filters)
	query := fmt.Sprintf("DELETE FROM %s%s", pq.QuoteIdentifier(r.tableName), whereClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, newDatabaseError("delete many", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, newDatabaseError("delete many rows affected", err)
	}

	return affected, nil
}

// FindByIDWithTx retrieves a record by ID within a transaction.
func (r *GenericRepository[I, T]) FindByIDWithTx(ctx context.Context, tx *sqlx.Tx, id I) (*T, error) {
	filters := []Filter{{Field: "id", Operator: "=", Value: id}}
	opts := &QueryOptions{Limit: 1}
	query, args := r.buildSelectQuery(filters, opts)

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, newDatabaseError("query", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, newNotFoundError(fmt.Sprintf("id %v", id))
	}

	entity, err := scanRowWithJSON[T](rows)
	if err != nil {
		return nil, newDatabaseError("scan row", err)
	}

	return entity, nil
}

// UpdateWithTx updates a record by ID within a transaction.
func (r *GenericRepository[I, T]) UpdateWithTx(ctx context.Context, tx *sqlx.Tx, id I, entity T) (*T, error) {
	fields, values := r.buildUpdateData(entity)
	if len(fields) == 0 {
		return nil, newValidationError("no fields to update")
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		strings.Join(fields, ", "),
		len(values)+1,
	)
	values = append(values, id)

	rows, err := tx.QueryxContext(ctx, query, values...)
	if err != nil {
		return nil, newDatabaseError("update", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, newNotFoundError(fmt.Sprintf("id %v", id))
	}

	updated, err := scanRowWithJSON[T](rows)
	if err != nil {
		return nil, newDatabaseError("scan row", err)
	}

	return updated, nil
}

// Count returns the number of records matching the given filters.
func (r *GenericRepository[I, T]) Count(ctx context.Context, filters []Filter) (int64, error) {
	whereClause, args := r.buildWhereClause(filters)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", pq.QuoteIdentifier(r.tableName), whereClause)

	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, newDatabaseError("count", err)
	}

	return count, nil
}

// Exists reports whether any record matches the given filters.
func (r *GenericRepository[I, T]) Exists(ctx context.Context, filters []Filter) (bool, error) {
	count, err := r.Count(ctx, filters)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsWithID reports whether a record with the given ID exists.
func (r *GenericRepository[I, T]) ExistsWithID(ctx context.Context, id I) (bool, error) {
	filters := []Filter{{Field: "id", Operator: "=", Value: id}}
	return r.Exists(ctx, filters)
}

// FindWithPagination retrieves a page of records and the total count matching the filters.
func (r *GenericRepository[I, T]) FindWithPagination(ctx context.Context, filters []Filter, page, pageSize int, orderBy OrderBy) ([]*T, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	total, err := r.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	opts := &QueryOptions{
		OrderBy: orderBy,
		Limit:   pageSize,
		Offset:  (page - 1) * pageSize,
	}

	entities, err := r.Find(ctx, filters, opts)
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Transaction executes fn inside a database transaction.
// The transaction is automatically rolled back on error and committed on success.
func (r *GenericRepository[I, T]) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return newDatabaseError("begin transaction", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return newDatabaseError("rollback transaction", rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return newDatabaseError("commit transaction", err)
	}

	return nil
}
