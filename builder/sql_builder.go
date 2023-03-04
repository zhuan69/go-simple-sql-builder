package builder

import (
	"context"
	"database/sql"
	"fmt"
	"simple-sql-builder/constants"
	"simple-sql-builder/observable"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlBuilder[RT any, RWT any] struct {
	ctx             context.Context
	SqlObserve      []observable.SqlObserve
	args            []any
	count           int
	called          int
	tableName       string
	typeSql         observable.SqlType
	connection      any
	execRowCommand  func() RT
	execRowsCommand func() (res RWT, err error)
}

func (sb *sqlBuilder[RT, RWT]) RowQuery() RT {
	return sb.execRowCommand()
}

func (sb *sqlBuilder[RT, RWT]) RowsQuery() (res RWT, err error) {
	return sb.execRowsCommand()
}

// Insert implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) Insert(colAndVal map[string]any) SqlBuilder[RT, RWT] {
	return sb.insert(colAndVal)
}

// Join implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) Join(joinType string, target string, conditional string) SqlBuilder[RT, RWT] {
	return sb.join(joinType, target, conditional)
}

// OrWhere implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) OrWhere(column string, value any) SqlBuilder[RT, RWT] {
	return sb.orWhere(column, value)
}

// OrderBy implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) OrderBy(column string, sort string) SqlBuilder[RT, RWT] {
	return sb.orderBy(column, sort)
}

// Select implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) Select(col []string) SqlBuilder[RT, RWT] {
	return sb.selectSql(col)
}

// ToQueryString implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) ToQueryString() string {
	return sb.sqlQueryString()
}

// Update implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) Update(colAndVal map[string]any) SqlBuilder[RT, RWT] {
	return sb.update(colAndVal)
}

// Where implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) Where(column string, value any) SqlBuilder[RT, RWT] {
	return sb.where(column, value)
}

func (sb *sqlBuilder[RT, RWT]) WhereLike(column string, value any) SqlBuilder[RT, RWT] {
	return sb.whereLike(column, value)
}

// GetArgsValue implements SqlBuilder
func (sb *sqlBuilder[RT, RWT]) GetArgsValue() []any {
	return sb.args
}

type SqlBuilder[RT any, RWT any] interface {
	RowQuery() RT
	RowsQuery() (res RWT, err error)
	ToQueryString() string
	Insert(colAndVal map[string]any) SqlBuilder[RT, RWT]
	Update(colAndVal map[string]any) SqlBuilder[RT, RWT]
	Where(column string, value any) SqlBuilder[RT, RWT]
	WhereLike(column string, value any) SqlBuilder[RT, RWT]
	OrWhere(column string, value any) SqlBuilder[RT, RWT]
	OrderBy(column string, sort string) SqlBuilder[RT, RWT]
	Select(col []string) SqlBuilder[RT, RWT]
	Join(joinType string, target string, conditional string) SqlBuilder[RT, RWT]
	GetArgsValue() []any
}

func NewMysqlBuilder(ctx context.Context, conn *sql.DB, tableName string) SqlBuilder[*sql.Row, *sql.Rows] {
	sb := &sqlBuilder[*sql.Row, *sql.Rows]{
		tableName:  tableName,
		typeSql:    observable.MYSQL,
		connection: conn,
		ctx:        ctx,
	}
	sb.execRowCommand = func() *sql.Row {
		return sb.connection.(*sql.DB).QueryRowContext(sb.ctx, sb.sqlQueryString(), sb.args...)
	}
	sb.execRowsCommand = func() (res *sql.Rows, err error) {
		res, err = sb.connection.(*sql.DB).QueryContext(sb.ctx, sb.sqlQueryString(), sb.args...)
		return res, err
	}
	return sb
}

