package ast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/neuronlabs/inflection"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/strcase"
	"golang.org/x/tools/go/packages"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/input"
	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tempfuncs"
)

const (
	kindInt     = "int"
	kindInt8    = "int8"
	kindInt16   = "int16"
	kindInt32   = "int32"
	kindInt64   = "int64"
	kindUint    = "uint"
	kindUint8   = "uint8"
	kindUint16  = "uint16"
	kindUint32  = "uint32"
	kindUint64  = "uint64"
	kindFloat32 = "float32"
	kindFloat64 = "float64"
	kindByte    = "byte"
	kindRune    = "rune"
	kindString  = "string"
	kindBool    = "bool"
	kindUintptr = "uintptr"
	kindPointer = "Pointer"
	kindNil     = "nil"
)

// NewModelGenerator creates new model generator.
func NewModelGenerator(namingConvention string, types, tags, exclude []string) *ModelGenerator {
	gen := &ModelGenerator{
		models:              map[string]*input.Model{},
		modelsFiles:         map[*input.Model]string{},
		Types:               types,
		Exclude:             exclude,
		Tags:                tags,
		importFields:        map[string]map[string][]*ast.Ident{},
		imports:             map[string]string{},
		loadedTypes:         map[string]*ast.TypeSpec{},
		modelImportedFields: map[*input.Model][]*importField{},
		typeMethods:         map[string][]*Method{},
	}
	switch namingConvention {
	case "kebab":
		gen.namerFunc = strcase.ToKebab
	case "lower_camel":
		gen.namerFunc = strcase.ToLowerCamel
	case "camel":
		gen.namerFunc = strcase.ToCamel
	default:
		gen.namerFunc = strcase.ToSnake
	}
	return gen
}

// ModelGenerator is the neuron model generator.
type ModelGenerator struct {
	namerFunc           func(s string) string
	pkgs                []*packages.Package
	Tags                []string
	Types               []string
	Exclude             []string
	loadedTypes         map[string]*ast.TypeSpec
	imports             map[string]string
	importFields        map[string]map[string][]*ast.Ident
	models              map[string]*input.Model
	modelsFiles         map[*input.Model]string
	modelImportedFields map[*input.Model][]*importField
	typeMethods         map[string][]*Method
}

// Method is the structure that defines model methods.
type Method struct {
	Pointer         bool
	Name            string
	ParameterTypes  []string
	ReturnTypes     []string
	ReturnStatement string
}

// IsNeuronCollectionName checks if the method is for neuron collection name.
func (m *Method) IsNeuronCollectionName() bool {
	return m.Name == "NeuronCollectionName" &&
		len(m.ParameterTypes) == 0 &&
		len(m.ReturnTypes) == 1 && m.ReturnTypes[0] == "string"
}

// Collections return generator collections.
func (g *ModelGenerator) Collections(packageName string, isModelImported bool) (collections []*input.CollectionInput) {
	for _, model := range g.Models() {
		collections = append(collections, model.CollectionInput(packageName, isModelImported))
	}
	return collections
}

// CollectionInput creates collection input for provided model.
func (g *ModelGenerator) CollectionInput(packageName string, isModelImported bool, modelName string) (*input.CollectionInput, error) {
	m, ok := g.models[modelName]
	if !ok {
		return nil, fmt.Errorf("model: '%s' not found", modelName)
	}
	return m.CollectionInput(packageName, isModelImported), nil
}

// HasCollectionInitializer checks if the package contains collection initializer.
func (g *ModelGenerator) HasCollectionInitializer() bool {
	var rootPkg string
	for _, model := range g.Models() {
		rootPkg = model.PackageName
	}

	for _, pkg := range g.pkgs {
		if rootPkg != pkg.Name {
			continue
		}
		for _, file := range pkg.CompiledGoFiles {
			log.Debugf("File: %s\n", file)
			if file == "initialize_collections.neuron.go" {
				return true
			}
		}
	}
	return false
}

// CollectionInitializer gets collection initializer.
func (g *ModelGenerator) CollectionInitializer() *input.Collections {
	model := g.Models()[0]
	col := &input.Collections{PackageName: model.PackageName}
	return col
}

