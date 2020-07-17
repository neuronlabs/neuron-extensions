package migrate

const (
	// TagDB is the tag used for the model structfields that defines 'pq' options.
	TagDB string = "db"
	// TagDatabaseName is the tag key that defines database column name.
	TagDatabaseName string = "name"
	// TagColumnIndex is the tag key that defines database column name.
	TagColumnIndex string = "index"
	// TagDataType is the tag used to set the column's data type.
	TagDataType string = "type"
	// ModelAlreadyPrepared is the key used.
	ModelAlreadyPrepared string = "pq:model_already_prepared"
	// SchemaNameKey is the models store key used to save the schema name.
	SchemaNameKey string = "pq:schema_name"
	// TableKey is the key for the ModelStruct Store that contains table definition.
	TableKey string = "pq:table"
	// ColumnKey is the key for the StructField's Store that contains a column definition.
	ColumnKey string = "pq:column"
	// OmitKey is the key for the StructField's Store that
	OmitKey string = "pq:omit"
	// NotNullKey is the StructField's Store key that defines the NotNull Constraint for the field.
	NotNullKey string = "pq:not_null"
	// DataTypeParametersStoreKey is the key used in the field's stores to the Data type paremeters.
	DataTypeParametersStoreKey string = "pq:data_type_parameters"
)
