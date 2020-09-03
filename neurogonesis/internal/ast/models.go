package ast

import (
	"errors"
	"go/ast"
	"strings"

	"github.com/neuronlabs/inflection"
	"github.com/neuronlabs/neuron/log"
	"golang.org/x/tools/go/packages"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/input"
)

func (g *ModelGenerator) extractFileModels(d *ast.GenDecl, file *ast.File, pkg *packages.Package) (models []*input.Model, err error) {
	for _, spec := range d.Specs {
		switch st := spec.(type) {
		case *ast.TypeSpec:
			structType, ok := st.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if st.Name == nil {
				continue
			}
			modelName := st.Name.Name
			if len(g.Types) != 0 {
				var matchedModel string
				for i, tp := range g.Types {
					if tp == st.Name.Name {
						matchedModel = tp
						g.Types = append(g.Types[:i], g.Types[i+1:]...)
						break
					}
				}
				if matchedModel == "" {
					continue
				}
			} else if len(g.Exclude) != 0 {
				var matchedModel string
				for _, tp := range g.Exclude {
					if tp == st.Name.Name {
						matchedModel = tp
						break
					}
				}
				if matchedModel != "" {
					continue
				}
			}

			model, err := g.extractModel(file, structType, pkg, modelName)
			if err == ErrModelAlreadyFound {
				continue
			}
			if err != nil {
				return nil, err
			}
			if model != nil {
				models = append(models, model)
			}
		default:
			continue
		}
	}
	return models, nil
}

var ErrModelAlreadyFound = errors.New("model already found")

func (g *ModelGenerator) extractModel(file *ast.File, structType *ast.StructType, pkg *packages.Package, modelName string) (model *input.Model, err error) {
	model = &input.Model{
		CollectionName: g.namerFunc(inflection.Plural(modelName)),
		Name:           modelName,
		Receivers:      make(map[string]int),
		PackageName:    pkg.Name,
		PackagePath:    pkg.PkgPath,
	}

	// Find primary field key.
	for i, structField := range structType.Fields.List {
		if len(structField.Names) == 0 {
			// Embedded fields are not taken into account.
			continue
		}
		name := structField.Names[0]
		if !name.IsExported() {
			// Private fields are not taken into account.
			continue
		}

		if !isExported(structField) {
			continue
		}
		field := input.Field{
			Index:      i,
			Name:       name.String(),
			NeuronName: g.namerFunc(name.String()),
			Type:       g.fieldTypeName(structField.Type),
			Model:      model,
			Ast:        structField,
		}
		model.StructFields = append(model.StructFields, &field)

		// Set the Tags for given field.
		if structField.Tag != nil {
			field.Tags = structField.Tag.Value
			tags := extractTags(field.Tags, "neuron", ";", ",")
			for _, tag := range tags {
				if tag.key == "-" {
					continue
				}
			}
		}

		log.Debugf("Model: '%s' Field: '%s' ", modelName, field.Name)
		if g.isFieldRelation(structField) {
			log.Debug("is relation")
			field.IsSlice = isMany(structField.Type)
			field.IsElemPointer = isElemPointer(structField)
			field.IsPointer = isPointer(structField)
			field.WrappedTypes = g.getFieldWrappedTypes(field.Ast)
			field.Selector = getSelector(structField.Type)
			model.Relations = append(model.Relations, &field)
			continue
		} else if importedField := g.isImported(file, structField); importedField != nil {
			log.Debug("is imported")
			importedField.Field = &field
			importedField.AstField = structField
			field.IsImported = true
			if isPrimary(structField) {
				model.Primary = importedField.Field
			}
			g.modelImportedFields[model] = append(g.modelImportedFields[model], importedField)
			continue
		}
		fieldPtr := &field
		if err := g.setModelField(structField, fieldPtr, false); err != nil {
			return nil, err
		}
		// Check if field is a primary key field.
		if isPrimary(structField) {
			model.Primary = fieldPtr
		}
		model.Fields = append(model.Fields, fieldPtr)
	}

	if model.Primary == nil {
		return nil, nil
	}
	defaultModelPackages := []string{
		"github.com/neuronlabs/neuron/errors",
	}
	if pkg.PkgPath != "github.com/neuronlabs/neuron/mapping" {
		model.Imports.Add("github.com/neuronlabs/neuron/mapping")
	}
	for _, pkg := range defaultModelPackages {
		model.Imports.Add(pkg)
	}

	log.Debugf("Adding model: '%s'\n", model.Name)
	if _, ok := g.models[model.Name]; ok {
		return nil, ErrModelAlreadyFound
	}
	g.models[model.Name] = model
	for _, relation := range model.Relations {
		if relation.IsSlice {
			model.MultiRelationer = true
		} else {
			model.SingleRelationer = true
		}
	}
	if len(model.Fields) > 0 {
		model.Fielder = true
	}

	for _, importedField := range g.modelImportedFields[model] {
		g.imports[importedField.Path] = importedField.Ident.Name
		pkgTypes := g.importFields[importedField.Path]
		if pkgTypes == nil {
			pkgTypes = map[string][]*ast.Ident{}
		}
		pkgTypes[importedField.Ident.Name] = append(pkgTypes[importedField.Ident.Name], importedField.Ident)
		g.importFields[importedField.Path] = pkgTypes
	}
	return model, nil
}

func (g *ModelGenerator) ResolveRelationSelectors() {
	for _, model := range g.models {
		for _, relation := range model.Relations {
			// if relation.Selector
			if relation.Selector == "" && !strings.ContainsRune(relation.Type, '.') {
				relation.Selector = model.PackageName

				index := strings.LastIndexAny(relation.Type, "[]*")
				if index == -1 {
					relation.Type = model.PackageName + "." + relation.Type
				} else {
					relation.Type = relation.Type[:index] + relation.Selector + "." + relation.Type[index+1:]
				}
			}
		}
	}
}
func (g *ModelGenerator) getArraySize(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.BasicLit:
		return x.Value
	case *ast.ArrayType:
		return g.getArrayTypeSize(x)
	case *ast.Ident:
		if x.Obj == nil {
			return ""
		}
		ts, ok := x.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return ""
		}
		return g.getArraySize(ts.Type)
	case *ast.StarExpr:
		return g.getArraySize(x.X)
	case *ast.SelectorExpr:
		return g.getArraySize(x.Sel)
	}
	return ""
}

func (g *ModelGenerator) getArrayTypeSize(sl *ast.ArrayType) string {
	switch tp := sl.Len.(type) {
	case *ast.BasicLit:
		return tp.Value
	case *ast.Ellipsis:
		return "..."
	case *ast.Ident:
		if tp.Obj == nil {
			return ""
		}
		tpl, ok := tp.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			tpl, ok = g.loadedTypes[tp.Name]
			if !ok {
				vs, ok := tp.Obj.Decl.(*ast.ValueSpec)
				if !ok {
					return ""
				}
				if len(vs.Values) == 1 {
					return g.getArraySize(vs.Values[0])
				}
				return ""
			}
		}
		return g.getArraySize(tpl.Type)
	default:
		return ""
	}
}