// ExtractPackages extracts all models for provided in the packages.
func (g *ModelGenerator) ExtractPackages() error {
	// Find all Types that might be potential models
	for _, pkg := range g.pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, isGenDecl := decl.(*ast.GenDecl)
				if !isGenDecl {
					continue
				}
				if genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					switch st := spec.(type) {
					case *ast.TypeSpec:
						if st.Name == nil {
							continue
						}
						g.loadedTypes[st.Name.Name] = st
					}
				}
			}
		}
	}

	for _, pkg := range g.pkgs {
		isTesting := strings.HasSuffix(pkg.ID, ".test]")

		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					if d.Tok == token.TYPE {
						models, err := g.extractFileModels(d, file, pkg)
						if err != nil {
							return err
						}
						for _, model := range models {
							g.modelsFiles[model] = filepath.Join(pkg.PkgPath, file.Name.Name)
							model.PackageName = pkg.Name
							model.TestFile = isTesting
						}
					}
				case *ast.FuncDecl:
					if d.Recv != nil {
						g.extractMethods(pkg.Name, d)
					}
				}
			}
		}
	}

	for _, modelName := range g.Types {
		if _, ok := g.models[modelName]; !ok {
			return fmt.Errorf("model: '%s' not found", modelName)
		}
	}

	if len(g.models) == 0 {
		return errors.New("no models found")
	}

	if err := g.parseImportPackages(); err != nil {
		return err
	}

	// Find the most common receiver for the models.
	for _, pkg := range g.pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				dt, ok := decl.(*ast.FuncDecl)
				if ok {
					if dt.Recv != nil {
						g.getMethodReceivers(dt)
						g.extractMethods(pkg.Name, dt)
					}
				}
			}
		}
	}

	// Search and set most common receiver for each model.
	for _, model := range g.models {
		g.setReceiver(model)
	}

	// Set primary fields for the models.
	for _, model := range g.models {
		if err := g.checkModelMethods(model); err != nil {
			return err
		}
		if err := g.setFieldsStringGetters(model); err != nil {
			return err
		}
		if err := g.setFieldStringParserFunctions(model); err != nil {
			return err
		}
		if err := g.setMany2ManyRelations(model); err != nil {
			return err
		}
	}

	// Sort model fields.
	for _, model := range g.models {
		model.SortFields()
	}
	return nil
}

func (g *ModelGenerator) setMany2ManyRelations(model *input.Model) error {
	for _, relation := range model.Relations {
		if !relation.IsSlice {
			continue
		}
		if !g.isMany2ManyRelation(relation.Ast) {
			continue
		}
		joinModel, ok := g.many2manyJoinModel(relation.Ast)
		if ok {
			if _, ok = g.loadedTypes[joinModel]; !ok {
				return fmt.Errorf("provided join model: '%s' is not found for the models: '%s' relation: '%s'", joinModel, model.Name, relation.Name)
			}
		} else {
			relatedModel := relation.BaseType()
			var selector string
			if i := strings.IndexRune(relatedModel, '.'); i != -1 {
				selector = relatedModel[:i]
				relatedModel = relatedModel[i+1:]
			}
			name1 := model.Name + inflection.Plural(relatedModel)
			name2 := relatedModel + inflection.Plural(model.Name)
			if joinModelType, ok := g.loadedTypes[name1]; ok {
				joinModel = joinModelType.Name.Name
			} else if joinModelType, ok := g.loadedTypes[name2]; ok {
				joinModel = joinModelType.Name.Name
			} else if joinModelType, ok := g.loadedTypes[selector+"."+name1]; ok {
				joinModel = joinModelType.Name.Name
			} else if joinModelType, ok := g.loadedTypes[selector+"."+name2]; ok {
				joinModel = joinModelType.Name.Name
			} else {
				return fmt.Errorf("no join model found for the models: '%s' relation: '%s'", model.Name, relation.Name)
			}
		}
		relation.JoinModel = joinModel
	}
	return nil
}

// Models returns generator models.
func (g *ModelGenerator) Models() (models []*input.Model) {
	for model := range g.modelsFiles {
		models = append(models, model)
	}
	// Sort the results by model name.
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	return models
}

