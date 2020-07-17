package filters

import (
	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/log"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/migrate"
)

// ParseFilters parses the filters into SQLQueries for the provided scope.
func ParseFilters(s *query.Scope, writer internal.QuotedWordWriteFunc) (SQLQueries, error) {
	queries := SQLQueries{}

	// at first get primary filters
	for _, filter := range s.Filters {
		if _, ok := filter.StructField.StoreGet(migrate.OmitKey); ok {
			log.Debug2f("Skipping foreign key filter with db:\"-\" omit option")
			continue
		}
		for i := range filter.Values {
			sqlizer, err := getOperatorSQLizer(filter.Values[i].Operator)
			if err != nil {
				err := errors.NewDet(query.ClassFilterFormat, "unsupported filter operator")
				err.WithDetailf("Provided unsupported operator: '%s' for given query.", filter.Values[i].Operator.Name)
				return nil, err
			}

			subQueries, err := sqlizer(s, writer, filter.StructField, &filter.Values[i])
			if err != nil {
				return nil, err
			}

			queries = append(queries, subQueries...)
		}
	}
	return queries, nil
}
