package pop

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/pop/v5/columns"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
)

func init() {
	modes["history"] = func(c *Connection, deets *ConnectionDetails) error {
		suffix := deets.Options["suffix"]
		if suffix == "" {
			suffix = "_history"
		}
		delete(deets.Options, "suffix")
		c.Dialect = &historyDB{
			dialect: c.Dialect,
			Suffix:  suffix,
			mapper:  reflectx.NewMapper("db"),
		}
		return nil
	}
}

var _ dialect = &historyDB{}

type historyDB struct {
	mapper *reflectx.Mapper
	dialect
	Suffix string
}

func (h historyDB) insertHistory(s store, m *Model, t time.Time) error {
	idField := m.IDField()
	id := m.ID()

	cols := columns.ForStruct(m.Value, m.TableName())
	cols.Add(idField)
	r := cols.Readable()
	w := cols.Writeable()
	w.Add(idField)

	createdAt, deletedAt := h.Quote("created_at"), h.Quote("deleted_at")

	sql := h.TranslateSQL(fmt.Sprintf(
		`INSERT INTO %v (%v, %v, %v)
		SELECT %v, ? AS %v, NULL AS %v
		FROM %v WHERE %v = ?`,
		h.Quote(m.TableName()+h.Suffix), w.QuotedString(h), createdAt, deletedAt,
		r.SelectString(), createdAt, deletedAt,
		h.Quote(m.TableName()), h.Quote(idField),
	))
	_, err := s.Exec(sql, t, id)
	return err
}
func (h historyDB) deleteHistory(s store, m *Model, t time.Time) error {
	idField := m.IDField()
	id := m.ID()

	deletedAt := h.Quote("deleted_at")

	sql := h.TranslateSQL(fmt.Sprintf(
		`UPDATE %v
		SET %v = ?
		WHERE %v = ? AND %v IS NULL;`,
		h.Quote(m.TableName()+h.Suffix),
		deletedAt, h.Quote(idField), deletedAt,
	))
	_, err := s.Exec(sql, t, id)
	return err
}

func (h historyDB) FizzTranslator() fizz.Translator {
	return historyTranslator{h.dialect.FizzTranslator(), h.Suffix}
}

func (h historyDB) Create(s store, m *Model, c columns.Columns) error {
	err := h.dialect.Create(s, m, c)
	if err != nil {
		return err
	}
	t := time.Now().UTC()
	return errors.Wrap(h.insertHistory(s, m, t), "historical insert")
}
func (h historyDB) Update(s store, m *Model, c columns.Columns) error {
	err := h.dialect.Update(s, m, c)
	if err != nil {
		return err
	}
	t := time.Now().UTC()
	err = errors.Wrap(h.deleteHistory(s, m, t), "historical delete")
	if err != nil {
		return err
	}
	return errors.Wrap(h.insertHistory(s, m, t), "historical insert")
}
func (h historyDB) Destroy(s store, m *Model) error {
	err := h.dialect.Destroy(s, m)
	if err != nil {
		return err
	}
	t := time.Now().UTC()
	return errors.Wrap(h.deleteHistory(s, m, t), "historical delete")
}

func (h historyDB) QueryHistory(q *Query, t time.Time) *Query {
	c := clause{
		"created_at <= ? AND (deleted_at IS NULL OR deleted_at > ?)",
		[]interface{}{t, t},
	}
	q.whereClauses = append(q.whereClauses, c)
	q.globalClauses = append(q.globalClauses, c)
	q.tablePattern = "%v" + h.Suffix
	return q
}

type queryHistoryer interface {
	QueryHistory(*Query, time.Time) *Query
}

// Returns a scope that will cause the Query object to query the historical
// tables in the database, instead of the primary ones, filtered to give
// information as it was at the specified time.
//
// If the current dialect of the connection does not support historical
// information, the scope will panic.
func HistoryScope(t time.Time) ScopeFunc {
	return func(q *Query) *Query {
		h, ok := q.Connection.Dialect.(queryHistoryer)
		if !ok {
			panic("may only query history for history dialect")
		}
		return h.QueryHistory(q, t)
	}
}

type historyTranslator struct {
	fizz.Translator
	Suffix string
}

func (h historyTranslator) apply(f func(fizz.Table) (string, error), t fizz.Table) (string, error) {
	if strings.HasSuffix(t.Name, h.Suffix) {
		return "", errors.New("operation already applies to historical tables")
	}
	sql, err := f(t)
	if err != nil {
		return "", err
	}
	t.Name += h.Suffix
	for i := range t.Indexes {
		t.Indexes[i].Unique = false
	}
	a, err := f(t)
	return sql + "\n" + a, err
}

func (h historyTranslator) CreateTable(t fizz.Table) (string, error) {
	if t.Options == nil {
		t.Options = make(map[string]interface{})
	}
	t.DisableTimestamps()
	sql, err := h.Translator.CreateTable(t)
	if err != nil {
		return "", err
	}
	t.Name += h.Suffix
	t.ForeignKeys = []fizz.ForeignKey{}
	t.Indexes = []fizz.Index{}
	for i := range t.Columns {
		t.Columns[i].Primary = false
	}

	if err = t.Column("created_at", "timestamp", fizz.Options{}); err != nil {
		return "", errors.Wrap(err, "historical table")
	}
	if err = t.Column("deleted_at", "timestamp", fizz.Options{"null": true}); err != nil {
		return "", errors.Wrap(err, "historical table")
	}
	a, err := h.Translator.CreateTable(t)
	return sql + "\n" + a, err
}

func (h historyTranslator) DropTable(t fizz.Table) (string, error) {
	return h.apply(h.Translator.DropTable, t)
}
func (h historyTranslator) RenameTable(t []fizz.Table) (string, error) {
	if strings.HasSuffix(t[0].Name, h.Suffix) || strings.HasSuffix(t[1].Name, h.Suffix) {
		return "", errors.New("operation already applies to historical tables")
	}
	sql, err := h.Translator.RenameTable(t)
	if err != nil {
		return "", err
	}
	t[0].Name += h.Suffix
	t[1].Name += h.Suffix
	a, err := h.Translator.RenameTable(t)
	return sql + "\n" + a, err
}
func (h historyTranslator) AddColumn(t fizz.Table) (string, error) {
	return h.apply(h.Translator.AddColumn, t)
}
func (h historyTranslator) ChangeColumn(t fizz.Table) (string, error) {
	return h.apply(h.Translator.ChangeColumn, t)
}
func (h historyTranslator) DropColumn(t fizz.Table) (string, error) {
	return h.apply(h.Translator.DropColumn, t)
}
func (h historyTranslator) RenameColumn(t fizz.Table) (string, error) {
	return h.apply(h.Translator.RenameColumn, t)
}