// ParsePackages analyzes the single package constructed from the patterns and Tags.
// ParsePackages exits if there is an error.
func (g *ModelGenerator) ParsePackages(patterns ...string) {
	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedImports | packages.NeedDeps | packages.NeedFiles | packages.NeedName,
		Tests: true,
	}
	if len(g.Tags) > 0 {
		cfg.BuildFlags = []string{fmt.Sprintf("-Tags=%s", strings.Join(g.Tags, " "))}
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No packages found:\n")
		os.Exit(1)
	}
	g.pkgs = pkgs
}

func (g *ModelGenerator) extractMethods(pkgName string, dt *ast.FuncDecl) {
	var pointer bool
	var typeName string
	for _, r := range dt.Recv.List {
		if len(r.Names) == 0 {
			continue
		}
		switch t := r.Type.(type) {
		case *ast.StarExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				continue
			}
			typeName = ident.Name
			break
		case *ast.Ident:
			typeName = t.Name
			break
		}
	}
	if len(typeName) == 0 {
		return
	}
	if unicode.IsLower([]rune(typeName)[0]) {
		return
	}

	typeName = pkgName + "." + typeName
	method := &Method{
		Pointer: pointer,
		Name:    dt.Name.Name,
	}
	for _, mt := range g.typeMethods[typeName] {
		if mt.Name == method.Name {
			return
		}
	}
	if dt.Type.Params != nil {
		for _, p := range dt.Type.Params.List {
			method.ParameterTypes = append(method.ParameterTypes, g.fieldTypeName(p.Type))
		}
	}
	if dt.Type.Results != nil {
		for _, r := range dt.Type.Results.List {
			method.ReturnTypes = append(method.ReturnTypes, g.fieldTypeName(r.Type))
		}
	}
	if method.IsNeuronCollectionName() {
		for _, stmt := range dt.Body.List {
			if rtrn, ok := stmt.(*ast.ReturnStmt); ok {
				if len(rtrn.Results) == 1 {
					if bs, ok := rtrn.Results[0].(*ast.BasicLit); ok {
						method.ReturnStatement = strings.TrimPrefix(bs.Value, "\"")
						method.ReturnStatement = strings.TrimSuffix(method.ReturnStatement, "\"")
					}
				}
			}
		}
	}
	log.Debug2f("Extracted method: %#v for type: '%s'\n", method, typeName)
	g.typeMethods[typeName] = append(g.typeMethods[typeName], method)
}

func (g *ModelGenerator) setReceiver(model *input.Model) {
	// Get the most common receiver from existing methods for given model.
	var (
		mostCommonReceiver string
		maxCount           int
	)
	for name, count := range model.Receivers {
		if maxCount > count {
			maxCount = count
			mostCommonReceiver = name
		}
	}
	// If no receivers were found yet - set the receiver to the lowered first letter of the model name.
	if mostCommonReceiver == "" {
		mostCommonReceiver = strings.ToLower(model.Name[:1])
	}
	model.Receiver = mostCommonReceiver
}

func (g *ModelGenerator) getMethodReceivers(dt *ast.FuncDecl) {
	for _, r := range dt.Recv.List {
		if len(r.Names) == 0 {
			continue
		}
		switch t := r.Type.(type) {
		case *ast.StarExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				continue
			}
			model, ok := g.models[ident.Name]
			if ok {
				model.Receivers[r.Names[0].Name]++
			}
		case *ast.Ident:
			model, ok := g.models[t.Name]
			if ok {
				model.Receivers[r.Names[0].Name]++
			}
		}
	}
}

func isElemPointer(field *ast.Field) bool {
	return isExprElemPointer(field.Type)
}

func isExprElemPointer(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.StarExpr:
		return isExprElemPointer(x.X)
	case *ast.ArrayType:
		_, isStar := x.Elt.(*ast.StarExpr)
		return isStar
	}
	return false
}

func (g *ModelGenerator) setModelField(astField *ast.Field, inputField *input.Field, imported bool) (err error) {
	isBS, isBSWrapped := isByteSliceWrapper(astField.Type)
	inputField.IsByteSlice = isBS
	inputField.Sortable = isSortable(astField)
	inputField.IsSlice = isMany(astField.Type)
	if inputField.IsSlice {
		if as := g.getArraySize(astField.Type); as != "" {
			inputField.ArraySize, err = strconv.Atoi(as)
			if err != nil {
				log.Debug("Array size is not a number: %v", err)
			}
		}
	}
	inputField.IsString = isDeepString(astField.Type)
	inputField.IsPointer = isPointer(astField)
	inputField.AlternateTypes = g.getAlternateTypes(astField.Type)
	if _, ok := tempfuncs.AlternateTypes[inputField.Type]; !ok {
		tempfuncs.AlternateTypes[inputField.Type] = inputField.AlternateTypes
	}
	g.setFieldZeroValue(inputField, astField.Type, imported)
	inputField.Selector = getSelector(astField.Type)
	if isBSWrapped {
		inputField.WrappedTypes = []string{"[]byte"}
	} else if !isBS {
		inputField.WrappedTypes = g.getFieldWrappedTypes(astField)
	}
	return err
}

