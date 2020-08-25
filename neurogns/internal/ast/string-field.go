package ast

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"

	"github.com/neuronlabs/neuron-extensions/neurogns/input"
	"github.com/neuronlabs/neuron-extensions/neurogns/internal/tempfuncs"
)

func (g *ModelGenerator) setFieldsStringGetters(model *input.Model) error {
	for _, field := range model.Fields {
		if err := g.setFieldStringGetter(model, field); err != nil {
			return err
		}
	}
	return nil
}

func (g *ModelGenerator) setFieldStringGetter(model *input.Model, field *input.Field) error {
	if !isTypeBasic(field.Ast.Type) {
		fieldType := field.Type
		if i := strings.IndexRune(fieldType, '*'); i != -1 {
			fieldType = fieldType[i+1:]
		}
		if i := strings.IndexRune(field.Type, '.'); i == -1 {
			fieldType = model.PackageName + "." + fieldType
		}

		for _, method := range g.typeMethods[fieldType] {
			if method.Name == "MarshalText" && len(method.ParameterTypes) == 0 && len(method.ReturnTypes) == 2 &&
				method.ReturnTypes[0] == "[]byte" && method.ReturnTypes[1] == "error" {
				field.IsTextMarshaler = true
				log.Debug2f("Model: '%s' Field: '%s' StringGetter: 'TextMarshaler'\n", model.Name, field.Name)
				return nil
			}
		}

		for _, method := range g.typeMethods[fieldType] {
			if method.Name == "String" && len(method.ParameterTypes) == 0 && len(method.ReturnTypes) == 1 &&
				method.ReturnTypes[0] == "string" {
				field.StringGetter = tempfuncs.StringerFInterface
				log.Debug2f("Model: '%s' Field: '%s' StringGetter: 'Stringer'\n", model.Name, field.Name)
				return nil
			}
		}
	}

	var (
		stringGetter string
		err          error
	)
	if field == model.Primary {
		stringGetter, err = g.getPrimaryStringerGetterFunction(field.Ast.Type)
		if err != nil {
			return errors.Wrapf(err, "model's: '%s' field: '%s' doesn't have string getter function", model.Name, field.Name)
		}
	} else {
		stringGetter, err = g.getFieldStringerGetterFunction(field.Ast.Type)
		if err != nil {
			log.Debugf("Model's: '%s' Field: '%s' doesn't have string getter function. Err: %s\n", model.Name, field.Name, err)
			return nil
		}
	}

	_, ok := tempfuncs.StringerFuncs[stringGetter]
	if !ok {
		return fmt.Errorf("stringer function: '%s' not found", stringGetter)
	}
	log.Debug2f("Model: '%s' Field: '%s' StringGetter: '%s'\n", model.Name, field.Name, stringGetter)

	field.StringGetter = stringGetter
	return nil
}

func (g *ModelGenerator) getPrimaryStringerGetterFunction(expr ast.Expr) (string, error) {
	switch x := expr.(type) {
	case *ast.Ident:
		return g.getIdentStringGetterFunction(x)
	case *ast.SelectorExpr:
		_, ok := x.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("unknown selector value: %v", x)
		}
		return g.getIdentStringGetterFunction(x.Sel)
	case *ast.StarExpr:
		return g.getPrimaryStringerGetterFunction(x.X)
	default:
		return "", fmt.Errorf("unknown primary field type: %+v", x)
	}
}

func (g *ModelGenerator) getFieldStringerGetterFunction(expr ast.Expr) (string, error) {
	switch x := expr.(type) {
	case *ast.Ident:
		return g.getIdentStringGetterFunction(x)
	case *ast.SelectorExpr:
		_, ok := x.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("unknown selector value: %v", x)
		}
		return g.getIdentStringGetterFunction(x.Sel)
	case *ast.StarExpr:
		return g.getFieldStringerGetterFunction(x.X)
	case *ast.ArrayType:
		return g.getFieldStringerGetterFunction(x.Elt)
	default:
		return "", fmt.Errorf("unknown primary field type: %+v", x)
	}
}

func (g *ModelGenerator) getIdentStringGetterFunction(ident *ast.Ident) (string, error) {
	if ident.Obj != nil {
		ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return "", errors.New("unknown type spec object")
		}
		return g.getFieldStringerGetterFunction(ts.Type)
	}
	switch ident.Name {
	case kindInt:
		return tempfuncs.StringerFormatInt, nil
	case kindInt8:
		return tempfuncs.StringerFormatInt8, nil
	case kindInt16:
		return tempfuncs.StringerFormatInt16, nil
	case kindInt32:
		return tempfuncs.StringerFormatInt32, nil
	case kindInt64:
		return tempfuncs.StringerFormatInt64, nil
	case kindUint:
		return tempfuncs.StringerFormatUint, nil
	case kindUint8:
		return tempfuncs.StringerFormatUint8, nil
	case kindUint16:
		return tempfuncs.StringerFormatUint16, nil
	case kindUint32:
		return tempfuncs.StringerFormatUint32, nil
	case kindUint64:
		return tempfuncs.StringerFormatUint64, nil
	case kindFloat32:
		return tempfuncs.StringerFormatFloat32, nil
	case kindFloat64:
		return tempfuncs.StringerFormatFloat64, nil
	case kindString:
		return tempfuncs.StringerString, nil
	case kindBool:
		return tempfuncs.StringerFormatBoolean, nil
	}
	return "", fmt.Errorf("no stringer function found for given field: %s", ident.Name)
}

