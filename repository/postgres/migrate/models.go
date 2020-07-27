package migrate

import (
	"context"
	"reflect"
	"sort"
	"time"

	"github.com/neuronlabs/inflection"
	"github.com/neuronlabs/strcase"

	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/log"
)

// AutoMigrateModels migrates the provided model definitions.
func AutoMigrateModels(ctx context.Context, conn internal.Connection, models ...*mapping.ModelStruct) error {
	return autoMigrateModels(ctx, conn, models...)
}

// PrepareModels prepares multiple models, get's their db tags.
func PrepareModels(models ...*mapping.ModelStruct) error {
	return prepareModels(models...)
}

// PrepareModel prepares the model for the pq db tags.
func PrepareModel(model *mapping.ModelStruct) error {
	return prepareModel(model)
}

// AutoMigrateModel prepares the model's struct and automatically migrates it's table into the provided database structure.
func AutoMigrateModel(ctx context.Context, conn internal.Connection, model *mapping.ModelStruct) error {
	if err := prepareModel(model); err != nil {
		return err
	}

	if err := autoMigrateModel(ctx, conn, model); err != nil {
		return err
	}

	t, err := modelsTable(model)
	if err != nil {
		return err
	}

	if err = t.autoMigrateConstraints(ctx, conn); err != nil {
		return err
	}

	return nil
}

/**

PRIVATE FUNCTIONS

*/

func autoMigrateModels(ctx context.Context, conn internal.Connection, models ...*mapping.ModelStruct) (err error) {
	log.Debugf("Auto migrating %d models...", len(models))
	ts := time.Now()
	defer func() {
		log.Debugf("Finished migrating models in: %d ns", time.Since(ts).Nanoseconds())
	}()

	// iterate over all models and do 'autoMigrateModel'
	for _, model := range models {
		log.Debugf("Migrating model: '%s'", model.Type().Name())

		if err = prepareModel(model); err != nil {
			return err
		}
		if err = autoMigrateModel(ctx, conn, model); err != nil {
			return err
		}
	}

	// migrate constraints
	for _, model := range models {
		t, err := modelsTable(model)
		if err != nil {
			return err
		}

		if err = t.autoMigrateConstraints(ctx, conn); err != nil {
			return err
		}
	}
	return nil
}

func autoMigrateModel(ctx context.Context, conn internal.Connection, m *mapping.ModelStruct) error {
	t, err := modelsTable(m)
	if err != nil {
		return err
	}

	if err = t.autoMigrate(ctx, conn); err != nil {
		return err
	}
	return nil
}

func prepareModels(models ...*mapping.ModelStruct) (err error) {
	for _, model := range models {
		// prepare model 'm'
		if err = prepareModel(model); err != nil {
			return err
		}
	}
	return nil
}

var _ sort.Interface = fieldsSorter{}

type fieldsSorter []*mapping.StructField

func (f fieldsSorter) Len() int {
	return len(f)
}

func (f fieldsSorter) Less(i, j int) bool {
	firstIndex := f[i].Index
	secondIndex := f[j].Index
	var isLess func(k int) bool

	isLess = func(k int) bool {
		if firstIndex[k] == secondIndex[k] {
			return isLess(k + 1)
		}
		return firstIndex[k] < secondIndex[k]
	}
	return isLess(0)
}

func (f fieldsSorter) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func prepareModel(model *mapping.ModelStruct) error {
	_, ok := model.StoreGet(ModelAlreadyPrepared)
	if ok {
		return nil
	}

	// Get the interface of new model type
	modelValue := reflect.New(model.Type()).Interface()
	// postgres schema name
	schemaName := "public"
	// check the interface
	schemaNamer, ok := modelValue.(SchemaNamer)
	if ok {
		schemaName = schemaNamer.PQSchemaName()
	}
	model.StoreSet(SchemaNameKey, schemaName)

	// tableName defines the table name used
	var tableName string
	tableNamer, ok := modelValue.(TableNamer)
	if ok {
		tableName = tableNamer.TableName()
	} else {
		tableName = strcase.ToSnake(inflection.Plural(model.Type().Name()))
	}
	// create the table with given schemaName, tableName and defined model.
	table := &Table{
		Schema: schemaName,
		Name:   tableName,
		model:  model,
	}
	// AddModel the TableName for given model
	model.StoreSet(TableKey, table)
	// Iterate over structFields and set the proper field dbname
	fields := mapping.OrderedFieldset(model.StructFields())
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch field.Kind() {
		case mapping.KindPrimary, mapping.KindAttribute, mapping.KindForeignKey:
		default:
			fields = append(fields[:i], fields[i+1:]...)
			i--
			continue
		}
	}
	// sort the fields
	sort.Sort(fields)

	// iterate over the fields to create the columns for the model.
	for _, field := range fields {
		// extractTags
		tags := field.ExtractCustomFieldTags(TagDB, mapping.AnnotationSeparator, " ")
		for _, tag := range tags {
			// check if field shouldn't be used in db context
			if tag.Key == "-" {
				// store it as OmitKey in the field's definition
				field.StoreSet(OmitKey, struct{}{})
				break
			}
		}
		if _, ok := field.StoreGet(OmitKey); ok {
			log.Debug2f("[%s] Field: '%s' is not used in the database context", model.Collection(), field.NeuronName())
			break
		}
		log.Debug2f("[%s] Creating column for field: %s", model.Collection(), field.NeuronName())
		err := table.createColumn(field, tags)
		if err != nil {
			return err
		}
	}

	model.StoreSet(ModelAlreadyPrepared, true)
	return nil
}