type fieldTag struct {
	key    string
	values []string
}

func isByteSlice(arr *ast.ArrayType) bool {
	ident, ok := arr.Elt.(*ast.Ident)
	if !ok {
		return false
	}
	if ident.Name != "byte" {
		return false
	}
	return arr.Len == nil
}

func isSortable(arr *ast.Field) bool {
	if arr.Tag == nil {
		return true
	}
	tags := extractTags(arr.Tag.Value, "neuron", ";", ",")
	for _, tag := range tags {
		switch tag.key {
		case "nosort", "no_sort":
			return false
		}
	}
	return true
}

func isByteSliceWrapper(expr ast.Expr) (isTypeByteSlice bool, isWrapper bool) {
	switch tp := expr.(type) {
	case *ast.ArrayType:
		if tp.Len != nil {
			return false, false
		}
		return isByteSlice(tp), false
	case *ast.Ident:
		if tp.Obj == nil {
			return false, false
		}
		typeSpec, ok := tp.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false, false
		}
		art, ok := typeSpec.Type.(*ast.ArrayType)
		if !ok {
			return false, false
		}
		if art.Len != nil {
			return false, false
		}
		ident, ok := art.Elt.(*ast.Ident)
		if !ok {
			return false, false
		}
		return ident.Name == "byte", true
	default:
		return false, false
	}
}

func isExported(field *ast.Field) bool {
	if len(field.Names) == 0 {
		return false
	}
	return field.Names[0].IsExported()
}

func (g *ModelGenerator) isImported(file *ast.File, field *ast.Field) *importField {
	switch tp := field.Type.(type) {
	case *ast.StarExpr:
		sel, isSelector := tp.X.(*ast.SelectorExpr)
		if !isSelector {
			return nil
		}
		return g.createImportField(file, sel)
	case *ast.SelectorExpr:
		return g.createImportField(file, tp)
	default:
		return nil
	}
}

func (g *ModelGenerator) createImportField(file *ast.File, sel *ast.SelectorExpr) *importField {
	pkgIdent, isIdent := sel.X.(*ast.Ident)
	if !isIdent {
		return nil
	}
	i := &importField{
		Ident: sel.Sel,
	}
	for _, imp := range file.Imports {
		p := strings.Trim(imp.Path.Value, "\"")
		if strings.HasSuffix(p, pkgIdent.Name) {
			i.Path = p
			break
		}
	}
	return i
}

func isPrimary(field *ast.Field) bool {
	// Find a neuron primary tag.
	if field.Tag != nil {
		tags := extractTags(field.Tag.Value, "neuron", ";", ",")

		for _, tag := range tags {
			if tag.key == "-" {
				return false
			}
			if strings.EqualFold(tag.key, "type") {
				for _, value := range tag.values {
					switch value {
					case "pk", "primary", "id":
						return true
					}
				}
				break
			}
		}
	}
	// Check if the name suggests it is the "ID" field.
	if strings.EqualFold(field.Names[0].Name, "ID") {
		return true
	}
	return false
}

func (g *ModelGenerator) isMany2ManyRelation(field *ast.Field) bool {
	if field.Tag == nil {
		return false
	}
	tags := extractTags(field.Tag.Value, "neuron", ";", ",")

	for _, tag := range tags {
		if strings.EqualFold(tag.key, "many2many") {
			return true
		}
	}
	return false
}

func (g *ModelGenerator) many2manyJoinModel(field *ast.Field) (string, bool) {
	tags := extractTags(field.Tag.Value, "neuron", ";", ",")

	for _, tag := range tags {
		if strings.EqualFold(tag.key, "many2many") {
			if len(tag.values) >= 1 {
				if tag.values[0] == "_" {
					return "", false
				}
				return tag.values[0], true
			}
		}
	}
	return "", false
}

