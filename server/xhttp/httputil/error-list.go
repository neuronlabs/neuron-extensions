package httputil

import (
	"github.com/neuronlabs/neuron/codec"
)

/**

STATUS 2xx

*/

// ErrWarningNotification warns on response with some value.
func ErrWarningNotification() *codec.Error {
	return &codec.Error{
		Title:  "The warning notification occurred.",
		Status: "200",
	}
}

/**

STATUS 400

*/

// ErrBadRequest is the API error thrown on bad request
func ErrBadRequest() *codec.Error {
	return &codec.Error{
		Title:  "The server cannot or will not process the request due to something that is perceived to be a client error",
		Status: "400",
	}
}

// ErrHeadersNotSupported errors thrown when the provided HTTP headers are not supported by the server
func ErrHeadersNotSupported() *codec.Error {
	return &codec.Error{
		Title: `The conditional headers provided in the request are not supported, 
		by the server.`,
		Status: "400",
	}
}

// ErrInvalidAuthenticationInfo defines the error when the authentication fails.
func ErrInvalidAuthenticationInfo() *codec.Error {
	return &codec.Error{
		Title:  `The authentication information was not provided in the correct format.`,
		Status: "401",
	}
}

// ErrInvalidHeaderValue responded when some HTTP header was not in a valid format.
func ErrInvalidHeaderValue() *codec.Error {
	return &codec.Error{
		Title:  "The value provided in one of the HTTP headers was not in the correct format.",
		Status: "400",
	}
}

// ErrInvalidInput returned when provided request input is not valid.
func ErrInvalidInput() *codec.Error {
	return &codec.Error{
		Title:  "One of the request inputs is not valid.",
		Status: "400",
	}
}

// ErrInvalidQueryParameter one of the query parameters has invalid value or format.
func ErrInvalidQueryParameter() *codec.Error {
	return &codec.Error{
		Title:  "An invalid value or format was specified for one of the query parameters.",
		Status: "400",
	}
}

// ErrInvalidResourceName defines an error when the specified resource name is not valid.
func ErrInvalidResourceName() *codec.Error {
	return &codec.Error{
		Title:  "The specified resource name is not valid.",
		Status: "400",
	}
}

// ErrTypeConflict defines an error when the data 'type' doesn't match endpoint's defined 'type'.
func ErrTypeConflict() *codec.Error {
	return &codec.Error{
		Title:  "Provided data 'type' doesn't match endpoint's type",
		Status: "409",
	}
}

// ErrIDConflict defines an error when the primary field value doesn't match endpoint's defined id.
func ErrIDConflict() *codec.Error {
	return &codec.Error{
		Title:  "Provided data 'id' doesn't match endpoint's value",
		Status: "409",
	}
}

// ErrInvalidURI error returned when the URI is not recognized by the server.
func ErrInvalidURI() *codec.Error {
	return &codec.Error{
		Title:  "The requested URI does not represent any resource on the server.",
		Status: "400",
	}
}

// ErrInvalidJSONDocument error returned when the specified JSON structure is not syntactically valid.
func ErrInvalidJSONDocument() *codec.Error {
	return &codec.Error{
		Title:  "The specified JSON is not syntactically valid.",
		Status: "400",
	}
}

// ErrInvalidJSONFieldValue error returned when one or more of the specified JSON fields was not in a correct format.
func ErrInvalidJSONFieldValue() *codec.Error {
	return &codec.Error{
		Title:  "The value provided for one of the JSON fields in the requested body was not in the correct format.",
		Status: "400",
	}
}

// ErrInvalidJSONFieldName error returned when one or more of the specified JSON fields was not in a correct format.
func ErrInvalidJSONFieldName() *codec.Error {
	return &codec.Error{
		Title:  "One of provided model fields doesn't exists",
		Status: "400",
	}
}

// ErrHashMismatch returns when the hash value specified in the request didn't match the one stored/computed by the server.
func ErrHashMismatch() *codec.Error {
	return &codec.Error{
		Title:  "The Hash value specified in the request did not match the value stored/computed by the server.",
		Status: "400",
	}
}

// ErrMetadataTooLarge the size of the specified metadata exceeds the limits.
func ErrMetadataTooLarge() *codec.Error {
	return &codec.Error{
		Title:  "The size of the specified metadata exceeds the maximum size permitted.",
		Status: "400",
	}
}

// ErrMissingRequiredQueryParameter returns when one or more of the query parameter is missing for the request.
func ErrMissingRequiredQueryParameter() *codec.Error {
	return &codec.Error{
		Title:  "A required query parameter was not specified for this request.",
		Status: "400",
	}
}

// ErrMissingRequiredHeader one of the required HTTP headers were not specified in the request.
func ErrMissingRequiredHeader() *codec.Error {
	return &codec.Error{
		Title:  "A required HTTP header was not specified.",
		Status: "400",
	}
}

// ErrMissingRequiredModelField one of the required model fields were not specified in the request body
func ErrMissingRequiredModelField() *codec.Error {
	return &codec.Error{
		Title:  "A required Model field was not specified in the request body.",
		Status: "400",
	}
}

// ErrInputOutOfRange one of the request inputs were out of range.
func ErrInputOutOfRange() *codec.Error {
	return &codec.Error{
		Title:  "One of the request inputs is out of range.",
		Status: "400",
	}
}

// ErrQueryParameterValueOutOfRange one of the specified query parameters in the request URI is outside the permissible range.
func ErrQueryParameterValueOutOfRange() *codec.Error {
	return &codec.Error{
		Title:  "A query parameter specified in the request URI is outside the permissible range.",
		Status: "400",
	}
}

// ErrUnsupportedHeader one of the HTTP headers specified in the request is not supported.
func ErrUnsupportedHeader() *codec.Error {
	return &codec.Error{
		Title:  "One of the HTTP headers specified in the request is not supported.",
		Status: "400",
	}
}

