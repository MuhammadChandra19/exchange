package postgresql

import (
	"fmt"
	"strings"
)

// queryBuilder implements QueryBuilder interface
type queryBuilder struct {
	selectCols  []string
	fromTable   string
	joins       []string
	whereCond   []string
	whereArgs   []any
	groupByCols []string
	havingCond  []string
	havingArgs  []any
	orderByCols []string
	limitVal    *int
	offsetVal   *int
	argCounter  int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() QueryBuilder {
	return &queryBuilder{
		selectCols:  make([]string, 0),
		joins:       make([]string, 0),
		whereCond:   make([]string, 0),
		whereArgs:   make([]any, 0),
		groupByCols: make([]string, 0),
		havingCond:  make([]string, 0),
		havingArgs:  make([]any, 0),
		orderByCols: make([]string, 0),
		argCounter:  0,
	}
}

func (qb *queryBuilder) Select(columns ...string) QueryBuilder {
	qb.selectCols = append(qb.selectCols, columns...)
	return qb
}

func (qb *queryBuilder) From(table string) QueryBuilder {
	qb.fromTable = table
	return qb
}

func (qb *queryBuilder) Where(condition string, args ...any) QueryBuilder {
	// Replace ? placeholders with $1, $2, etc. (PostgreSQL style)
	for range args {
		qb.argCounter++
		condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", qb.argCounter), 1)
	}
	qb.whereCond = append(qb.whereCond, condition)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

func (qb *queryBuilder) Join(table, condition string) QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return qb
}

func (qb *queryBuilder) LeftJoin(table, condition string) QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

func (qb *queryBuilder) RightJoin(table, condition string) QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition))
	return qb
}

func (qb *queryBuilder) GroupBy(columns ...string) QueryBuilder {
	qb.groupByCols = append(qb.groupByCols, columns...)
	return qb
}

func (qb *queryBuilder) Having(condition string, args ...any) QueryBuilder {
	// Replace ? placeholders with $1, $2, etc.
	for range args {
		qb.argCounter++
		condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", qb.argCounter), 1)
	}
	qb.havingCond = append(qb.havingCond, condition)
	qb.havingArgs = append(qb.havingArgs, args...)
	return qb
}

func (qb *queryBuilder) OrderBy(column string, desc ...bool) QueryBuilder {
	order := "ASC"
	if len(desc) > 0 && desc[0] {
		order = "DESC"
	}
	qb.orderByCols = append(qb.orderByCols, fmt.Sprintf("%s %s", column, order))
	return qb
}

func (qb *queryBuilder) Limit(limit int) QueryBuilder {
	qb.limitVal = &limit
	return qb
}

func (qb *queryBuilder) Offset(offset int) QueryBuilder {
	qb.offsetVal = &offset
	return qb
}

func (qb *queryBuilder) Build() (string, []any) {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	if len(qb.selectCols) == 0 {
		query.WriteString("*")
	} else {
		query.WriteString(strings.Join(qb.selectCols, ", "))
	}

	// FROM clause
	if qb.fromTable != "" {
		query.WriteString(" FROM ")
		query.WriteString(qb.fromTable)
	}

	// JOIN clauses
	if len(qb.joins) > 0 {
		query.WriteString(" ")
		query.WriteString(strings.Join(qb.joins, " "))
	}

	// WHERE clause
	if len(qb.whereCond) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.whereCond, " AND "))
	}

	// GROUP BY clause
	if len(qb.groupByCols) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupByCols, ", "))
	}

	// HAVING clause
	if len(qb.havingCond) > 0 {
		query.WriteString(" HAVING ")
		query.WriteString(strings.Join(qb.havingCond, " AND "))
	}

	// ORDER BY clause
	if len(qb.orderByCols) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderByCols, ", "))
	}

	// LIMIT clause
	if qb.limitVal != nil {
		qb.argCounter++
		query.WriteString(fmt.Sprintf(" LIMIT $%d", qb.argCounter))
	}

	// OFFSET clause
	if qb.offsetVal != nil {
		qb.argCounter++
		query.WriteString(fmt.Sprintf(" OFFSET $%d", qb.argCounter))
	}

	// Combine all arguments
	allArgs := make([]any, 0, len(qb.whereArgs)+len(qb.havingArgs)+2)
	allArgs = append(allArgs, qb.whereArgs...)
	allArgs = append(allArgs, qb.havingArgs...)

	if qb.limitVal != nil {
		allArgs = append(allArgs, *qb.limitVal)
	}
	if qb.offsetVal != nil {
		allArgs = append(allArgs, *qb.offsetVal)
	}

	return query.String(), allArgs
}