func (g *ModelGenerator) isFieldRelation(field *ast.Field) bool {
	if field.Tag != nil {
		tags := extractTags(field.Tag.Value, "neuron", ";", ",")

		for _, tag := range tags {
			if tag.key == "-" {
				return false
			}
			if strings.EqualFold(tag.key, "type") {
				for _, value := range tag.values {
					switch value {
					case "relation", "rel", "relationship":
						return true
					}
				}
				break
			}
		}
	}
	return g.isRelation(field.Type)
}

func isPointer(field *ast.Field) bool {
	_, ok := field.Type.(*ast.StarExpr)
	return ok
}

func isMany(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			return false
		}
		// If this is a wrapper around the slice it would be a type
		dt, ok := t.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false
		}
		return isMany(dt.Type)
	case *ast.ArrayType:
		return true
	case *ast.StarExpr:
		return isMany(t.X)
	case *ast.SelectorExpr:
		if t.Sel.Obj != nil {
			ts, ok := t.Sel.Obj.Decl.(*ast.TypeSpec)
			if ok {
				return isMany(ts.Type)
			}
		}
		return false
	default:
		return false
	}
}

func (g *ModelGenerator) isRelation(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			tp, ok := g.loadedTypes[t.Name]
			if !ok {
				log.Debugf(" (obj == nil [%v] - not a relation) ", t)
				return false
			}
			return g.isRelation(tp.Type)
		}
		ts, ok := t.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			log.Debugf(" (obj.Decl is not a type spec: %T) ", t.Obj.Decl)
			return false
		}
		return g.isRelation(ts.Type)
	case *ast.StarExpr:
		return g.isRelation(t.X)
	case *ast.StructType:
		// Search for the primary key field.
		for _, structField := range t.Fields.List {
			if len(structField.Names) == 0 {
				continue
			}
			name := structField.Names[0]
			if !name.IsExported() {
				continue
			}
			if isPrimary(structField) {
				return true
			}
		}
	case *ast.ArrayType:
		return g.isRelation(t.Elt)
	case *ast.SelectorExpr:
		return g.isRelation(t.Sel)
	}
	return false
}

func (g *ModelGenerator) fieldTypeName(expr ast.Expr) string {
	switch tp := expr.(type) {
	case *ast.Ident:
		return tp.Name
	case *ast.ArrayType:
		return "[" + g.getArrayTypeSize(tp) + "]" + g.fieldTypeName(tp.Elt)
	case *ast.StarExpr:
		return "*" + g.fieldTypeName(tp.X)
	case *ast.StructType:
		name := "struct{"
		for i, field := range tp.Fields.List {
			if len(field.Names) > 0 {
				name += g.fieldTypeName(field.Names[0]) + " "
			}
			name += g.fieldTypeName(field.Type)
			if i != len(tp.Fields.List)-1 {
				name += ";"
			}
		}
		name += "}"
		return name
	case *ast.MapType:
		return "map[" + g.fieldTypeName(tp.Key) + "]" + g.fieldTypeName(tp.Key)
	case *ast.ChanType:
		switch tp.Dir {
		case ast.RECV:
			return "chan<- " + g.fieldTypeName(tp.Value)
		case ast.SEND:
			return "<- chan " + g.fieldTypeName(tp.Value)
		default:
			return "chan " + g.fieldTypeName(tp.Value)
		}
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.SelectorExpr:
		return g.fieldTypeName(tp.X) + "." + tp.Sel.Name
	case *ast.Ellipsis:
		return g.fieldTypeName(tp.Elt)
	default:
		log.Debugf("Unknown field type: %#v\n", tp)
	}
	return ""
}

