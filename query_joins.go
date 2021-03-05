package pop

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/pop/v5/logging"
)

func (q *Query) join(joinType, table, on string, args []interface{}) *Query {
	if q.RawSQL.Fragment != "" {
		log(logging.Warn, "Query is setup to use raw SQL")
		return q
	}

	c := joinClause{
		JoinType: joinType,
		Table:    fmt.Sprintf(q.tablePattern+" AS %v", table, table),
		On:       clauses{clause{Fragment: on, Arguments: args}},
	}
	// replace `%TABLE_NAME%` with `table`
	for i := range q.globalClauses {
		q.globalClauses[i].Fragment = strings.ReplaceAll(q.globalClauses[i].Fragment, AliasToken, table)
	}
	c.On = append(c.On, q.globalClauses...)
	q.joinClauses = append(q.joinClauses, c)
	return q
}

// Join will append a JOIN clause to the query
func (q *Query) Join(table string, on string, args ...interface{}) *Query {
	return q.join("JOIN", table, on, args)
}

// LeftJoin will append a LEFT JOIN clause to the query
func (q *Query) LeftJoin(table string, on string, args ...interface{}) *Query {
	return q.join("LEFT JOIN", table, on, args)
}

// RightJoin will append a RIGHT JOIN clause to the query
func (q *Query) RightJoin(table string, on string, args ...interface{}) *Query {
	return q.join("RIGHT JOIN", table, on, args)
}

// LeftOuterJoin will append a LEFT OUTER JOIN clause to the query
func (q *Query) LeftOuterJoin(table string, on string, args ...interface{}) *Query {
	return q.join("LEFT OUTER JOIN", table, on, args)
}

// RightOuterJoin will append a RIGHT OUTER JOIN clause to the query
func (q *Query) RightOuterJoin(table string, on string, args ...interface{}) *Query {
	return q.join("RIGHT OUTER JOIN", table, on, args)
}

// InnerJoin will append an INNER JOIN clause to the query
func (q *Query) InnerJoin(table string, on string, args ...interface{}) *Query {
	return q.join("INNER JOIN", table, on, args)
}
