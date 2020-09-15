package pop

import "strings"

// EagerMode type for all eager modes supported in pop.
type EagerMode uint8

const (
	eagerModeNil EagerMode = iota
	// EagerDefault is the current implementation, the default
	// behavior of pop. This one introduce N+1 problem and will be used as
	// default value for backward compatibility.
	EagerDefault

	// EagerPreload mode works similar to Preload mode used in Rails ActiveRecord.
	// Avoid N+1 problem by reducing the number of hits to the database but
	// increase memory use to process and link associations to parent.
	EagerPreload

	// EagerInclude This mode works similar to Include mode used in rails ActiveRecord.
	// Use Left Join clauses to load associations. Not working yet.
	EagerInclude
)

// default loading Association Strategy definition.
var loadingAssociationsStrategy = EagerDefault

// SetEagerMode changes overall mode when eager loading.
// this will change the default loading associations strategy for all Eager queries.
// This will affect all queries when eager loading is used.
func SetEagerMode(eagerMode EagerMode) {
	loadingAssociationsStrategy = eagerMode
}

// AvailableDialects lists the available database dialects
var AvailableDialects []string

var dialectSynonyms = make(map[string]string)

// map of dialect specific url parsers
var urlParser = make(map[string]func(*ConnectionDetails) error)

// map of dialect specific connection details finalizers
var finalizer = make(map[string]func(*ConnectionDetails))

// map of connection creators
var newConnection = make(map[string]func(*ConnectionDetails) (dialect, error))

// map of modes that can be enabled
var modes = map[string]func(*Connection, *ConnectionDetails) error{
	"":        func(*Connection, *ConnectionDetails) error { return nil },
	"default": func(*Connection, *ConnectionDetails) error { return nil },
}

// DialectSupported checks support for the given database dialect
func DialectSupported(d string) bool {
	for _, ad := range AvailableDialects {
		if ad == d {
			return true
		}
	}
	return false
}

func normalizeSynonyms(dialect string) string {
	d := strings.ToLower(dialect)
	if syn, ok := dialectSynonyms[d]; ok {
		d = syn
	}
	return d
}