func extractTags(structTag string, tagName string, tagSeparator, valueSeparator string) []*fieldTag {
	structTag = strings.TrimPrefix(structTag, "`")
	structTag = strings.TrimSuffix(structTag, "`")
	tag, ok := reflect.StructTag(structTag).Lookup(tagName)
	if !ok {
		return nil
	}

	var (
		separators []int
		tags       []*fieldTag
		options    []string
	)

	tagSeparatorRune := []rune(tagSeparator)[0]

	// find all the separators
	for i, r := range tag {
		if i != 0 && r == tagSeparatorRune {
			// check if the  rune before is not an 'escape'
			if tag[i-1] != '\\' {
				separators = append(separators, i)
			}
		}
	}

	// iterate over the option separators
	for i, sep := range separators {
		if i == 0 {
			options = append(options, tag[:sep])
		} else {
			options = append(options, tag[separators[i-1]+1:sep])
		}

		if i == len(separators)-1 {
			options = append(options, tag[sep+1:])
		}
	}
	// if no separators found add the option as whole tag tag
	if options == nil {
		options = append(options, tag)
	}
	// options should be now a legal values defined for the struct tag
	for _, o := range options {
		var equalIndex int
		// find the equalIndex
		for i, r := range o {
			if r == '=' {
				if i != 0 && o[i-1] != '\\' {
					equalIndex = i
					break
				}
			}
		}
		fTag := &fieldTag{}
		if equalIndex != 0 {
			// The left part is the key.
			fTag.key = o[:equalIndex]
			// The right would be the values.
			fTag.values = strings.Split(o[equalIndex+1:], valueSeparator)
		} else {
			// In that case only the key should exists.
			fTag.key = o
		}
		tags = append(tags, fTag)
	}
	return tags
}

func (g *ModelGenerator) getFieldWrappedTypes(field *ast.Field) []string {
	expr := field.Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	var sel string
	var ident *ast.Ident
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		ident = selector.Sel
		sel = g.fieldTypeName(selector.X)
	} else if x, ok := field.Type.(*ast.Ident); ok {
		ident = x
	} else {
		return nil
	}
	var (
		ts *ast.TypeSpec
		ok bool
	)
	if ident.Obj == nil {
		name := ident.Name
		if sel != "" {
			name = sel + "." + name
		}
		ts, ok = g.loadedTypes[name]
		if !ok {
			return nil
		}
	} else {
		ts, ok = ident.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil
		}
	}

	return g.getWrappedTypes(ts.Type, sel)
}

func (g *ModelGenerator) getWrappedTypes(expr ast.Expr, sel string) []string {
	switch x := expr.(type) {
	case *ast.Ident:
		return g.getWrappedIdent(sel, x)
	case *ast.StarExpr:
		types := g.getWrappedTypes(x.X, sel)
		for i := range types {
			types[i] = "*" + types[0]
		}
		return types
	case *ast.SelectorExpr:
		return g.getWrappedSelector(x)
	// TODO: add case *ast.ArrayType
	case *ast.ArrayType:
		tp := "["
		if size := g.getArrayTypeSize(x); size != "" && size != "..." {
			tp += size
		}
		tp += "]"
		tps := g.getWrappedTypes(x.Elt, sel)
		for i := range tps {
			tps[i] = tp + tps[i]
		}
		return tps
	default:
		return []string{}
	}
}

func (g *ModelGenerator) getWrappedSelector(expr *ast.SelectorExpr) []string {
	packageName := g.fieldTypeName(expr.X)
	return g.getWrappedIdent(packageName, expr.Sel)
}

func (g *ModelGenerator) getWrappedIdent(selector string, expr *ast.Ident) []string {
	name := expr.Name
	if selector != "" {
		name = fmt.Sprintf("%s.%s", selector, name)
	}
	var (
		ts *ast.TypeSpec
		ok bool
	)
	if expr.Obj == nil {
		ts, ok = g.loadedTypes[expr.Name]
		if !ok {
			return []string{expr.Name}
		}
	} else {
		ts, ok = expr.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return []string{name}
		}
	}
	return append([]string{name}, g.getWrappedTypes(ts.Type, "")...)
}

func getSelector(expr ast.Expr) string {
	switch tp := expr.(type) {
	case *ast.StarExpr:
		return getSelector(tp.X)
	case *ast.SelectorExpr:
		ident, ok := tp.X.(*ast.Ident)
		if !ok {
			return ""
		}
		return ident.Name
	default:
		return ""
	}
}

func (g *ModelGenerator) setFieldZeroValue(field *input.Field, expr ast.Expr, imported bool) {
	// Check if given type implements ZeroChecker.
	// TODO: add check if type implements query.ZeroChecker.
	field.Zero = g.getZeroValue(expr)
	array, ok := expr.(*ast.ArrayType)
	if ok && array.Len == nil {
		field.BeforeZero = "len("
		field.AfterZero = ") == 0"
	} else {
		field.AfterZero = " == " + field.Zero
	}
}

