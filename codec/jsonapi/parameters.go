package jsonapi

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	neuronCodec "github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// ParseParameters implements neuronCodec.ParametersParser interface.
func (c Codec) ParseParameters(ctrl *controller.Controller, q *query.Scope, parameters query.Parameters) (err error) {
	var (
		includes             query.Parameter
		pageSize, pageNumber int64
		hasLimitOffset       bool
	)
	fields := map[*mapping.ModelStruct]mapping.FieldSet{}

	for _, parameter := range parameters {
		switch {
		case parameter.Key == query.ParamPageLimit:
			if err := c.parseLimit(parameter, q); err != nil {
				return err
			}
			hasLimitOffset = true
		case parameter.Key == query.ParamPageOffset:
			if err := c.parseOffset(parameter, q); err != nil {
				return err
			}
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
		case strings.HasPrefix(parameter.Key, filter.ParamFilter):
			split, err := query.SplitBracketParameter(parameter.Key[len(filter.ParamFilter):])
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
			q.Filter(ff)
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
			q.StoreSet(StoreKeyMarshalLinks, marshalLinksValue)
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
	if fieldSet, ok := fields[q.ModelStruct]; ok {
		q.FieldSets = []mapping.FieldSet{fieldSet}
	}
	return nil
}

func (c Codec) parseOffset(parameter query.Parameter, q *query.Scope) error {
	offset, err := parameter.Int64()
	if err != nil {
		return err
	}
	q.Offset(offset)
	return nil
}

func (c Codec) parseLimit(parameter query.Parameter, q *query.Scope) error {
	limit, err := parameter.Int64()
	if err != nil {
		return err
	}
	q.Limit(limit)
	return nil
}

func (c Codec) parsePageBasedPagination(pageSize int64, pageNumber int64, hasLimitOffset bool) (query.Pagination, error) {
	if pageNumber == 0 {
		pageNumber = 1
	}
	if pageSize == 0 {
		return query.Pagination{}, errors.NewDetf(query.ClassInvalidParameter, "provided invalid pagination").
			WithDetail(fmt.Sprintf("Page size parameter: '%s' not defined.", ParamPageSize))
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

func (c Codec) parseFilterParameter(mStruct *mapping.ModelStruct, split []string, parameter query.Parameter) (filter.Filter, error) {
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
		o, ok := filter.Operators.Get(split[1])
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
		f := filter.New(sField, o, values...)
		return f, nil
	case mapping.KindAttribute:
		if len(split) != 2 {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s'", parameter.Key)
		}
		o, ok := filter.Operators.Get(split[1])
		if !ok {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "provided invalid query operator: %s", split[1])
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return nil, errors.NewDetf(neuronCodec.ClassInternal, "provided model is not a mapping.Fielder")
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
		return filter.New(sField, o, values...), nil
	case mapping.KindRelationshipSingle, mapping.KindRelationshipMultiple:
		if len(split) == 1 {
			return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s'", parameter.Key)
		}
		sub, err := c.parseFilterParameter(sField.Relationship().RelatedModelStruct(), split[1:], parameter)
		if err != nil {
			return nil, err
		}
		return filter.Relation{StructField: sField, Nested: []filter.Filter{sub}}, nil
	default:
		return nil, errors.NewDetf(query.ClassInvalidParameter, "invalid filter parameter: '%s' - invalid filter fields", parameter.Key)
	}
}

func (c Codec) parseFieldsParameter(ctrl *controller.Controller, q *query.Scope, parameter query.Parameter, fields map[*mapping.ModelStruct]mapping.FieldSet) error {
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
		sField, ok := model.StructFieldByName(field)
		if !ok || sField.IsHidden() {
			return errors.Newf(query.ClassInvalidParameter, "field: '%s' not found for the model", field)
		}
		switch sField.Kind() {
		case mapping.KindAttribute, mapping.KindRelationshipSingle, mapping.KindRelationshipMultiple:
			if fs.Contains(sField) {
				return errors.Newf(query.ClassInvalidParameter, "duplicated field '%s' in '%s' parameter", field, query.ParamFields)
			}
			fs = append(fs, sField)
		case mapping.KindPrimary:
			return errors.NewDet(query.ClassInvalidParameter, "cannot set 'id' field in the query fieldset")
		default:
			return errors.Newf(query.ClassInvalidParameter, "field: '%s' not found for the model", field)
		}
	}
	fields[model] = fs
	return nil
}

func (c Codec) parseIncludesParameter(q *query.Scope, parameter query.Parameter, fields map[*mapping.ModelStruct]mapping.FieldSet) error {
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

func (c Codec) addIncludedParameter(mStruct *mapping.ModelStruct, parameters []string, included []*query.IncludedRelation, fields map[*mapping.ModelStruct]mapping.FieldSet) (*query.IncludedRelation, error) {
	field := parameters[0]
	sField, ok := mStruct.RelationByName(field)
	if !ok || sField.IsHidden() {
		return nil, errors.Newf(query.ClassInvalidParameter, "relation: '%s' not found for the model", field)
	}

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
		ir, err := c.addIncludedParameter(sField.Relationship().RelatedModelStruct(), parameters[1:], includedRelation.IncludedRelations, fields)
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
}

func FormatPagination(p *query.Pagination, temp url.Values, pageBased bool) {
	var k, v string

	if !pageBased {
		limit, offset := p.Limit, p.Offset
		if limit != 0 {
			k = query.ParamPageLimit
			v = strconv.FormatInt(limit, 10)
			temp.Set(k, v)
		}
		if offset != 0 {
			k = query.ParamPageOffset
			v = strconv.FormatInt(offset, 10)
			temp.Set(k, v)
		}
	} else {
		pageNumber := (p.Offset + p.Limit) / p.Limit
		temp.Set(ParamPageNumber, strconv.FormatInt(pageNumber, 10))
		temp.Set(ParamPageSize, strconv.FormatInt(p.Limit, 10))
	}
}
