package query

import (
	"fmt"
	"strings"
)

// QueryBuilder constructs SQL queries from visual components
type QueryBuilder struct {
	table      string
	schema     string
	columns    []Column
	joins      []Join
	conditions []Condition
	groupBy    []string
	orderBy    []OrderBy
	limit      int
	offset     int
	distinct   bool
}

// Column represents a selected column
type Column struct {
	Name       string     `json:"name"`
	Alias      string     `json:"alias,omitempty"`
	Table      string     `json:"table,omitempty"`
	Aggregate  Aggregate  `json:"aggregate,omitempty"`
	Expression string     `json:"expression,omitempty"` // For computed columns
}

// Aggregate represents an aggregate function
type Aggregate string

const (
	AggNone    Aggregate = ""
	AggCount   Aggregate = "COUNT"
	AggSum     Aggregate = "SUM"
	AggAvg     Aggregate = "AVG"
	AggMin     Aggregate = "MIN"
	AggMax     Aggregate = "MAX"
	AggArrayAgg Aggregate = "ARRAY_AGG"
	AggStringAgg Aggregate = "STRING_AGG"
	// PostGIS aggregates
	AggSTUnion    Aggregate = "ST_Union"
	AggSTCollect  Aggregate = "ST_Collect"
	AggSTExtent   Aggregate = "ST_Extent"
)

// Join represents a table join
type Join struct {
	Type       JoinType `json:"type"`
	Table      string   `json:"table"`
	Schema     string   `json:"schema,omitempty"`
	Alias      string   `json:"alias,omitempty"`
	OnLeft     string   `json:"on_left"`     // Left side of ON condition
	OnRight    string   `json:"on_right"`    // Right side of ON condition
	OnOperator string   `json:"on_operator"` // Usually "="
}

// JoinType represents the type of join
type JoinType string

const (
	JoinInner JoinType = "INNER JOIN"
	JoinLeft  JoinType = "LEFT JOIN"
	JoinRight JoinType = "RIGHT JOIN"
	JoinFull  JoinType = "FULL OUTER JOIN"
	JoinCross JoinType = "CROSS JOIN"
)

// Condition represents a WHERE condition
type Condition struct {
	Column    string      `json:"column"`
	Operator  Operator    `json:"operator"`
	Value     interface{} `json:"value"`
	ValueType string      `json:"value_type"` // "literal", "column", "subquery"
	Logic     Logic       `json:"logic"`      // AND, OR for combining conditions
	Negate    bool        `json:"negate"`     // NOT
}

// Operator represents a comparison operator
type Operator string

const (
	OpEqual          Operator = "="
	OpNotEqual       Operator = "!="
	OpLess           Operator = "<"
	OpLessEqual      Operator = "<="
	OpGreater        Operator = ">"
	OpGreaterEqual   Operator = ">="
	OpLike           Operator = "LIKE"
	OpILike          Operator = "ILIKE"
	OpIn             Operator = "IN"
	OpNotIn          Operator = "NOT IN"
	OpIsNull         Operator = "IS NULL"
	OpIsNotNull      Operator = "IS NOT NULL"
	OpBetween        Operator = "BETWEEN"
	// PostGIS operators
	OpSTIntersects   Operator = "ST_Intersects"
	OpSTContains     Operator = "ST_Contains"
	OpSTWithin       Operator = "ST_Within"
	OpSTDWithin      Operator = "ST_DWithin"
	OpSTEquals       Operator = "ST_Equals"
	OpSTTouches      Operator = "ST_Touches"
	OpSTOverlaps     Operator = "ST_Overlaps"
	OpSTCrosses      Operator = "ST_Crosses"
)

// Logic represents logical operators
type Logic string

const (
	LogicAnd Logic = "AND"
	LogicOr  Logic = "OR"
)

// OrderBy represents an ORDER BY clause
type OrderBy struct {
	Column    string `json:"column"`
	Direction string `json:"direction"` // ASC, DESC
	NullsPos  string `json:"nulls_pos"` // NULLS FIRST, NULLS LAST
}

// QueryDefinition is the serializable form of a visual query
type QueryDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Schema      string      `json:"schema"`
	Table       string      `json:"table"`
	Columns     []Column    `json:"columns"`
	Joins       []Join      `json:"joins,omitempty"`
	Conditions  []Condition `json:"conditions,omitempty"`
	GroupBy     []string    `json:"group_by,omitempty"`
	OrderBy     []OrderBy   `json:"order_by,omitempty"`
	Limit       int         `json:"limit,omitempty"`
	Offset      int         `json:"offset,omitempty"`
	Distinct    bool        `json:"distinct,omitempty"`
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columns:    []Column{},
		joins:      []Join{},
		conditions: []Condition{},
		groupBy:    []string{},
		orderBy:    []OrderBy{},
		limit:      100,
	}
}

