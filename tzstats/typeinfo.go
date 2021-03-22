// Copyright (c) 2018-2019 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	tagName = "json"
)

// TypeInfo holds details for the representation of a type.
type TypeInfo struct {
	Name     string
	Fields   []FieldInfo
	IsGoType bool
	TagName  string
}

// FieldInfo holds details for the representation of a single field.
type FieldInfo struct {
	Idx      []int
	Name     string
	Alias    string
	Flags    []string
	TypeName string
}

func (t TypeInfo) Aliases() []string {
	s := make([]string, len(t.Fields))
	for i, v := range t.Fields {
		s[i] = v.Alias
	}
	return s
}

func (t TypeInfo) FieldNames() []string {
	s := make([]string, len(t.Fields))
	for i, v := range t.Fields {
		s[i] = v.Name
	}
	return s
}

func (f FieldInfo) String() string {
	return fmt.Sprintf("FieldInfo: %s typ=%s idx=%v", f.Name, f.TypeName, f.Idx)
}

var tinfoMap = make(map[reflect.Type]*TypeInfo)
var tinfoLock sync.RWMutex

var (
	textUnmarshalerType   = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	textMarshalerType     = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	binaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	binaryMarshalerType   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	byteSliceType         = reflect.TypeOf([]byte(nil))
)

// GetTypeInfo returns the typeInfo structure with details necessary
// for marshaling and unmarshaling typ.
func GetTypeInfo(v interface{}, tagname string) (*TypeInfo, error) {
	// Load value from interface
	val := reflect.Indirect(reflect.ValueOf(v))
	if !val.IsValid() {
		return nil, fmt.Errorf("invalid value of type %T", v)
	}
	if tagname == "" {
		tagname = tagName
	}
	return getReflectTypeInfo(val.Type(), tagname)
}

func getReflectTypeInfo(typ reflect.Type, tagname string) (*TypeInfo, error) {
	tinfoLock.RLock()
	tinfo, ok := tinfoMap[typ]
	tinfoLock.RUnlock()
	if ok {
		return tinfo, nil
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %s (%s) is not a struct", typ.String(), typ.Kind())
	}
	tinfo = &TypeInfo{
		Name:     typ.String(),
		IsGoType: true,
		TagName:  tagname,
	}
	n := typ.NumField()
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if (f.PkgPath != "" && !f.Anonymous) || f.Tag.Get(tinfo.TagName) == "-" {
			continue // Private field
		}

		// For embedded structs, embed its fields.
		if f.Anonymous {
			t := f.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() == reflect.Struct {
				inner, err := getReflectTypeInfo(t, tinfo.TagName)
				if err != nil {
					return nil, err
				}
				for _, finfo := range inner.Fields {
					finfo.Idx = append([]int{i}, finfo.Idx...)
					if err := addFieldInfo(typ, tinfo, &finfo); err != nil {
						return nil, err
					}
				}
				continue
			}
		}

		finfo, err := structFieldInfo(typ, &f, tinfo.TagName)
		if err != nil {
			return nil, err
		}

		// Add the field if it doesn't conflict with other fields.
		if err := addFieldInfo(typ, tinfo, finfo); err != nil {
			return nil, err
		}
	}
	tinfoLock.Lock()
	tinfoMap[typ] = tinfo
	tinfoLock.Unlock()
	return tinfo, nil
}

// structFieldInfo builds and returns a fieldInfo for f.
func structFieldInfo(typ reflect.Type, f *reflect.StructField, tagname string) (*FieldInfo, error) {
	finfo := &FieldInfo{Idx: f.Index, Name: f.Name, TypeName: f.Type.String()}
	switch tags := strings.Split(f.Tag.Get(tagname), ","); len(tags) {
	case 0:
		finfo.Alias = finfo.Name
	case 1:
		finfo.Alias = tags[0]
	default:
		finfo.Alias = tags[0]
		finfo.Flags = tags[1:]
	}
	return finfo, nil
}

func addFieldInfo(typ reflect.Type, tinfo *TypeInfo, newf *FieldInfo) error {
	var conflicts []int
	// Find all conflicts.
	for i := range tinfo.Fields {
		oldf := &tinfo.Fields[i]
		if newf.Name == oldf.Name {
			conflicts = append(conflicts, i)
		}
	}

	// Return the first error.
	for _, i := range conflicts {
		oldf := &tinfo.Fields[i]
		f1 := typ.FieldByIndex(oldf.Idx)
		f2 := typ.FieldByIndex(newf.Idx)
		return fmt.Errorf("%s: %s field %q with tag %q conflicts with field %q with tag %q",
			tinfo.TagName, typ, f1.Name, f1.Tag.Get(tinfo.TagName), f2.Name, f2.Tag.Get(tinfo.TagName))
	}

	// Without conflicts, add the new field and return.
	tinfo.Fields = append(tinfo.Fields, *newf)
	return nil
}

// value returns v's field value corresponding to finfo.
// It's equivalent to v.FieldByIndex(finfo.idx), but initializes
// and dereferences pointers as necessary.
func (finfo *FieldInfo) Value(v reflect.Value) reflect.Value {
	for i, x := range finfo.Idx {
		if i > 0 {
			t := v.Type()
			if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}

	return v
}

// Load value from interface, but only if the result will be
// usefully addressable.
func (finfo *FieldInfo) DerefIndirect(v interface{}) reflect.Value {
	return derefValue(reflect.ValueOf(v))
}

func derefValue(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		e := val.Elem()
		if e.Kind() == reflect.Ptr && !e.IsNil() {
			val = e
		}
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}
	return val
}
