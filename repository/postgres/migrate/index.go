package migrate

// Index defines the postgres Index.
type Index struct {
	// Name is the index name.
	Name string
	// Type defines the index type.
	Type IndexType
	// Columns are the columns specified for the index.
	Columns []*Column
}

// IndexType defines the postgres index type.
type IndexType int

const (
	// BTree is the Postgres B-Tree index type - default.
	BTree IndexType = iota
	// Hash is the postgres Hash index type.
	Hash
	// GiST is the postgres GiST index type.
	GiST
	// GIN is the postgres GIN index type.
	GIN
)

const (
	// BTreeTag is the BTree index type tag.
	BTreeTag = "btree"
	// HashTag is the Hash index type tag.
	HashTag = "hash"
	// GiSTTag is the GiST index type tag.
	GiSTTag = "gist"
	// GINTag is the GIN index type tag.
	GINTag = "gin"
)