func (qb *queryBuilder) Reset() QueryBuilder {
	return NewQueryBuilder()
}

// insertBuilder implements InsertBuilder interface
type insertBuilder struct {
	table      string
	columns    []string
	values     [][]any
	onConflict string
	returning  []string
	argCounter int
}

// NewInsertBuilder creates a new insert builder
func NewInsertBuilder() InsertBuilder {
	return &insertBuilder{
		columns:    make([]string, 0),
		values:     make([][]any, 0),
		returning:  make([]string, 0),
		argCounter: 0,
	}
}

func (ib *insertBuilder) Into(table string) InsertBuilder {
	ib.table = table
	return ib
}

func (ib *insertBuilder) Columns(columns ...string) InsertBuilder {
	ib.columns = columns
	return ib
}

func (ib *insertBuilder) Values(values ...any) InsertBuilder {
	ib.values = append(ib.values, values)
	return ib
}

func (ib *insertBuilder) ValuesMap(valueMap map[string]any) InsertBuilder {
	if len(ib.columns) == 0 {
		// Set columns from map keys
		for col := range valueMap {
			ib.columns = append(ib.columns, col)
		}
	}

	values := make([]any, len(ib.columns))
	for i, col := range ib.columns {
		values[i] = valueMap[col]
	}
	ib.values = append(ib.values, values)
	return ib
}

func (ib *insertBuilder) OnConflict(constraint string) InsertBuilder {
	ib.onConflict = fmt.Sprintf("ON CONFLICT (%s)", constraint)
	return ib
}

func (ib *insertBuilder) OnConflictDoNothing() InsertBuilder {
	ib.onConflict += " DO NOTHING"
	return ib
}

func (ib *insertBuilder) OnConflictDoUpdate(setClause string, args ...any) InsertBuilder {
	// Replace ? placeholders with $1, $2, etc.
	for range args {
		ib.argCounter++
		setClause = strings.Replace(setClause, "?", fmt.Sprintf("$%d", ib.argCounter), 1)
	}
	ib.onConflict += fmt.Sprintf(" DO UPDATE SET %s", setClause)
	return ib
}

func (ib *insertBuilder) Returning(columns ...string) InsertBuilder {
	ib.returning = columns
	return ib
}

func (ib *insertBuilder) Build() (string, []any) {
	var query strings.Builder
	var args []any

	query.WriteString("INSERT INTO ")
	query.WriteString(ib.table)

	if len(ib.columns) > 0 {
		query.WriteString(" (")
		query.WriteString(strings.Join(ib.columns, ", "))
		query.WriteString(")")
	}

	query.WriteString(" VALUES ")

	valuePlaceholders := make([]string, len(ib.values))
	argIndex := ib.argCounter

	for i, row := range ib.values {
		placeholders := make([]string, len(row))
		for j := range row {
			argIndex++
			placeholders[j] = fmt.Sprintf("$%d", argIndex)
			args = append(args, row[j])
		}
		valuePlaceholders[i] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	query.WriteString(strings.Join(valuePlaceholders, ", "))

	if ib.onConflict != "" {
		query.WriteString(" ")
		query.WriteString(ib.onConflict)
	}

	if len(ib.returning) > 0 {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(ib.returning, ", "))
	}

	return query.String(), args
}

func (ib *insertBuilder) Reset() InsertBuilder {
	return NewInsertBuilder()
}