// FromDefinition creates a query builder from a definition
func FromDefinition(def QueryDefinition) *QueryBuilder {
	return &QueryBuilder{
		schema:     def.Schema,
		table:      def.Table,
		columns:    def.Columns,
		joins:      def.Joins,
		conditions: def.Conditions,
		groupBy:    def.GroupBy,
		orderBy:    def.OrderBy,
		limit:      def.Limit,
		offset:     def.Offset,
		distinct:   def.Distinct,
	}
}

// Table sets the main table
func (qb *QueryBuilder) Table(schema, table string) *QueryBuilder {
	qb.schema = schema
	qb.table = table
	return qb
}

// Select adds columns to select
func (qb *QueryBuilder) Select(columns ...Column) *QueryBuilder {
	qb.columns = append(qb.columns, columns...)
	return qb
}

// SelectAll selects all columns
func (qb *QueryBuilder) SelectAll() *QueryBuilder {
	qb.columns = []Column{{Name: "*"}}
	return qb
}

// Join adds a join
func (qb *QueryBuilder) Join(join Join) *QueryBuilder {
	qb.joins = append(qb.joins, join)
	return qb
}

// Where adds a condition
func (qb *QueryBuilder) Where(condition Condition) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// GroupBy sets GROUP BY columns
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// OrderBy adds ORDER BY clause
func (qb *QueryBuilder) OrderBy(order OrderBy) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, order)
	return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limit = n
	return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(n int) *QueryBuilder {
	qb.offset = n
	return qb
}

// Distinct sets DISTINCT
func (qb *QueryBuilder) Distinct(d bool) *QueryBuilder {
	qb.distinct = d
	return qb
}

// Build generates the SQL query
func (qb *QueryBuilder) Build() (string, []interface{}, error) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 1

	// SELECT clause
	sql.WriteString("SELECT ")
	if qb.distinct {
		sql.WriteString("DISTINCT ")
	}

	if len(qb.columns) == 0 {
		sql.WriteString("*")
	} else {
		var cols []string
		for _, col := range qb.columns {
			colStr := qb.buildColumn(col)
			cols = append(cols, colStr)
		}
		sql.WriteString(strings.Join(cols, ", "))
	}

	// FROM clause
	sql.WriteString("\nFROM ")
	if qb.schema != "" {
		sql.WriteString(fmt.Sprintf("%s.", quoteIdentifier(qb.schema)))
	}
	sql.WriteString(quoteIdentifier(qb.table))

	// JOIN clauses
	for _, join := range qb.joins {
		sql.WriteString(fmt.Sprintf("\n%s ", join.Type))
		if join.Schema != "" {
			sql.WriteString(fmt.Sprintf("%s.", quoteIdentifier(join.Schema)))
		}
		sql.WriteString(quoteIdentifier(join.Table))
		if join.Alias != "" {
			sql.WriteString(fmt.Sprintf(" AS %s", quoteIdentifier(join.Alias)))
		}
		if join.Type != JoinCross {
			op := join.OnOperator
			if op == "" {
				op = "="
			}
			sql.WriteString(fmt.Sprintf(" ON %s %s %s", join.OnLeft, op, join.OnRight))
		}
	}

	// WHERE clause
	if len(qb.conditions) > 0 {
		sql.WriteString("\nWHERE ")
		for i, cond := range qb.conditions {
			if i > 0 {
				sql.WriteString(fmt.Sprintf(" %s ", cond.Logic))
			}
			condStr, condArgs := qb.buildCondition(cond, &argIndex)
			sql.WriteString(condStr)
			args = append(args, condArgs...)
		}
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		sql.WriteString("\nGROUP BY ")
		sql.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		sql.WriteString("\nORDER BY ")
		var orders []string
		for _, ob := range qb.orderBy {
			orderStr := ob.Column
			if ob.Direction != "" {
				orderStr += " " + ob.Direction
			}
			if ob.NullsPos != "" {
				orderStr += " " + ob.NullsPos
			}
			orders = append(orders, orderStr)
		}
		sql.WriteString(strings.Join(orders, ", "))
	}

	// LIMIT clause
	if qb.limit > 0 {
		sql.WriteString(fmt.Sprintf("\nLIMIT %d", qb.limit))
	}

	// OFFSET clause
	if qb.offset > 0 {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return sql.String(), args, nil
}

