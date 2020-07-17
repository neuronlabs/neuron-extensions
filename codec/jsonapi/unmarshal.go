package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

type ctxKey struct{}

var (
	marshalFields = &ctxKey{}
	marshalLinks  = &ctxKey{}
)

// ParamLinks defines the links parameter name.
const (
	ParamLinks      = "links"
	ParamPageSize   = "page[size]"
	ParamPageNumber = "page[number]"
)

// UnmarshalModels implements codec.Codec interface.
func (c *Codec) UnmarshalModels(r io.Reader, modelStruct *mapping.ModelStruct, options *codec.UnmarshalOptions) ([]mapping.Model, error) {
	payloader, err := unmarshalPayload(r, options)
	if err != nil {
		return nil, err
	}

	var includes map[string]*Node
	if len(payloader.GetIncluded()) != 0 {
		includes = map[string]*Node{}
		for _, included := range payloader.GetIncluded() {
			includes[includedKeyFunc(included)] = included
		}
	}

	var models []mapping.Model
	for _, node := range payloader.GetNodes() {
		model := mapping.NewModel(modelStruct)
		if err = unmarshalNode(modelStruct, node, model, includes, options); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// UnmarshalQuery implements codec.QueryUnmarshaler interface.
func (c *Codec) UnmarshalQuery(r io.Reader, modelStruct *mapping.ModelStruct, options *codec.UnmarshalOptions) (*query.Scope, error) {
	payload, err := unmarshalPayload(r, options)
	if err != nil {
		return nil, err
	}

	// TODO: set metadata to the query scope.
	// TODO: set included relations from the options.
	var includes map[string]*Node
	if len(payload.GetIncluded()) != 0 {
		includes = map[string]*Node{}
		for _, included := range payload.GetIncluded() {
			includes[includedKeyFunc(included)] = included
		}
	}

	var models []mapping.Model
	for _, node := range payload.GetNodes() {
		model := mapping.NewModel(modelStruct)
		if err = unmarshalNode(modelStruct, node, model, includes, options); err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	q := query.NewScope(modelStruct, models...)
	return q, nil
}

// ParseParameters implements codec.ParametersParser interface.
func (c *Codec) ParseParameters(ctrl *controller.Controller, q *query.Scope, parameters query.Parameters) (err error) {
	var (
		includes             query.Parameter
		pageSize, pageNumber int64
		hasLimitOffset       bool
	)
	fields := map[*mapping.ModelStruct]mapping.FieldSet{}

	for _, parameter := range parameters {
		switch {
		case parameter.Key == query.ParamPageLimit:
			limit, err := parameter.Int64()
			if err != nil {
				return err
			}
			q.Limit(limit)
			hasLimitOffset = true
		case parameter.Key == query.ParamPageOffset:
			offset, err := parameter.Int64()
			if err != nil {
				return err
			}
			q.Offset(offset)
			hasLimitOffset = true
		case parameter.Key == ParamPageSize:
			pageSize, err = parameter.Int64()
			if err != nil {
				return err
			}
			if pageSize <= 0 {
				return errors.NewDetf(query.ClassInvalidParameter, "invalid %s parameter value", parameter.Key).WithDetail("page number cannot be lower or equal to 0")
			}
		case parameter.Key == ParamPageNumber:
			pageNumber, err = parameter.Int64()
			if err != nil {
				return err
			}
			if pageNumber <= 0 {
				return errors.NewDetf(query.ClassInvalidParameter, "invalid %s parameter value", parameter.Key).WithDetail("page number cannot be lower than 0")
			}
		case parameter.Key == query.ParamInclude:
			includes = parameter
		case strings.HasPrefix(parameter.Key, query.ParamFields):
			if err := c.parseFieldsParameter(ctrl, q, parameter, fields); err != nil {
				return err
			}
		case strings.HasPrefix(parameter.Key, query.ParamFilter):
			split, err := query.SplitBracketParameter(parameter.Key[len(query.ParamFilter):])
			if err != nil {
				return err
			}
			if len(split) == 1 {
				return errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s'", parameter.Key)
			}
			mStruct := q.ModelStruct
			ff, err := c.parseFilterParameter(mStruct, split, parameter)
			if err != nil {
				return err
			}
			if err := q.Filter(ff); err != nil {
				return err
			}
		case parameter.Key == query.ParamSort:
			sortFields := parameter.StringSlice()
			for _, sortField := range sortFields {
				if err := q.OrderBy(sortField); err != nil {
					return err
				}
			}
		case parameter.Key == ParamLinks:
			marshalLinksValue := true
			if parameter.Value != "" {
				v, err := parameter.Boolean()
				if err != nil {
					return err
				}
				marshalLinksValue = v
			}
			q.StoreSet(marshalLinks, marshalLinksValue)
		default:
			// TODO: provide a way to use custom query parameters - for registered key values.
			return errors.NewDetf(query.ClassInvalidParameter, "provided invalid query parameter: %s", parameter.Key)
		}
	}

	if pageSize != 0 || pageNumber != 0 {
		p, err := c.parsePageBasedPagination(pageSize, pageNumber, hasLimitOffset)
		if err != nil {
			return err
		}
		q.Pagination = &p
	}

	if includes != (query.Parameter{}) {
		if err := c.parseIncludesParameter(q, includes, fields); err != nil {
			return err
		}
	}
	if len(fields) != 0 {
		// Store the fields for given query - could be used later.
		q.StoreSet(marshalFields, fields)
	}
	return nil
}

func (c *Codec) parsePageBasedPagination(pageSize int64, pageNumber int64, hasLimitOffset bool) (query.Pagination, error) {
	if pageSize == 0 || pageNumber == 0 {
		return query.Pagination{}, errors.NewDetf(query.ClassInvalidParameter, "provided invalid pagination").
			WithDetail(fmt.Sprintf("Both values '%s' and '%s' must be set.", ParamPageNumber, ParamPageSize))
	}
	if hasLimitOffset {
		return query.Pagination{}, errors.NewDetf(query.ClassInvalidParameter, "provided invalid pagination").
			WithDetail(fmt.Sprintf("Cannot use both page and limit/offset based pagination at the same time."))
	}
	return query.Pagination{
		Limit:  pageSize,
		Offset: (pageNumber - 1) * pageSize,
	}, nil
}

func (c *Codec) parseFilterParameter(mStruct *mapping.ModelStruct, split []string, parameter query.Parameter) (*query.FilterField, error) {
	sField, ok := mStruct.FieldByName(split[0])
	if !ok || sField.IsHidden() {
		return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s' - invalid fields", parameter.Key)
	}
	model := mapping.NewModel(mStruct)
	switch sField.Kind() {
	case mapping.KindPrimary:
		if len(split) != 2 {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s'", parameter.Key)
		}
		o, ok := query.FilterOperators.Get(split[1])
		if !ok {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "provided invalid query operator: %s", split[1])
		}
		var values []interface{}
		// Parse filter values.
		if o.IsRangeable() {
			stringValues := strings.Split(parameter.Value, ",")
			for _, stringValue := range stringValues {
				if err := model.SetPrimaryKeyStringValue(stringValue); err != nil {
					return nil, err
				}
				values = append(values, model.GetPrimaryKeyValue())
			}
		} else {
			if err := model.SetPrimaryKeyStringValue(parameter.Value); err != nil {
				return nil, err
			}
			values = append(values, model.GetPrimaryKeyValue())
		}
		return query.NewFilterField(sField, o, values...), nil
	case mapping.KindAttribute:
		if len(split) != 2 {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s'", parameter.Key)
		}
		o, ok := query.FilterOperators.Get(split[1])
		if !ok {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "provided invalid query operator: %s", split[1])
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return nil, errors.NewDetf(codec.ClassInternal, "provided model is not a mapping.Fielder")
		}
		var values []interface{}
		// Parse attribute filter values.
		switch {
		case o.IsRangeable():
			stringValues := strings.Split(parameter.Value, ",")
			for _, stringValue := range stringValues {
				value, err := fielder.ParseFieldsStringValue(sField, stringValue)
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
		case o.IsStringOnly():
			values = append(values, parameter.Value)
		default:
			value, err := fielder.ParseFieldsStringValue(sField, parameter.Value)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return query.NewFilterField(sField, o, values...), nil
	case mapping.KindRelationshipSingle, mapping.KindRelationshipMultiple:
		if len(split) == 1 {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s'", parameter.Key)
		}
		sub, err := c.parseFilterParameter(sField.Relationship().Struct(), split[1:], parameter)
		if err != nil {
			return nil, err
		}
		return &query.FilterField{
			StructField: sField,
			Nested:      []*query.FilterField{sub},
		}, nil
	default:
		return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s' - invalid filter fields", parameter.Key)
	}
}

func (c *Codec) parseFieldsParameter(ctrl *controller.Controller, q *query.Scope, parameter query.Parameter, fields map[*mapping.ModelStruct]mapping.FieldSet) error {
	split, err := query.SplitBracketParameter(parameter.Key[len(query.ParamFields):])
	if err != nil {
		return err
	}

	if len(split) != 1 {
		err := errors.NewDetf(query.ClassInvalidParameter, "invalid fields parameter")
		err.Details = fmt.Sprintf("The fields parameter has invalid form. %s", parameter.Key)
		return err
	}
	model, ok := ctrl.ModelMap.GetByCollection(split[0])
	if !ok {
		if log.CurrentLevel() == log.LevelDebug3 {
			log.Debug3f("[%s] invalid fieldset model: '%s'", q.ID, split[0])
		}
		err := errors.NewDetf(query.ClassInvalidParameter, "invalid query parameter")
		err.Details = fmt.Sprintf("Fields query parameter contains invalid collection name: '%s'", split[0])
		return err
	}
	fs := mapping.FieldSet{}
	for _, field := range parameter.StringSlice() {
		sField, ok := model.FieldByName(field)
		if !ok || sField.IsHidden() {
			return errors.Newf(query.ClassInvalidParameter, "field: '%s' not found for the model", field)
		}
		switch sField.Kind() {
		case mapping.KindAttribute, mapping.KindRelationshipSingle, mapping.KindRelationshipMultiple:
			if fs.Contains(sField) {
				return errors.Newf(query.ClassInvalidParameter, "duplicated field '%s' in '%s' parameter", field, query.ParamFields)
			}
			fs = append(fs, sField)
		default:
			return errors.Newf(query.ClassInvalidParameter, "field: '%s' not found for the model", field)
		}
	}
	fields[model] = fs
	return nil
}

func (c *Codec) parseIncludesParameter(q *query.Scope, parameter query.Parameter, fields map[*mapping.ModelStruct]mapping.FieldSet) error {
	for _, field := range parameter.StringSlice() {
		included := strings.Split(field, ".")

		ir, err := c.addIncludedParameter(q.ModelStruct, included, q.IncludedRelations, fields)
		if err != nil {
			return err
		}
		if ir != nil {
			q.IncludedRelations = append(q.IncludedRelations, ir)
		}
	}
	return nil
}

func (c *Codec) addIncludedParameter(mStruct *mapping.ModelStruct, parameters []string, included []*query.IncludedRelation, fields map[*mapping.ModelStruct]mapping.FieldSet) (*query.IncludedRelation, error) {
	field := parameters[0]
	sField, ok := mStruct.FieldByName(field)
	if !ok || sField.IsHidden() {
		return nil, errors.Newf(query.ClassInvalidParameter, "field: '%s' not found for the model", field)
	}
	switch sField.Kind() {
	case mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
		var includedRelation *query.IncludedRelation
		appendTop := true
		for _, relation := range included {
			if relation.StructField == sField {
				includedRelation = relation
				appendTop = false
				break
			}
		}
		if includedRelation == nil {
			includedRelation = c.newIncludedRelation(sField, fields)
		}
		if len(parameters) > 1 {
			ir, err := c.addIncludedParameter(sField.Relationship().Struct(), parameters[1:], includedRelation.IncludedRelations, fields)
			if err != nil {
				return nil, err
			}
			if ir != nil {
				includedRelation.IncludedRelations = append(includedRelation.IncludedRelations, ir)
			}
		}
		if appendTop {
			return includedRelation, nil
		}
		return nil, nil
	default:
		return nil, errors.Newf(query.ClassInvalidParameter, "field: '%s' not found for the model", field)
	}
}

func (c *Codec) newIncludedRelation(sField *mapping.StructField, fields map[*mapping.ModelStruct]mapping.FieldSet) (includedRelation *query.IncludedRelation) {
	includedRelation = &query.IncludedRelation{StructField: sField}
	fs, ok := fields[sField.ModelStruct()]
	if ok {
		includedRelation.Fieldset = append(includedRelation.Fieldset, sField.Relationship().Struct().Primary())
		for _, field := range fs {
			switch field.Kind() {
			case mapping.KindAttribute:
				includedRelation.Fieldset = append(includedRelation.Fieldset, field)
			case mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
				includedRelation.IncludedRelations = append(includedRelation.IncludedRelations, c.newIncludedRelation(field, fields))
			}
		}
	} else {
		// By default set full fieldset and all possible relationships.
		includedRelation.Fieldset = sField.ModelStruct().Fields()
		for _, relation := range sField.ModelStruct().RelationFields() {
			includedRelation.IncludedRelations = append(includedRelation.IncludedRelations, c.newIncludedRelation(relation, fields))
		}
	}

	return includedRelation
}

func unmarshalPayload(in io.Reader, options *codec.UnmarshalOptions) (Payloader, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	t, err := dec.Token()
	if err != nil {
		return nil, err
	}

	if t != json.Delim('{') {
		return nil, errors.New(codec.ClassUnmarshalDocument, "invalid document")
	}
	if err := unmarshalPayloadFindData(dec); err != nil {
		return nil, errors.New(codec.ClassUnmarshalDocument, "invalid input")
	}

	t, err = dec.Token()
	if err != nil {
		return nil, err
	}
	var payloader Payloader
	switch t {
	case json.Delim('{'):
		payloader = &SinglePayload{}
	case json.Delim('['):
		payloader = &ManyPayload{}
	default:
		return nil, errors.New(codec.ClassUnmarshalDocument, "invalid input")
	}
	r.Seek(0, io.SeekStart)
	dec = json.NewDecoder(r)
	if options != nil && options.StrictUnmarshal {
		dec.DisallowUnknownFields()
	}
	if err = dec.Decode(payloader); err != nil {
		return nil, err
	}
	return payloader, nil
}

func unmarshalPayloadFindData(dec *json.Decoder) (err error) {
	var (
		openBracketCount int
		t                json.Token
	)
	for {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		switch tk := t.(type) {
		case string:
			if tk == "data" && openBracketCount == 0 {
				return nil
			}
		case json.Delim:
			switch tk {
			case json.Delim('{'):
				openBracketCount++
			case json.Delim('}'):
				openBracketCount--
			}
		}
	}
}

// func unmarshalHandleDecodeError(err error) error {
// 	// handle the incoming error
// 	switch e := err.(type) {
// 	case errors.DetailedError:
// 		return err
// 	case *json.SyntaxError:
// 		err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "syntax error")
// 		err.WithDetailf("Document syntax error: '%s'. At data offset: '%d'", e.Error(), e.Offset)
// 		return err
// 	case *json.UnmarshalTypeError:
// 		if e.Type == reflect.TypeOf(SinglePayload{}) || e.Type == reflect.TypeOf(ManyPayload{}) {
// 			err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "invalid jsonapi document syntax")
// 			return err
// 		}
// 		err := errors.NewDet(class.EncodingUnmarshalInvalidType, "invalid field type")
//
// 		var fieldType string
// 		switch e.Field {
// 		case "id", "type", "client-id":
// 			fieldType = e.Type.String()
// 		case "relationships", "attributes", "links", "meta":
// 			fieldType = "object"
// 		}
// 		err.WithDetailf("Invalid type for: '%s' field. Required type '%s' but is: '%v'", e.Field, fieldType, e.Value)
// 		return err
// 	default:
// 		if e == io.EOF || e == io.ErrUnexpectedEOF {
// 			err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "reader io.EOF occurred")
// 			err.WithDetailf("invalid document syntax")
// 			return err
// 		}
// 		err := errors.NewDetf(class.EncodingUnmarshal, "unknown unmarshal error: %s", e.Error())
// 		return err
// 	}
// }

func unmarshalNode(mStruct *mapping.ModelStruct, data *Node, model mapping.Model, included map[string]*Node, options *codec.UnmarshalOptions) error {
	if data.Type != model.NeuronCollectionName() {
		err := errors.NewDet(codec.ClassUnmarshal, "unmarshal collection name doesn't match the root struct")
		err.Details = fmt.Sprintf("unmarshal collection: '%s' doesn't match root collection:'%s'", data.Type, model.NeuronCollectionName())
		return err
	}
	// Set primary key value.
	if data.ID != "" {
		if err := model.SetPrimaryKeyStringValue(data.ID); err != nil {
			return err
		}
	}

	// Set attributes.
	if data.Attributes != nil {
		fielder, isFielder := model.(mapping.Fielder)
		if !isFielder {
			if len(mStruct.Attributes()) > 0 {
				return errors.New(codec.ClassInternal, "provided model is not a Fielder")
			} else if options != nil && options.StrictUnmarshal {
				return errors.New(codec.ClassUnmarshal, "provided model doesn't have any attributes")
			}
		} else {
			// Iterate over the data attributes
			for attrName, attrValue := range data.Attributes {
				var isHidden bool
				modelAttr, ok := mStruct.Attribute(attrName)
				if ok {
					isHidden = modelAttr.IsHidden()
					for _, tag := range modelAttr.ExtractFieldTags(StructTag) {
						if tag.Key == "-" {
							isHidden = true
							break
						}
					}
				} else {
					// try to find the field by the jsonapi struct tag
					var fieldName string
					for _, field := range mStruct.Attributes() {
						for _, tag := range field.ExtractFieldTags(StructTag) {
							if tag.Key == "-" || tag.Key == "omitempty" {
								continue
							}
							fieldName = tag.Key
							break
						}
						if fieldName == attrName {
							modelAttr = field
							ok = true
							break
						}
					}
				}

				if !ok || (ok && isHidden) {
					if options.StrictUnmarshal {
						err := errors.NewDet(codec.ClassUnmarshal, "unknown field name")
						err.Details = fmt.Sprintf("provided unknown field name: '%s', for the collection: '%s'.", attrName, data.Type)
						return err
					}
					continue
				}
				if err := fielder.SetFieldValue(modelAttr, attrValue); err != nil {
					return err
				}
			}
		}
	}

	if data.Relationships != nil {
		for relName, relValue := range data.Relationships {
			var isHidden bool
			relationshipStructField, ok := mStruct.RelationByName(relName)
			if ok {
				isHidden = relationshipStructField.IsHidden()
				for _, tag := range relationshipStructField.ExtractFieldTags(StructTag) {
					if tag.Key == "-" {
						isHidden = true
						break
					}
				}
			}
			// Try to find the field by the jsonapi struct tag.
			if !ok {
				var fieldName string
				for _, field := range mStruct.RelationFields() {
					for _, tag := range relationshipStructField.ExtractFieldTags(StructTag) {
						if tag.Key == "-" || tag.Key == "omitempty" {
							continue
						}
						fieldName = tag.Key
						break
					}
					if fieldName == relName {
						relationshipStructField = field
						ok = true
						break
					}
				}
			}

			if !ok || (ok && isHidden) {
				if options != nil && options.StrictUnmarshal {
					err := errors.NewDet(codec.ClassUnmarshal, "unknown field name")
					err.Details = fmt.Sprintf("Provided unknown field name: '%s', for the collection: '%s'.", relName, data.Type)
					return err
				}
				continue
			}

			if relationshipStructField.Kind() == mapping.KindRelationshipMultiple {
				mr, ok := model.(mapping.MultiRelationer)
				if !ok {
					return errors.New(codec.ClassInternal, "model is not a multi relationer")
				}
				// to-many relationship
				relationship := new(RelationshipManyNode)

				buf := bytes.NewBuffer(nil)
				if err := json.NewEncoder(buf).Encode(data.Relationships[relName]); err != nil {
					log.Debug2f("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					err := errors.NewDet(codec.ClassUnmarshal, "invalid relationship format")
					err.Details = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				if err := json.NewDecoder(buf).Decode(relationship); err != nil {
					log.Debug2f("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					err := errors.NewDet(codec.ClassUnmarshal, "invalid relationship format")
					err.Details = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				relStruct := relationshipStructField.Relationship().Struct()
				for _, n := range relationship.Data {
					relModel := mapping.NewModel(relStruct)
					if err := unmarshalNode(relStruct, fullNode(n, included), relModel, included, options); err != nil {
						log.Debug2f("unmarshalNode.RelationshipMany - unmarshalNode failed. %v", err)
						return err
					}
					if err := mr.AddRelationModel(relationshipStructField, relModel); err != nil {
						return err
					}
				}
			} else if relationshipStructField.Kind() == mapping.KindRelationshipSingle {
				sr, ok := model.(mapping.SingleRelationer)
				if !ok {
					return errors.New(codec.ClassInternal, "provided model is not a single relationer")
				}
				relationship := new(RelationshipOneNode)
				buf := bytes.NewBuffer(nil)

				if err := json.NewEncoder(buf).Encode(relValue); err != nil {
					log.Debug2f("Controller.UnmarshalNode.relationshipSingle json.Encode failed. %v", err)
					err := errors.NewDet(codec.ClassUnmarshal, "invalid relationship format")
					err.Details = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				if err := json.NewDecoder(buf).Decode(relationship); err != nil {
					log.Debug2f("Controller.UnmarshalNode.RelationshipSingel json.Decode failed. %v", err)
					err := errors.NewDet(codec.ClassUnmarshal, "invalid relationship format")
					err.Details = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				if relationship.Data == nil {
					continue
				}

				relStruct := relationshipStructField.Relationship().Struct()
				relModel := mapping.NewModel(relStruct)

				if err := unmarshalNode(relStruct, fullNode(relationship.Data, included), relModel, included, options); err != nil {
					return err
				}
				if err := sr.SetRelationModel(relationshipStructField, relModel); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func fullNode(n *Node, included map[string]*Node) *Node {
	if included == nil {
		return n
	}
	includedKey := includedKeyFunc(n)
	if in, ok := included[includedKey]; ok {
		return in
	}
	return n
}

func includedKeyFunc(n *Node) string {
	return fmt.Sprintf("%s,%s", n.Type, n.ID)
}