// ErrUnsupportedField one of the fields specified in the request is not supported.
func ErrUnsupportedField() *codec.Error {
	return &codec.Error{
		Title:  "One of the fields specified in the request body is not supported.",
		Status: "400",
	}
}

// ErrUnsupportedQueryParameter one of the query parameters in the request URI is not supported.
func ErrUnsupportedQueryParameter() *codec.Error {
	return &codec.Error{
		Title:  "One of the query parameters in the request URI is not supported.",
		Status: "400",
	}
}

// ErrUnsupportedFilterOperator one of the filter operators is not supported.
func ErrUnsupportedFilterOperator() *codec.Error {
	return &codec.Error{
		Title:  "One of the filter operators is not supported.",
		Status: "400",
	}
}

/**

STATUS 403

*/

// ErrForbidden server understood the request but refuses to authorize it.
func ErrForbiddenAuthorize() *codec.Error {
	return &codec.Error{
		Title:  "The server understood the request but refuses to authorize it.",
		Status: "403",
	}
}

// ErrForbiddenOperation server understood the request but current operation is not allowed.
func ErrForbiddenOperation() *codec.Error {
	return &codec.Error{
		Title:  "The server understood the request but current operation is forbidden.",
		Status: "403",
	}
}

// ErrAccountDisabled provided account is disabled.
func ErrAccountDisabled() *codec.Error {
	return &codec.Error{
		Title:  "The specified account is disabled.",
		Status: "403",
	}
}

// ErrInvalidAuthorizationHeader server failed to authenticate the request due to invalid authorization header.
func ErrInvalidAuthorizationHeader() *codec.Error {
	return &codec.Error{
		Title:  `Server failed to authenticate the request. Make sure the value of Authorization header is formed correctly including the signature.`,
		Status: "403",
	}
}

// ErrInsufficientAccountPermissions provided account has insufficient permissions for given request.
func ErrInsufficientAccountPermissions() *codec.Error {
	return &codec.Error{
		Title:  "The account being accessed does not have sufficient permissions to execute this operation.",
		Status: "403",
	}
}

// ErrInvalidCredentials provided invalid account credentials - access denied.
func ErrInvalidCredentials() *codec.Error {
	return &codec.Error{
		Title:  "Access is denied due to invalid credentials.",
		Status: "403",
	}
}

// ErrEndpointForbidden forbidden access above given API endpoint.
func ErrEndpointForbidden() *codec.Error {
	return &codec.Error{
		Title:  "Provided endpoint is forbidden.",
		Status: "403",
	}
}

/**

STATUS 404

*/

// ErrResourceNotFound provided resource doesn't exists.
func ErrResourceNotFound() *codec.Error {
	return &codec.Error{
		Title:  "The specified resource does not exists.",
		Status: "404",
	}
}

/**

STATUS 405

*/

// ErrMethodNotAllowed given method is not allowed for the specified URI
func ErrMethodNotAllowed() *codec.Error {
	return &codec.Error{
		Title:  "The resource doesn't support the specified HTTP method.",
		Status: "405",
	}
}

/**

STATUS 406

*/

// ErrNotAcceptable one of the header contains values that are not possible to get the response by the server.
func ErrNotAcceptable() *codec.Error {
	return &codec.Error{
		Title:  "The server cannot produce a response matching the list of acceptable values defined in the request's proactive content negotiation headers",
		Status: "406",
	}
}

// ErrLanguageNotAcceptable languages provided in the request are not supported.
func ErrLanguageNotAcceptable() *codec.Error {
	return &codec.Error{
		Title:  "The language provided in the request is not supported.",
		Status: "406",
	}
}

// ErrLanguageHeaderNotAcceptable provided request headers contains not supported language.
func ErrLanguageHeaderNotAcceptable() *codec.Error {
	return &codec.Error{
		Title:  "The language provided in the request header is not supported.",
		Status: "406",
	}
}

/**

STATUS 409

*/

// ErrAccountAlreadyExists creating account failed - user already exists.
func ErrAccountAlreadyExists() *codec.Error {
	return &codec.Error{
		Title:  "The account provided in the request already exists.",
		Status: "409",
	}
}

// ErrResourceAlreadyExists the specified resource already exists.
func ErrResourceAlreadyExists() *codec.Error {
	return &codec.Error{
		Title:  "The specified resource already exists.",
		Status: "409",
	}
}

/**

STATUS 413

*/

// ErrRequestBodyTooLarge the size of the request body exceeds the maximum permitted size.
func ErrRequestBodyTooLarge() *codec.Error {
	return &codec.Error{
		Title:  "The size of the request body exceeds the maximum permitted size.",
		Status: "413",
	}
}

/**

STATUS 500

*/

// ErrInternalError server encountered internal error.
func ErrInternalError() *codec.Error {
	return &codec.Error{
		Title:  "The server encountered an internal error. Please retry the request.",
		Status: "500",
	}
}

// ErrOperationTimedOut the operation could not be completed within the permitted time.
func ErrOperationTimedOut() *codec.Error {
	return &codec.Error{
		Title:  "The operation could not be completed within the permitted time.",
		Status: "500",
	}
}

/**

STATUS 503

*/

// ErrServiceUnavailable the server is currently unable to receive requests.
func ErrServiceUnavailable() *codec.Error {
	return &codec.Error{
		Title:  "The server is currently unable to receive requests. Please retry your request.",
		Status: "503",
	}
}

// ErrTooManyOperationsPerAccount too many requests for given account.
func ErrTooManyOperationsPerAccount() *codec.Error {
	return &codec.Error{
		Title:  "There were too many requests allowed for the given account.",
		Status: "503",
	}
}
