package pop

import (
	"fmt"
	"strings"
)

type joinClause struct {
	JoinType string
	Table    string
	On       clauses
}

type joinClauses []joinClause

func (c joinClause) String() string {
	sql := fmt.Sprintf("%s %s", c.JoinType, c.Table)

	if len(c.On) > 0 {
		sql += " ON " + c.On.Join(" AND ")
	}

	return sql
}

func (c joinClause) Args() []interface{} {
	return c.On.Args()
}

func (c joinClauses) String() string {
	var cs []string
	for _, cl := range c {
		cs = append(cs, cl.String())
	}
	return strings.Join(cs, " ")
}