func NewPgsqlBuilder(ctx context.Context, conn *pgxpool.Pool, tableName string) SqlBuilder[pgx.Row, pgx.Rows] {
	sb := &sqlBuilder[pgx.Row, pgx.Rows]{
		tableName:  tableName,
		typeSql:    observable.PGSQL,
		connection: conn,
		ctx:        ctx,
	}
	sb.execRowCommand = func() pgx.Row {
		return sb.connection.(*pgxpool.Pool).QueryRow(sb.ctx, sb.sqlQueryString(), sb.args...)
	}
	sb.execRowsCommand = func() (res pgx.Rows, err error) {
		res, err = sb.connection.(*pgxpool.Pool).Query(sb.ctx, sb.sqlQueryString(), sb.args...)
		return res, err
	}
	return sb
}

func (sb *sqlBuilder[RT, RWT]) whereLike(column string, value any) *sqlBuilder[RT, RWT] {
	sb.count = sb.count + 1
	sb.called = sb.called + 1
	so := observable.NewSqlObserve(constants.WHERE_LIKE_KEY, sb.typeSql, sb.tableName, true, sb.count, sb.called)
	so.SetColumn(column)
	so.SetValue(value)
	sb.registerObserve(so)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) sqlQueryString() string {
	var query string
	for _, v := range sb.SqlObserve {
		query += v.GetQuery()
		if v.GetValues() != nil {
			sb.args = append(sb.args, v.GetValues()...)
		}
	}
	return query
}

func (sb *sqlBuilder[RT, RWT]) registerObserve(so observable.SqlObserve) {
	sb.SqlObserve = append(sb.SqlObserve, so)
}

func (sb *sqlBuilder[RT, RWT]) insert(colAndVal map[string]any) *sqlBuilder[RT, RWT] {
	sb.setColAndValInserUpdate(constants.INSERT_KEY, colAndVal)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) update(colAndVal map[string]any) *sqlBuilder[RT, RWT] {
	sb.setColAndValInserUpdate(constants.UPDATE_KEY, colAndVal)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) setColAndValInserUpdate(key string, colAndVal map[string]any) {
	so := observable.NewSqlObserve(key, sb.typeSql, sb.tableName, true, 0, 0)
	so.SetCalledTime(len(colAndVal))
	for k, v := range colAndVal {
		sb.count = sb.count + 1
		so.SetCounter(sb.count)
		so.SetColumn(k)
		so.SetValue(v)
	}
	sb.registerObserve(so)
}

func (sb *sqlBuilder[RT, RWT]) orWhere(column string, value any) *sqlBuilder[RT, RWT] {
	sb.count = sb.count + 1
	so := observable.NewSqlObserve(constants.OR_KEY, sb.typeSql, sb.tableName, true, sb.count, 0)
	so.SetColumn(column)
	so.SetValue(value)
	sb.registerObserve(so)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) where(column string, value any) *sqlBuilder[RT, RWT] {
	sb.count = sb.count + 1
	sb.called = sb.called + 1
	so := observable.NewSqlObserve(constants.WHERE_KEY, sb.typeSql, sb.tableName, true, sb.count, sb.called)
	so.SetColumn(column)
	so.SetValue(value)
	sb.registerObserve(so)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) orderBy(column string, sort string) *sqlBuilder[RT, RWT] {
	so := observable.NewSqlObserve(constants.ORDER_BY_KEY, sb.typeSql, sb.tableName, false, 0, 0)
	so.SetColumn(column)
	so.SetValue(sort)
	sb.registerObserve(so)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) selectSql(col []string) *sqlBuilder[RT, RWT] {
	so := observable.NewSqlObserve(constants.SELECT_KEY, sb.typeSql, sb.tableName, false, 0, 0)
	so.SetColumn(col...)
	sb.registerObserve(so)
	return sb
}

func (sb *sqlBuilder[RT, RWT]) join(joinType string, target string, conditional string) *sqlBuilder[RT, RWT] {
	so := observable.NewSqlObserve(fmt.Sprintf("%s %s", strings.ToUpper(joinType), constants.JOIN_KEY), sb.typeSql, sb.tableName, false, 0, 0)
	so.SetValue(conditional)
	sb.registerObserve(so)
	return sb
}
