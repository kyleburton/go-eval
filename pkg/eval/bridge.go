// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eval

import (
	"go/token"
	"log"
	"reflect"
	"math/big"
)

/*
 * Type bridging
 */

var (
	evalTypes   = make(map[reflect.Type]Type)
	nativeTypes = make(map[Type]reflect.Type)
)

// TypeFromNative converts a regular Go type into a the corresponding
// interpreter Type.
func TypeFromNative(t reflect.Type) Type {
	if et, ok := evalTypes[t]; ok {
		return et
	}

	var nt *NamedType
	if t.Name() != "" {
		name := t.PkgPath() + "·" + t.Name()
		nt = &NamedType{token.NoPos, name, nil, true, make(map[string]Method)}
		evalTypes[t] = nt
	}

	var et Type
	switch t.Kind() {
	case reflect.Bool:
		et = BoolType

	case reflect.Float32:
		et = Float32Type
	case reflect.Float64:
		et = Float64Type

	case reflect.Int16:
		et = Int16Type
	case reflect.Int32:
		et = Int32Type
	case reflect.Int64:
		et = Int64Type
	case reflect.Int8:
		et = Int8Type
	case reflect.Int:
		et = IntType

	case reflect.Uint16:
		et = Uint16Type
	case reflect.Uint32:
		et = Uint32Type
	case reflect.Uint64:
		et = Uint64Type
	case reflect.Uint8:
		et = Uint8Type
	case reflect.Uint:
		et = UintType
	case reflect.Uintptr:
		et = UintptrType

	case reflect.String:
		et = StringType
	case reflect.Array:
		et = NewArrayType(int64(t.Len()), TypeFromNative(t.Elem()))
	case reflect.Chan:
		log.Panicf("%T not implemented", t)
	case reflect.Func:
		nin := t.NumIn()
		// Variadic functions have DotDotDotType at the end
		variadic := t.IsVariadic()
		if variadic {
			nin--
		}
		in := make([]Type, nin)
		for i := range in {
			in[i] = TypeFromNative(t.In(i))
		}
		out := make([]Type, t.NumOut())
		for i := range out {
			out[i] = TypeFromNative(t.Out(i))
		}
		et = NewFuncType(in, variadic, out)
	case reflect.Interface:
		log.Panicf("%T not implemented", t)
	case reflect.Map:
		log.Panicf("%T not implemented", t)
	case reflect.Ptr:
		et = NewPtrType(TypeFromNative(t.Elem()))
	case reflect.Slice:
		et = NewSliceType(TypeFromNative(t.Elem()))
	case reflect.Struct:
		n := t.NumField()
		fields := make([]StructField, n)
		for i := 0; i < n; i++ {
			sf := t.Field(i)
			// TODO(austin) What to do about private fields?
			fields[i].Name = sf.Name
			fields[i].Type = TypeFromNative(sf.Type)
			fields[i].Anonymous = sf.Anonymous
		}
		et = NewStructType(fields)
	case reflect.UnsafePointer:
		log.Panicf("%T not implemented", t)
	default:
		log.Panicf("unexpected reflect.Type: %T", t)
	}

	if nt != nil {
		if _, ok := et.(*NamedType); !ok {
			nt.Complete(et)
			et = nt
		}
	}

	nativeTypes[et] = t
	evalTypes[t] = et

	return et
}

// TypeOfNative returns the interpreter Type of a regular Go value.
func TypeOfNative(v interface{}) Type { return TypeFromNative(reflect.TypeOf(v)) }

/*
 * Function bridging
 */

type NativeFunc struct {
	fn      func(*Thread, []Value, []Value)
	in, out int
}

func (f *NativeFunc) NewFrame() *Frame {
	vars := make([]Value, f.in+f.out)
	return &Frame{nil, vars}
}

func (f *NativeFunc) Call(t *Thread) { f.fn(t, t.f.Vars[0:f.in], t.f.Vars[f.in:f.in+f.out]) }

// FuncFromNative creates an interpreter function from a native
// function that takes its in and out arguments as slices of
// interpreter Value's.  While somewhat inconvenient, this avoids
// value marshalling.
func FuncFromNative(fn func(*Thread, []Value, []Value), t *FuncType) FuncValue {
	return &funcV{&NativeFunc{fn, len(t.In), len(t.Out)}}
}

// FuncFromNativeTyped is like FuncFromNative, but constructs the
// function type from a function pointer using reflection.  Typically,
// the type will be given as a nil pointer to a function with the
// desired signature.
func FuncFromNativeTyped(fn func(*Thread, []Value, []Value), t interface{}) (*FuncType, FuncValue) {
	ft := TypeOfNative(t).(*FuncType)
	return ft, FuncFromNative(fn, ft)
}

type vstruct []interface{}

type varray []interface{}

type vslice struct {
	arr      varray
	len, cap int
}

type vmap map[interface{}]interface{}


func ToValue(val interface{}) Value {
	switch val := val.(type) {
	case bool:
		r := boolV(val)
		return &r
	case uint8:
		r := uint8V(val)
		return &r
	case uint:
		r := uintV(val)
		return &r
	case int:
		r := intV(val)
		return &r
	case *big.Int:
		return &idealIntV{val}
	case float64:
		r := float64V(val)
		return &r
	case *big.Rat:
		return &idealFloatV{val}
	case string:
		r := stringV(val)
		return &r
	case vstruct:
		elems := make([]Value, len(val))
		for i, e := range val {
			elems[i] = ToValue(e)
		}
		r := structV(elems)
		return &r
	case varray:
		elems := make([]Value, len(val))
		for i, e := range val {
			elems[i] = ToValue(e)
		}
		r := arrayV(elems)
		return &r
	case vslice:
		return &sliceV{Slice{ToValue(val.arr).(ArrayValue), int64(val.len), int64(val.cap)}}
	case vmap:
		target := evalMap{}
		for k, v := range val {
			target[k] = ToValue(v)
		}
		r := mapV{target}
		return &r
	case Func:
		return &funcV{val}
	}
	log.Panicf("ToValue(%T) not implemented", val)
	panic("unreachable")
}