func (g *ModelGenerator) setFieldStringParserFunctions(model *input.Model) error {
	for _, field := range model.Fields {
		if err := g.setFieldStringParser(model, field); err != nil {
			return err
		}
	}
	return nil
}

func (g *ModelGenerator) setFieldStringParser(model *input.Model, field *input.Field) error {
	log.Debug2f("Model: '%s' Field: '%s' setting string parser\n", model.Name, field.Name)
	if !isTypeBasic(field.Ast.Type) {
		fieldType := field.Type
		if i := strings.IndexRune(fieldType, '*'); i != -1 {
			fieldType = fieldType[i+1:]
		}
		if i := strings.IndexRune(field.Type, '.'); i == -1 {
			fieldType = model.PackageName + "." + fieldType
		}

		for _, method := range g.typeMethods[fieldType] {
			if method.Name == "UnmarshalText" {
				if len(method.ParameterTypes) == 1 && method.ParameterTypes[0] == "[]byte" && len(method.ReturnTypes) == 1 && method.ReturnTypes[0] == "error" {
					field.IsTextUnmarshaler = true
					log.Debug2f("Field: '%s' StringParser TextUnmarshaler'\n", field.Name)
					return nil
				}
			}
		}
	}
	var (
		stringParser string
		err          error
	)
	if field == model.Primary {
		stringParser, err = g.getPrimaryStringParserFunction(field.Ast.Type)
		if err != nil {
			return errors.Wrapf(err, "model's: '%s' field: '%s' doesn't have string parser function", model.Name, field.Name)
		}
	} else {
		stringParser, err = g.getAttributeStringParserFunction(field.Ast.Type)
		if err != nil {
			log.Debug2f("[WARNING] Model's: '%s' field '%s' don't have string parser function.\n", model.Name, field.Name)
			return nil
		}
	}
	if stringParser == tempfuncs.DummyStringParser {
		field.IsString = true
		return nil
	}
	_, ok := tempfuncs.Parsers[stringParser]
	if !ok {
		return fmt.Errorf("string parser function: '%s' not found", stringParser)
	}
	log.Debug2f("Field: '%s' StringParser '%s'\n", field.Name, stringParser)
	field.StringSetter = stringParser
	return nil
}

func (g *ModelGenerator) getPrimaryStringParserFunction(expr ast.Expr) (string, error) {
	switch x := expr.(type) {
	case *ast.Ident:
		return g.getIdentStringParserFunction(x)
	case *ast.SelectorExpr:
		_, ok := x.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("unknown selector value: %v", x)
		}
		return g.getIdentStringParserFunction(x.Sel)
	case *ast.StarExpr:
		return g.getPrimaryStringParserFunction(x.X)
	default:
		return "", fmt.Errorf("unknown primary field type: %+v", x)
	}
}

func (g *ModelGenerator) getAttributeStringParserFunction(expr ast.Expr) (string, error) {
	switch x := expr.(type) {
	case *ast.Ident:
		return g.getIdentStringParserFunction(x)
	case *ast.SelectorExpr:
		_, ok := x.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("unknown selector value: %v", x)
		}
		return g.getIdentStringParserFunction(x.Sel)
	case *ast.StarExpr:
		return g.getAttributeStringParserFunction(x.X)
	case *ast.ArrayType:
		return g.getAttributeStringParserFunction(x.Elt)
	default:
		return "", fmt.Errorf("unknown primary field type: %+v", x)
	}
}

func (g *ModelGenerator) getIdentStringParserFunction(ident *ast.Ident) (string, error) {
	if ident.Obj != nil {
		ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return "", errors.New("unknown type spec object")
		}
		return g.getAttributeStringParserFunction(ts.Type)
	}
	switch ident.Name {
	case kindInt:
		return tempfuncs.ParserInt, nil
	case kindInt8:
		return tempfuncs.ParserInt8, nil
	case kindInt16:
		return tempfuncs.ParserInt16, nil
	case kindInt32:
		return tempfuncs.ParserInt32, nil
	case kindInt64:
		return tempfuncs.ParserInt64, nil
	case kindUint:
		return tempfuncs.ParserUint, nil
	case kindUint8:
		return tempfuncs.ParserUint8, nil
	case kindUint16:
		return tempfuncs.ParserUint16, nil
	case kindUint32:
		return tempfuncs.ParserUint32, nil
	case kindUint64:
		return tempfuncs.ParserUint64, nil
	case kindFloat32:
		return tempfuncs.ParserFloat32, nil
	case kindFloat64:
		return tempfuncs.ParserFloat64, nil
	case kindString:
		return tempfuncs.DummyStringParser, nil
	case kindBool:
		return tempfuncs.ParserBoolean, nil
	}
	return "", fmt.Errorf("no stringer function found for given field: %s", ident.Name)
}

func isDeepString(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.Ident:
		return isIdentDeepString(x)
	case *ast.SelectorExpr:
		_, ok := x.X.(*ast.Ident)
		if !ok {
			return false
		}
		return isIdentDeepString(x.Sel)
	case *ast.StarExpr:
		return isDeepString(x.X)
	default:
		return false
	}
}

func isIdentDeepString(ident *ast.Ident) bool {
	if ident.Obj != nil {
		ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false
		}
		return isDeepString(ts.Type)
	}
	switch ident.Name {
	case kindString:
		return true
	}
	return false
}
