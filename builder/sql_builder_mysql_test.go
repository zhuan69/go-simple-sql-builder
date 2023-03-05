package builder_test

import (
	"context"
	"go-simple-sql-builder/builder"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItCanGenerateInsertCommandMysql(t *testing.T) {
	builder := builder.NewMysqlBuilder(context.Background(), nil, "testing_table")
	builder.Insert(map[string]any{
		"testing_1": "testing",
		"testing_2": "testing_2",
		"testing_3": 123,
	})
	assert.Equal(t, strings.Count(builder.ToQueryString(), "?"), 3)
	assert.Len(t, builder.GetArgsValue(), 3)
	assert.Contains(t, builder.GetArgsValue(), "testing")
	assert.Contains(t, builder.GetArgsValue(), "testing_2")
	assert.Contains(t, builder.ToQueryString(), "INSERT INTO testing_table")
	assert.Contains(t, builder.GetArgsValue(), 123)
}

func TestItCanGenerateUpdateCommandMysql(t *testing.T) {
	builder := builder.NewMysqlBuilder(context.Background(), nil, "testing_update")
	builder.Update(map[string]any{
		"update_1": "update",
		"update_2": "new value",
		"update_3": 123.5,
	})
	assert.Equal(t, strings.Count(builder.ToQueryString(), "?"), 3)
	assert.Len(t, builder.GetArgsValue(), 3)
	assert.Contains(t, builder.ToQueryString(), "UPDATE testing_update SET")
	assert.Contains(t, builder.GetArgsValue(), "update")
	assert.Contains(t, builder.GetArgsValue(), "new value")
	assert.Contains(t, builder.GetArgsValue(), 123.5)
}
func TestItCanGenerateWhereCommandMysql(t *testing.T) {
	builder := builder.NewMysqlBuilder(context.Background(), nil, "testing_update")
	t.Run("Testing it can generate 'where' conditional queries", func(t *testing.T) {
		builder.Where("conditional", time.Now().Format("2006-01-02 15:04:05"))
		assert.Contains(t, builder.ToQueryString(), "WHERE conditional=?")
		assert.Equal(t, strings.Count(builder.ToQueryString(), "?"), 1)
		assert.Contains(t, builder.GetArgsValue(), time.Now().Format("2006-01-02 15:04:05"))
	})
	t.Run("Testing it can generate 'and' conditional query when where invoked more than 1", func(t *testing.T) {
		builder.Where("another_conditional", "testing")
		assert.Contains(t, builder.ToQueryString(), "AND another_conditional=?")
		assert.Equal(t, strings.Count(builder.ToQueryString(), "?"), 2)
		assert.Contains(t, builder.GetArgsValue(), "testing")
	})
	t.Run("Testing it can generate 'like' conditional query", func(t *testing.T) {
		builder.WhereLike("like_pattern", `%value%`)
		assert.Contains(t, builder.ToQueryString(), "AND like_pattern LIKE")
		assert.Equal(t, strings.Count(builder.ToQueryString(), "?"), 3)
		assert.Contains(t, builder.GetArgsValue(), `%value%`)
	})
	t.Run("Testing it can generate 'or' conditional query", func(t *testing.T) {
		builder.OrWhere("or_pattern", 4321)
		assert.Contains(t, builder.ToQueryString(), "OR or_pattern=?")
		assert.Equal(t, strings.Count(builder.ToQueryString(), "?"), 4)
		assert.Contains(t, builder.GetArgsValue(), 4321)
	})
}

func TestItCanGenerateSelectQueryMysql(t *testing.T) {
	builder := builder.NewMysqlBuilder(context.Background(), nil, "testing_select")
	builder.Select([]string{
		"column_1 as col_1",
		"column_2",
		"column_3",
	})
	assert.Contains(t, builder.ToQueryString(), "column_1 as col_1")
	assert.Contains(t, builder.ToQueryString(), "column_2")
	assert.Contains(t, builder.ToQueryString(), "column_3")
	assert.Contains(t, builder.ToQueryString(), "SELECT")
	assert.Contains(t, builder.ToQueryString(), "FROM testing_select")
}

func TestItCanGenerateJoinTableQueryMysql(t *testing.T) {
	builder := builder.NewMysqlBuilder(context.Background(), nil, "testing_join")
	builder.JoinTable("left", "first_table as ft", "ft.id=testing_join.id")
	assert.Equal(t, " LEFT JOIN first_table as ft ON ft.id=testing_join.id", builder.ToQueryString())
}

func TestItCanGenerateOrderByQueryMysql(t *testing.T) {
	builder := builder.NewMysqlBuilder(context.Background(), nil, "testing_order")
	builder.OrderBy("order_col", "asc")
	assert.Equal(t, " ORDER BY order_col asc", builder.ToQueryString())
}
