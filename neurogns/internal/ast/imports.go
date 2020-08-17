package ast

import (
	"errors"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"

	"github.com/neuronlabs/neuron-extensions/neurogns/input"
)

type importField struct {
	AstField *ast.Field
	Field    *input.Field
	Path     string
	Ident    *ast.Ident
}

func (g *ModelGenerator) parseImportPackages() error {
	if len(g.imports) == 0 {
		fmt.Println("nothing to import...")
		return nil
	}
	var pkgPaths []string
	for pkg := range g.importFields {
		pkgPaths = append(pkgPaths, pkg)
	}
	cfg := &packages.Config{
		Mode:       packages.NeedSyntax | packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		BuildFlags: g.Tags,
	}
	pkgs, err := packages.Load(cfg, pkgPaths...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 1 {
		return errors.New("error while loading import packages")
	}
	for _, pkg := range pkgs {
		// fmt.Printf("Imported package: %s\n", pkg.Name)
		importTypes := g.importFields[pkg.ID]
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						tp, isType := spec.(*ast.TypeSpec)
						if !isType {
							continue
						}
						idents, ok := importTypes[tp.Name.Name]
						if !ok {
							continue
						}
						obj := ast.NewObj(ast.Typ, tp.Name.Name)
						obj.Decl = tp
						for _, ident := range idents {
							ident.Obj = obj
						}
					}
				case *ast.FuncDecl:
					// Check if the function is a method.
					if d.Recv == nil {
						continue
					}
					g.extractMethods(pkg.Name, d)
				default:
					continue
				}
			}
		}
	}

	for model, importedFields := range g.modelImportedFields {
		modelImports := map[string]struct{}{}
		for _, importedField := range importedFields {
			modelImports[importedField.Path] = struct{}{}
			if g.isFieldRelation(importedField.AstField) {
				fmt.Printf("Model: '%s' Field: '%s' is relation\n", model.Name, importedField.Field.Name)
				importedField.Field.IsSlice = isMany(importedField.AstField.Type)
				importedField.Field.IsElemPointer = isElemPointer(importedField.AstField)
				importedField.Field.IsPointer = isPointer(importedField.AstField)
				model.Relations = append(model.Relations, importedField.Field)
				continue
			}
			if err = g.setModelField(importedField.AstField, importedField.Field, true); err != nil {
				return err
			}
			model.Fields = append(model.Fields, importedField.Field)
		}

		// Add all imports for given model.
		for imp := range modelImports {
			model.AddImport(imp)
		}
	}
	return nil
}