func (qb *QueryBuilder) buildColumn(col Column) string {
	var result string

	if col.Expression != "" {
		result = col.Expression
	} else {
		if col.Table != "" {
			result = fmt.Sprintf("%s.%s", quoteIdentifier(col.Table), quoteIdentifier(col.Name))
		} else {
			result = quoteIdentifier(col.Name)
		}
	}

	if col.Aggregate != AggNone {
		result = fmt.Sprintf("%s(%s)", col.Aggregate, result)
	}

	if col.Alias != "" {
		result = fmt.Sprintf("%s AS %s", result, quoteIdentifier(col.Alias))
	}

	return result
}

func (qb *QueryBuilder) buildCondition(cond Condition, argIndex *int) (string, []interface{}) {
	var sql string
	var args []interface{}

	column := cond.Column

	if cond.Negate {
		sql = "NOT "
	}

	switch cond.Operator {
	case OpIsNull:
		sql += fmt.Sprintf("%s IS NULL", column)
	case OpIsNotNull:
		sql += fmt.Sprintf("%s IS NOT NULL", column)
	case OpIn, OpNotIn:
		// Handle IN/NOT IN with array
		if values, ok := cond.Value.([]interface{}); ok {
			placeholders := make([]string, len(values))
			for i, v := range values {
				placeholders[i] = fmt.Sprintf("$%d", *argIndex)
				args = append(args, v)
				*argIndex++
			}
			sql += fmt.Sprintf("%s %s (%s)", column, cond.Operator, strings.Join(placeholders, ", "))
		}
	case OpBetween:
		if values, ok := cond.Value.([]interface{}); ok && len(values) == 2 {
			sql += fmt.Sprintf("%s BETWEEN $%d AND $%d", column, *argIndex, *argIndex+1)
			args = append(args, values[0], values[1])
			*argIndex += 2
		}
	case OpSTIntersects, OpSTContains, OpSTWithin, OpSTEquals, OpSTTouches, OpSTOverlaps, OpSTCrosses:
		// PostGIS spatial predicates
		sql += fmt.Sprintf("%s(%s, $%d)", cond.Operator, column, *argIndex)
		args = append(args, cond.Value)
		*argIndex++
	case OpSTDWithin:
		// ST_DWithin requires distance parameter
		if values, ok := cond.Value.([]interface{}); ok && len(values) == 2 {
			sql += fmt.Sprintf("ST_DWithin(%s, $%d, $%d)", column, *argIndex, *argIndex+1)
			args = append(args, values[0], values[1])
			*argIndex += 2
		}
	default:
		// Standard comparison operators
		if cond.ValueType == "column" {
			sql += fmt.Sprintf("%s %s %s", column, cond.Operator, cond.Value)
		} else {
			sql += fmt.Sprintf("%s %s $%d", column, cond.Operator, *argIndex)
			args = append(args, cond.Value)
			*argIndex++
		}
	}

	return sql, args
}

// ToDefinition exports the query as a definition
func (qb *QueryBuilder) ToDefinition(name string) QueryDefinition {
	return QueryDefinition{
		Name:       name,
		Schema:     qb.schema,
		Table:      qb.table,
		Columns:    qb.columns,
		Joins:      qb.joins,
		Conditions: qb.conditions,
		GroupBy:    qb.groupBy,
		OrderBy:    qb.orderBy,
		Limit:      qb.limit,
		Offset:     qb.offset,
		Distinct:   qb.distinct,
	}
}

// quoteIdentifier quotes a SQL identifier
func quoteIdentifier(name string) string {
	// Don't quote if already quoted or is *
	if strings.HasPrefix(name, "\"") || name == "*" {
		return name
	}
	// Quote if contains special characters or is a reserved word
	if strings.ContainsAny(name, " -./") || isReservedWord(name) {
		return fmt.Sprintf("\"%s\"", name)
	}
	return name
}

// isReservedWord checks if a name is a PostgreSQL reserved word
func isReservedWord(name string) bool {
	reserved := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "JOIN": true,
		"ORDER": true, "GROUP": true, "LIMIT": true, "OFFSET": true,
		"AND": true, "OR": true, "NOT": true, "NULL": true,
		"TABLE": true, "INDEX": true, "VIEW": true, "AS": true,
		"ON": true, "USING": true, "ALL": true, "DISTINCT": true,
	}
	return reserved[strings.ToUpper(name)]
}
