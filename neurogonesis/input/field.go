package input

import (
	"go/ast"
	"strings"
)

// Field is a structure used to insert into model field template.
type Field struct {
	Index            int
	Name, NeuronName string
	// Type is current field Type for given model.
	Type                        string
	BeforeZero, AfterZero, Zero string
	// Getting string value flags.
	IsTextUnmarshaler, IsTextMarshaler, IsString bool

	AlternateTypes, WrappedTypes                   []string
	Scanner, Sortable, ZeroChecker                 bool
	Tags                                           string
	IsPointer, IsElemPointer, IsSlice, IsByteSlice bool
	IsImported                                     bool
	ArraySize                                      int
	// Selector is the import package name for given field
	// i.e.:
	// 	for field - time.Time - the Type would be 'Time' and Selector 'time'.
	Selector     string
	StringGetter string
	StringSetter string

	Model     *Model
	JoinModel string
	Ast       *ast.Field
}

// BaseType returns field type without pointer or slices.
func (f *Field) BaseType() string {
	index := strings.LastIndexAny(f.Type, "[]*")
	if index == -1 {
		return f.Type
	}
	return f.Type[index+1:]
}

// BaseUnwrappedType.
func (f *Field) BaseUnwrappedType() string {
	var t string
	if len(f.WrappedTypes) == 1 {
		t = f.WrappedTypes[0]
	} else {
		t = f.Type
	}
	index := strings.LastIndexAny(t, "[]*")
	if index == -1 {
		return t
	}
	return t[index+1:]
}

func (f *Field) RelationBaseType() string {
	return f.BaseUnwrappedType()
}

func (f *Field) IsBaseRelationPointer() bool {
	if len(f.WrappedTypes) != 1 {
		return f.IsElemPointer
	}
	t := f.WrappedTypes[0]

	index := strings.IndexRune(t, '*')
	return index != -1
}

// IsZero returns string template IsZero checker.
func (f *Field) IsZero() string {
	if f.ZeroChecker {
		return f.Model.Receiver + "." + f.Name + ".IsZero()"
	}
	return f.BeforeZero + f.Model.Receiver + "." + f.Name + f.AfterZero
}

// GetZero gets the zero value string for given field.
func (f *Field) GetZero() string {
	if f.ZeroChecker {
		return f.Model.Receiver + f.Name + ".GetZero()"
	}
	return f.Zero
}
