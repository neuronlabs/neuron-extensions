package jsonapi

const (
	// MediaType is the identifier for the JSON API media type
	// see http://jsonapi.org/format/#document-structure
	MimeType = "application/vnd.api+json"
	// ISO8601TimeFormat is the time formatting for the ISO 8601.
	ISO8601TimeFormat = "2006-01-02T15:04:05Z"
	// KeyTotal is the meta key for the 'total' instances information.
	KeyTotal = "total"
	// StructTag is the jsonapi defined struct tag.
	StructTag = "jsonapi"
)
