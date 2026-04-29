package repo

import (
	"github.com/jmoiron/sqlx"
)

type IDsConstraint interface {
	string | int
}

type StructsConstraint[I IDsConstraint] interface {
	GetID() I
}

type GenericRepository[I IDsConstraint, T StructsConstraint[I]] struct {
	db        *sqlx.DB
	tableName string
}

func NewRepository[I IDsConstraint, T StructsConstraint[I]](db *sqlx.DB, tableName string) *GenericRepository[I, T] {
	return &GenericRepository[I, T]{
		db:        db,
		tableName: tableName,
	}
}
