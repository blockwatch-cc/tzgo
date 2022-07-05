// Copyright (c) 2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
    "encoding"
    "fmt"
    "reflect"
    "strconv"
    "strings"
    "sync"
)

const tagName = "prim"

// typeInfo holds details for the representation of a type.
type typeInfo struct {
    fields []fieldInfo
}

// fieldInfo holds details for the representation of a single field.
type fieldInfo struct {
    idx    []int  // Go struct index
    name   string // field name (= Go struct name)
    typ    OpCode
    path   []int
    nofail bool
}

func (f fieldInfo) String() string {
    return fmt.Sprintf("FieldInfo: name=%s typ=%s goloc=%v primloc=%v", f.name, f.typ, f.idx, f.path)
}

var tinfoMap = make(map[reflect.Type]*typeInfo)
var tinfoLock sync.RWMutex

var (
    binaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
    binaryMarshalerType   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
    textUnmarshalerType   = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
    textMarshalerType     = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
    primMarshalerType     = reflect.TypeOf((*PrimMarshaler)(nil)).Elem()
    primUnmarshalerType   = reflect.TypeOf((*PrimUnmarshaler)(nil)).Elem()
    stringerType          = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
    byteSliceType         = reflect.TypeOf([]byte(nil))
    szPrim                = int(reflect.TypeOf(Prim{}).Size())
)

// getTypeInfo returns the typeInfo structure with details necessary
// for marshaling and unmarshaling of typ.
func getTypeInfo(typ reflect.Type) (*typeInfo, error) {
    tinfoLock.RLock()
    tinfo, ok := tinfoMap[typ]
    tinfoLock.RUnlock()
    if ok {
        return tinfo, nil
    }
    tinfo = &typeInfo{}
    if typ.Kind() != reflect.Struct {
        return nil, fmt.Errorf("micheline: type %s (%s) is not a struct", typ.String(), typ.Kind())
    }
    n := typ.NumField()
    for i := 0; i < n; i++ {
        f := typ.Field(i)
        if (f.PkgPath != "" && !f.Anonymous) || f.Tag.Get(tagName) == "-" {
            continue // Private field
        }

        // For embedded structs, embed their fields.
        if f.Anonymous {
            t := f.Type
            if t.Kind() == reflect.Ptr {
                t = t.Elem()
            }
            if t.Kind() == reflect.Struct {
                inner, err := getTypeInfo(t)
                if err != nil {
                    return nil, err
                }
                for _, finfo := range inner.fields {
                    finfo.idx = append([]int{i}, finfo.idx...)
                    if err := addFieldInfo(typ, tinfo, &finfo); err != nil {
                        return nil, err
                    }
                }
                continue
            }
        }

        finfo, err := structFieldInfo(typ, &f)
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
func structFieldInfo(typ reflect.Type, f *reflect.StructField) (*fieldInfo, error) {
    finfo := &fieldInfo{
        idx:  f.Index,
        name: f.Name,
    }
    tag := f.Tag.Get(tagName)
    kind := f.Type.Kind()
    typname := f.Type.String()

    // Parse struct tag
    tokens := strings.Split(tag, ",")
    if len(tokens) > 1 {
        finfo.name = tokens[0]
        for _, flag := range tokens[1:] {
            ff := strings.Split(flag, "=")
            switch ff[0] {
            case "path":
                for _, v := range strings.Split(strings.TrimSuffix(strings.TrimPrefix(ff[1], "/"), "/"), "/") {
                    i, err := strconv.Atoi(v)
                    if err != nil {
                        return nil, fmt.Errorf("micheline: invalid path %q in field %s: %v", ff[1], f.Name, err)
                    }
                    finfo.path = append(finfo.path, i)
                }
            case "nofail":
                finfo.nofail = true
            }
        }
    }
    switch kind {
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        finfo.typ = T_NAT
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        finfo.typ = T_INT
    case reflect.Slice:
        if f.Type.Implements(binaryUnmarshalerType) {
            finfo.typ = T_BYTES
        } else if f.Type.Implements(textUnmarshalerType) {
            finfo.typ = T_STRING
        } else if f.Type == byteSliceType {
            finfo.typ = T_BYTES
        } else {
            finfo.typ = T_LIST
        }
    case reflect.Map:
        finfo.typ = T_MAP
    case reflect.String:
        finfo.typ = T_STRING
    case reflect.Bool:
        finfo.typ = T_BOOL
    case reflect.Struct:
        switch typname {
        case "time.Time":
            finfo.typ = T_TIMESTAMP
        case "tezos.Address":
            finfo.typ = T_ADDRESS
        case "tezos.Z":
            finfo.typ = T_NAT
        case "tezos.N":
            finfo.typ = T_NAT
        case "tezos.Key":
            finfo.typ = T_KEY
        case "tezos.Signature":
            finfo.typ = T_SIGNATURE
        case "tezos.ChainIdHash":
            finfo.typ = T_CHAIN_ID
        default:
            if f.Type.Implements(binaryUnmarshalerType) {
                finfo.typ = T_BYTES
            } else {
                return nil, fmt.Errorf("micheline: unsupported embedded struct type %s", f.Type)
            }
        }
    default:
        return nil, fmt.Errorf("micheline: unsupported type %s (%v) for field %s",
            f.Type, f.Type.Kind(), finfo.name)
    }

    return finfo, nil
}

func addFieldInfo(typ reflect.Type, tinfo *typeInfo, newf *fieldInfo) error {
    var conflicts []int
    // Find all conflicts.
    for i := range tinfo.fields {
        oldf := &tinfo.fields[i]
        if newf.name == oldf.name {
            conflicts = append(conflicts, i)
        }
    }

    // Return the first error.
    for _, i := range conflicts {
        oldf := &tinfo.fields[i]
        f1 := typ.FieldByIndex(oldf.idx)
        f2 := typ.FieldByIndex(newf.idx)
        return fmt.Errorf("micheline: %s field %q with tag %q conflicts with field %q with tag %q", typ, f1.Name, f1.Tag.Get(tagName), f2.Name, f2.Tag.Get(tagName))
    }

    // Without conflicts, add the new field and return.
    tinfo.fields = append(tinfo.fields, *newf)
    return nil
}

// value returns v's field value corresponding to finfo.
// It's equivalent to v.FieldByIndex(finfo.idx), but initializes
// and dereferences pointers as necessary.
func (finfo *fieldInfo) value(v reflect.Value) reflect.Value {
    for i, x := range finfo.idx {
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
func derefIndirect(v interface{}) reflect.Value {
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

func indirectType(typ reflect.Type) reflect.Type {
    if typ.Kind() == reflect.Ptr {
        val := reflect.New(typ.Elem())
        return val.Elem().Type()
    }
    return typ
}