func (g *ModelGenerator) getZeroValue(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		var (
			ts *ast.TypeSpec
			ok bool
		)
		if x.Obj == nil {
			switch x.Name {
			case kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
				kindByte, kindRune, kindFloat32, kindFloat64:
				return "0"
			case kindString:
				return "\"\""
			case kindBool:
				return "false"
			case kindUintptr:
				return "0"
			case kindPointer:
				return kindNil
			default:
				ts, ok = g.loadedTypes[x.Name]
			}
		} else {
			ts, ok = x.Obj.Decl.(*ast.TypeSpec)
		}
		if !ok {
			return g.fieldTypeName(expr) + "{}"
		}
		switch ts.Type.(type) {
		case *ast.StructType:
		default:
			return fmt.Sprintf("%s(%s)", g.fieldTypeName(expr), g.getZeroValue(ts.Type))
		}
		return g.fieldTypeName(expr) + "{}"
	case *ast.StarExpr:
		return kindNil
	case *ast.ArrayType:
		if x.Len == nil {
			// A slice can be nil
			return kindNil
		}
		// The array must be defined to zero values.
		return g.fieldTypeName(expr) + "{}"
	case *ast.MapType:
		return kindNil
	case *ast.StructType:
		return g.fieldTypeName(expr) + "{}"
	case *ast.ChanType:
		return kindNil
	case *ast.SelectorExpr:
		selector, ok := x.X.(*ast.Ident)
		if !ok {
			return g.fieldTypeName(expr) + "{}"
		}
		return selector.Name + "." + g.getZeroValue(x.Sel)
	default:
		return kindNil
	}
}

func (g *ModelGenerator) getAlternateTypes(expr ast.Expr) []string {
	switch exprType := expr.(type) {
	case *ast.Ident:
		return g.getIdentAlternateTypes(exprType)
	case *ast.ArrayType:
		return g.getArrayAlternateTypes(exprType)
	case *ast.SelectorExpr:
		return g.getIdentAlternateTypes(exprType.Sel)
	case *ast.StarExpr:
		return g.getAlternateTypes(exprType.X)
	}
	return []string{}
}

func (g *ModelGenerator) getArrayAlternateTypes(expr *ast.ArrayType) []string {
	switch expr.Len.(type) {
	case *ast.Ident:
		return nil
	case *ast.BasicLit: // Array
		return nil
	case *ast.Ellipsis: // [...] Array
		return nil
	default: // Slice
		// check if this is a byte slice
		ident, ok := expr.Elt.(*ast.Ident)
		if !ok {
			return nil
		}
		if ident.Obj != nil {
			return nil
		}
		if ident.Name == kindByte {
			return []string{"string"}
		}
		return nil
	}
}

func (g *ModelGenerator) getIdentAlternateTypes(expr *ast.Ident) []string {
	var (
		ts *ast.TypeSpec
		ok bool
	)
	if expr.Obj == nil {
		ts, ok = g.loadedTypes[expr.Name]
		if !ok {
			return g.getBasicAlternateTypes(expr)
		}
	} else {
		ts, ok = expr.Obj.Decl.(*ast.TypeSpec)
	}
	if !ok {
		return []string{}
	}
	return g.getAlternateTypes(ts.Type)
}

func (g *ModelGenerator) getBasicAlternateTypes(expr *ast.Ident) (alternateTypes []string) {
	switch expr.Name {
	case kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
		kindByte, kindRune, kindFloat32, kindFloat64:
		for _, kind := range []string{kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
			kindFloat32, kindFloat64} {
			if kind != expr.Name {
				alternateTypes = append(alternateTypes, kind)
			}
		}
	case kindString:
		alternateTypes = []string{"[]byte"}
	}
	return alternateTypes
}

func isTypeBasic(expr ast.Expr) bool {
	switch d := expr.(type) {
	case *ast.Ident:
		switch d.Name {
		case kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
			kindByte, kindRune, kindFloat32, kindFloat64, kindString, kindBool:
			return true
		}
	case *ast.StarExpr:
		return isTypeBasic(d.X)
	case *ast.ArrayType:
		return isTypeBasic(d.Elt)
	}
	return false
}
