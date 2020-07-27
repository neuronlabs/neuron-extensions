package jsonapi

const (
	// MediaType is the identifier for the JSON API media type
	// see http://jsonapi.org/format/#document-structure
	MimeType = "application/vnd.api+json"
	// KeyTotal is the meta key for the 'total' instances information.
	KeyTotal = "total"
	// StructTag is the jsonapi defined struct tag.
	StructTag = "jsonapi"
)
