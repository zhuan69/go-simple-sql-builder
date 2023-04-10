package observable

import (
	"fmt"
	"strings"

	constants "github.com/zhuan69/go-simple-sql-builder/constants"
)

type SqlType string

const (
	MYSQL SqlType = "mysql"
	PGSQL SqlType = "pgsql"
)

type SqlObserve struct {
	query          strings.Builder
	command        string
	column         []string
	value          []any
	parameterized  bool
	counter        int
	called         int
	tableName      string
	typeSql        SqlType
	paramterPrefix string
}

var conditionalCommandQuery = []string{
	constants.AND_KEY,
	constants.WHERE_LIKE_KEY,
	constants.LIMIT_KEY,
	constants.OR_KEY,
	constants.OFFSET_KEY,
	constants.WHERE_KEY,
}

var baseCommandQuery = []string{
	constants.DELETE_KEY,
	constants.INSERT_KEY,
	constants.SELECT_KEY,
	constants.UPDATE_KEY,
}

func NewSqlObserve(
	command string,
	typeSql SqlType,
	tableName string,
	parameterized bool,
	counter int,
	called int) SqlObserve {
	return SqlObserve{
		command:       command,
		parameterized: parameterized,
		typeSql:       typeSql,
		tableName:     tableName,
		called:        called,
		counter:       counter,
	}
}

func (so *SqlObserve) SetCounter(count int) {
	so.counter = count
}

func (so *SqlObserve) SetCalledTime(count int) {
	so.called = count
}

func (so *SqlObserve) SetValue(val ...any) {
	so.value = append(so.value, val...)
}

func (so *SqlObserve) SetColumn(col ...string) {
	so.column = append(so.column, col...)
}

func (so *SqlObserve) GetValues() []any {
	return so.value
}

func (so *SqlObserve) GetQuery() string {
	colSize := len(so.column)
	valSize := len(so.value)
	if so.command == constants.WHERE_KEY && so.called > 1 {
		so.command = constants.AND_KEY
	}
	if so.isBaseCommandQuery() {
		if so.command == constants.INSERT_KEY {
			if colSize != valSize {
				return fmt.Sprintf("EXPECTED COLUMN AGRS:%d BUT GOT:%d", colSize, valSize)
			}
			so.buildInsertQuery()
			return so.query.String()
		}
		if so.command == constants.SELECT_KEY {
			so.buildSelectQuery()
			return so.query.String()
		}
		if so.command == constants.UPDATE_KEY {
			if colSize != valSize {
				return fmt.Sprintf("EXPECTED COLUMN AGRS:%d BUT GOT:%d", colSize, valSize)
			}
			so.buildUpdateQuery()
			return so.query.String()
		}
	}
	if strings.Contains(so.command, constants.JOIN_KEY) {
		so.buildJoinQuery()
		return so.query.String()
	}
	if so.isParameterizedConditionalQuery() {
		so.parameteredQuery(so.command, so.column[0], so.counter)
		return so.query.String()
	}
	if so.parameterized {
		so.parameteredQuery(so.command, so.column[0], so.counter)
		return so.query.String()
	}
	if so.column[0] == "" && so.value == nil {
		return so.query.String()
	}
	so.normalQuery(so.command, so.column[0], so.value[0])
	return so.query.String()
}

func (so *SqlObserve) buildJoinQuery() {
	query := fmt.Sprintf("%s %s ON %s", so.command, so.tableName, so.value[0])
	so.value = nil
	so.formatQuery(query)
}

func (so *SqlObserve) formatQuery(query string) {
	so.query.WriteString(" " + query)
}

func (so *SqlObserve) buildSelectQuery() {
	size := len(so.column) - 1
	var col strings.Builder
	for i, v := range so.column {
		col.WriteString(v)
		if i != size {
			col.WriteString(",")
		}
	}
	so.query.WriteString(fmt.Sprintf("SELECT %s FROM %s", col.String(), so.tableName))
}

func (so *SqlObserve) buildUpdateQuery() {
	size := len(so.column) - 1
	var q strings.Builder
	for i, v := range so.column {
		so.sanitizeParameterPrefix(i + 1)
		argVal := so.value[i]
		q.WriteString(fmt.Sprintf("%s=%s", v, so.paramterPrefix))
		if i != size {
			q.WriteString(",")
		}
		so.value[i] = argVal
	}
	so.query.WriteString(fmt.Sprintf("UPDATE %s SET %s", so.tableName, q.String()))
}

func (so *SqlObserve) buildInsertQuery() {
	size := len(so.column) - 1
	col := strings.Builder{}
	val := strings.Builder{}
	col.WriteString("(")
	val.WriteString("(")
	for i, v := range so.column {
		so.sanitizeParameterPrefix(i + 1)
		argVal := so.value[i]
		col.WriteString(v)
		val.WriteString(so.paramterPrefix)
		if i != size {
			col.WriteString(",")
			val.WriteString(",")
		}
		so.value[i] = argVal
	}
	col.WriteString(")")
	val.WriteString(")")
	so.query.WriteString(fmt.Sprintf("INSERT INTO %s %s VALUES %s", so.tableName, col.String(), val.String()))
}

func (so *SqlObserve) parameteredQuery(command string, col string, num int) {
	so.sanitizeParameterPrefix(num)
	query := fmt.Sprintf("%s %s=%s", command, col, so.paramterPrefix)
	if command == constants.WHERE_LIKE_KEY {
		query = fmt.Sprintf("WHERE %s LIKE %s", col, so.paramterPrefix)
		if so.called > 1 {
			query = fmt.Sprintf("AND %s LIKE %s", col, so.paramterPrefix)
		}
	}
	so.formatQuery(query)
}

func (so *SqlObserve) normalQuery(command string, col string, val any) {
	q := fmt.Sprintf("%s %s=%v", command, col, val)
	if command == constants.ORDER_BY_KEY {
		q = fmt.Sprintf("%s %s %v", command, col, val)
	}
	so.value = nil
	so.formatQuery(q)
}

func (so *SqlObserve) sanitizeParameterPrefix(num int) {
	if so.typeSql == MYSQL {
		so.paramterPrefix = "?"
	}
	if so.typeSql == PGSQL {
		so.paramterPrefix = fmt.Sprintf("$%d", num)
	}
}

func (so *SqlObserve) isParameterizedConditionalQuery() bool {
	return querySearch(conditionalCommandQuery, 0, len(conditionalCommandQuery)-1, so.command) != -1
}

func (so *SqlObserve) isBaseCommandQuery() bool {
	return querySearch(baseCommandQuery, 0, len(baseCommandQuery)-1, so.command) != -1
}
func querySearch(data []string, begin int, end int, target string) int {
	if begin > end {
		return -1
	}
	median := (begin + end) / 2
	if data[begin] == target {
		return begin
	}
	if data[median] > target {
		return querySearch(data, begin, median-1, target)
	}
	if data[median] == target {

		return median
	}
	if data[median] < target {
		return querySearch(data, median+1, end, target)
	}
	return -1
}
