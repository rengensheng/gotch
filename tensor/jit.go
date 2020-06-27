package tensor

// JIT interface to run model trained/saved using PyTorch Python API.

// #include "stdlib.h"
import "C"

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"

	// "github.com/sugarme/gotch"
	lib "github.com/sugarme/gotch/libtch"
)

type CIValue struct {
	civalue lib.Civalue
}

type IValueKind struct {
	reflect.Type
}

var (
	NoneVal        IValueKind = IValueKind{reflect.TypeOf(nil)}
	TensorVal      IValueKind = IValueKind{reflect.TypeOf(Tensor{})}
	DoubleVal      IValueKind = IValueKind{reflect.TypeOf(float64(1))}
	IntVal         IValueKind = IValueKind{reflect.TypeOf(int64(1))}
	BoolVal        IValueKind = IValueKind{reflect.TypeOf(true)}
	TupleVal       IValueKind = IValueKind{reflect.TypeOf([]IValue{})}
	IntListVal     IValueKind = IValueKind{reflect.TypeOf([]int64{})}
	DoubleListVal  IValueKind = IValueKind{reflect.TypeOf([]float64{})}
	BoolListVal    IValueKind = IValueKind{reflect.TypeOf([]bool{})}
	StringVal      IValueKind = IValueKind{reflect.TypeOf("")}
	TensorListVal  IValueKind = IValueKind{reflect.TypeOf([]Tensor{})}
	GenericListVal IValueKind = IValueKind{reflect.TypeOf([]IValue{})}
	GenericDictVal IValueKind = IValueKind{reflect.TypeOf(map[IValue]IValue{})} // 2 elements. ? map[IValue]IValue
	GenericVal     IValueKind = IValueKind{reflect.TypeOf(IValue{})}
)

type IValue struct {
	value interface{}
	kind  IValueKind
	name  string
}

// NewIValue creates a new IValue from given value of various types.
func NewIValue(v interface{}) (retVal IValue) {

	retVal = IValue{value: v}
	if v == nil {
		retVal.kind = NoneVal
		retVal.name = "None"
		return retVal
	}

	switch reflect.TypeOf(v).Kind().String() {
	case "Tensor":
		retVal.kind = TensorVal
		retVal.name = "Tensor"
	case "float64":
		retVal.kind = DoubleVal
		retVal.name = "Double"
	case "float32":
		retVal.kind = GenericVal
		retVal.name = "Generic"
	case "int64":
		retVal.kind = IntVal
		retVal.name = "Int"
	case "int":
		retVal.kind = GenericVal
		retVal.name = "Generic"
	case "int32":
		retVal.kind = GenericVal
		retVal.name = "Generic"
	case "bool":
		retVal.kind = BoolVal
		retVal.name = "Bool"
	case "string":
		retVal.kind = StringVal
		retVal.name = "String"
	case "slice":
		switch reflect.TypeOf(v).Elem().Kind().String() {
		case "IValue":
			switch len(v.([]IValue)) {
			case 2:
				retVal.kind = TupleVal
				retVal.name = "Tuple"
			default:
				retVal.kind = GenericListVal
				retVal.name = "GenericList"
			}
		case "Tensor":
			retVal.kind = TensorListVal
			retVal.name = "TensorList"
		case "int64":
			retVal.kind = IntListVal
			retVal.name = "IntList"
		case "float64":
			retVal.kind = DoubleListVal
			retVal.name = "DoubleList"
		case "float32":
			retVal.kind = GenericListVal
			retVal.name = "GenericList"
		case "int32":
			retVal.kind = GenericListVal
			retVal.name = "GenericList"
		case "int":
			retVal.kind = GenericListVal
			retVal.name = "GenericList"
		case "string":
			retVal.kind = GenericListVal
			retVal.name = "GenericList"
		case "bool":
			retVal.kind = BoolListVal
			retVal.name = "BoolList"
		}
	case "map":
		// TODO: exclude map of type other than IValue type
		retVal.kind = GenericDictVal
		retVal.name = "GenericDict"
	default:
		log.Fatalf("NewIValue method call - Unsupport type(%v)\n", reflect.TypeOf(v).Kind().String())
	}

	return retVal
}

// IValue methods:
// ===============

func (iv IValue) ToCIValue() (retVal CIValue, err error) {

	switch iv.name {
	case "None":
		cval := lib.AtiNone()
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		return CIValue{civalue: cval}, nil

	case "Tensor":
		cval := lib.AtiTensor(iv.value.(Tensor).ctensor)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "Int":
		cval := lib.AtiInt(iv.value.(int64))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		return CIValue{civalue: cval}, nil

	case "Double":
		cval := lib.AtiDouble(iv.value.(float64))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "Bool":
		cval := lib.AtiBool(iv.value.(bool))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "Tuple":
		var v []IValue = iv.value.([]IValue)
		var cvals []lib.Civalue
		for _, i := range v {
			cval, err := i.ToCIValue()
			if err != nil {
				err = fmt.Errorf("ToCIValue method call err - Tuple case: %v\n", err)
				return retVal, err
			}
			cvals = append(cvals, cval.civalue)
		}

		tuple := lib.AtiTuple(cvals, len(cvals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: tuple}, nil

	case "GenericList":
		// GenericList can be: string, int, int32, float32
		// TODO: refactor to function
		// NOTE: atm, `GenericList` are all unsupported cases
		var cvals []lib.Civalue
		vtyp := reflect.TypeOf(iv.value).Elem().Kind().String()
		switch vtyp {
		case "string":
			var v []string = iv.value.([]string)
			for _, i := range v {
				ival := NewIValue(i)
				cval, err := ival.ToCIValue()
				if err != nil {
					err = fmt.Errorf("ToCIValue method call err - GenericList case: %v\n", err)
					return retVal, err
				}
				cvals = append(cvals, cval.civalue)
			}

		case "int":
			var v []int = iv.value.([]int)
			for _, i := range v {
				ival := NewIValue(i)
				cval, err := ival.ToCIValue()
				if err != nil {
					log.Fatalf("ToCIValue method call err - int case: %v\n", err)
				}
				cvals = append(cvals, cval.civalue)
			}
		case "int32":
			var v []int32 = iv.value.([]int32)
			for _, i := range v {
				ival := NewIValue(i)
				cval, err := ival.ToCIValue()
				if err != nil {
					log.Fatalf("ToCIValue method call err - int32 case: %v\n", err)
				}
				cvals = append(cvals, cval.civalue)
			}
		case "float32":
			var v []float32 = iv.value.([]float32)
			for _, i := range v {
				ival := NewIValue(i)
				cval, err := ival.ToCIValue()
				if err != nil {
					log.Fatalf("ToCIValue method call err - float32 case: %v\n", err)
				}
				cvals = append(cvals, cval.civalue)
			}
		default:
			log.Fatalf("ToCIValue method call err - Default case: Unsupport type (%v)\n", vtyp)

		}

		list := lib.AtiGenericList(cvals, len(cvals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: list}, nil

	case "IntList":
		var vals []int64 = iv.value.([]int64)
		cval := lib.AtiIntList(vals, len(vals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "DoubleList":
		var vals []float64 = iv.value.([]float64)
		cval := lib.AtiDoubleList(vals, len(vals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "BoolList":
		var vals []bool = iv.value.([]bool)
		cval := lib.AtiBoolList(vals, len(vals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "TensorList":
		var vals []Tensor = iv.value.([]Tensor)
		var cvals []lib.Ctensor
		for _, i := range vals {
			cvals = append(cvals, i.ctensor)
		}
		list := lib.AtiTensorList(cvals, len(cvals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: list}, nil

	case "String":
		cval := lib.AtiString(iv.value.(string))
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: cval}, nil

	case "GenericDict":
		var cvals []lib.Civalue
		keyType := reflect.TypeOf(iv.value).Key().Kind().String()
		valType := reflect.TypeOf(iv.value).Elem().Kind().String()

		// 1. Create key and value lists seperately
		switch {
		case keyType == "int64" && valType == "int64":
			var m map[int64]int64 = iv.value.(map[int64]int64)
			var vals []int64
			for k, v := range m {
				vals = append(vals, k, v)
			}
			for _, v := range vals {
				ival := NewIValue(v)
				cval, err := ival.ToCIValue()
				if err != nil {
					log.Fatalf("ToCIValue method call err - GenericDict case: %v\n", err)
				}
				cvals = append(cvals, cval.civalue)
			}

		case keyType == "float64" && valType == "float64":
			var m map[float64]float64 = iv.value.(map[float64]float64)
			var vals []float64
			for k, v := range m {
				vals = append(vals, k, v)
			}
			for _, v := range vals {
				ival := NewIValue(v)
				cval, err := ival.ToCIValue()
				if err != nil {
					log.Fatalf("ToCIValue method call err - GenericDict case: %v\n", err)
				}
				cvals = append(cvals, cval.civalue)
			}

		case keyType == "float32" && valType == "float32":
			var m map[float32]float32 = iv.value.(map[float32]float32)
			var vals []float32
			for k, v := range m {
				vals = append(vals, k, v)
			}
			for _, v := range vals {
				ival := NewIValue(v)
				cval, err := ival.ToCIValue()
				if err != nil {
					log.Fatalf("ToCIValue method call err - GenericDict case: %v\n", err)
				}
				cvals = append(cvals, cval.civalue)
			}

		// TODO: map[int64]Tensor
		// TODO: map[float64]Tensor
		// TODO: map[string]Tensor
		// TODO: map[bool]Tensor
		// ...

		default:
			log.Fatalf("ToCIValue method call - GenericDict case: unsupported key type(%v) or value type(%v) \n", keyType, valType)
		}

		// 2. Pairing key and value in a slice (cvals)
		dict := lib.AtiGenericDict(cvals, len(cvals)/2)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		return CIValue{civalue: dict}, nil

	case "Generic":
		log.Fatalf("ToCIValue method call - Generic case: unsupport type(%v)\n", reflect.TypeOf(iv.value).Kind().String())

	default:
		log.Fatalf("ToCIValue method call - Generic case: unsupport type(%v)\n", reflect.TypeOf(iv.value).Kind().String())
	}

	return retVal, nil
}

// IValueFromC returns an IValue from given CIValue.
//
// It consumes the pointer and frees the associated memory.
func IValueFromC(cval CIValue) (retVal IValue, err error) {

	// tag will be a value of int32
	tag := lib.AtiTag(cval.civalue)
	if err = TorchErr(); err != nil {
		return retVal, err
	}

	switch tag {
	case 0:
		retVal = IValue{
			value: nil,
			kind:  NoneVal,
			name:  "None",
		}
	case 1:
		tensor := lib.AtiToTensor(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		retVal = IValue{
			value: tensor,
			kind:  TensorVal,
			name:  "Tensor",
		}
	case 2:
		v := lib.AtiToDouble(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		retVal = IValue{
			value: v,
			kind:  DoubleVal,
			name:  "Double",
		}
	case 3:
		v := lib.AtiToInt(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		retVal = IValue{
			value: v,
			kind:  IntVal,
			name:  "Int",
		}

	case 4:
		v := lib.AtiToBool(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		retVal = IValue{
			value: v,
			kind:  BoolVal,
			name:  "Bool",
		}

	case 5: // Tuple []IValue 2 elements
		// 1. Determine tuple length
		len := lib.AtiTupleLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		// 2. Call with first pointer and length
		ptr1 := (*lib.Civalue)(unsafe.Pointer(C.malloc(0)))
		lib.AtiToTuple(cval.civalue, ptr1, int(len))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get list of Civalue tuple elements
		var civalues []CIValue
		civalues = append(civalues, CIValue{civalue: *ptr1})
		currPtr := ptr1
		for i := 1; i < int(len); i++ {
			nextPtr := (*lib.Civalue)(unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1)))
			civalues = append(civalues, CIValue{civalue: *nextPtr})
			currPtr = nextPtr
		}

		// 4. Get Ivalue from Civalue for each tuple element
		var vals []interface{}
		for _, civalue := range civalues {
			v, err := IValueFromC(civalue)
			if err != nil {
				return retVal, err
			}
			vals = append(vals, v)
		}

		retVal = IValue{
			value: vals,
			kind:  TupleVal,
			name:  "Tuple",
		}

	case 6: // IntList
		// 1. Len
		len := lib.AtiLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 2. Call
		ptr1 := unsafe.Pointer(C.malloc(0))
		lib.AtiToIntList(cval.civalue, ptr1, int(len))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get int list
		var intVals []int64
		intVals = append(intVals, *(*int64)(unsafe.Pointer(ptr1)))
		currPtr := ptr1
		for i := 1; i < int(len); i++ {
			nextPtr := unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1))
			intVals = append(intVals, *(*int64)(unsafe.Pointer(nextPtr)))
			currPtr = nextPtr
		}

		retVal = IValue{
			value: intVals,
			kind:  IntListVal,
			name:  "IntList",
		}

	case 7: // DoubleList
		// 1. Len
		len := lib.AtiLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 2. Call
		ptr1 := unsafe.Pointer(C.malloc(0))
		lib.AtiToDoubleList(cval.civalue, ptr1, int(len))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get int list
		var floatVals []float64
		floatVals = append(floatVals, *(*float64)(unsafe.Pointer(ptr1)))
		currPtr := ptr1
		for i := 1; i < int(len); i++ {
			nextPtr := unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1))
			floatVals = append(floatVals, *(*float64)(unsafe.Pointer(nextPtr)))
			currPtr = nextPtr
		}

		retVal = IValue{
			value: floatVals,
			kind:  DoubleListVal,
			name:  "DoubleList",
		}

	case 8: // BoolList
		// 1. Len
		len := lib.AtiLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 2. Call
		ptr1 := unsafe.Pointer(C.malloc(0))
		lib.AtiToBoolList(cval.civalue, ptr1, int(len))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get values
		var vals []int32
		var bvals []bool
		vals = append(vals, *(*int32)(unsafe.Pointer(ptr1)))
		currPtr := ptr1
		for i := 1; i < int(len); i++ {
			nextPtr := unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1))
			vals = append(vals, *(*int32)(unsafe.Pointer(nextPtr)))
			currPtr = nextPtr
		}

		for _, i := range vals {
			bval := false
			if i == 1 {
				bval = true
			}
			bvals = append(bvals, bval)
		}

		retVal = IValue{
			value: bvals,
			kind:  BoolListVal,
			name:  "BoolList",
		}

	case 9: // String
		v := lib.AtiToString(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		retVal = IValue{
			value: v,
			kind:  StringVal,
			name:  "String",
		}

	case 10: // TensorList
		// 1. Len
		len := lib.AtiLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 2. Call
		ptr1 := (*lib.Ctensor)(unsafe.Pointer(C.malloc(0)))
		lib.AtiToTensorList(cval.civalue, ptr1, int(len))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get values
		var tensors []Tensor
		tensors = append(tensors, Tensor{ctensor: *ptr1})
		currPtr := ptr1
		for i := 1; i < int(len); i++ {
			nextPtr := (*lib.Ctensor)(unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1)))
			tensors = append(tensors, Tensor{ctensor: *nextPtr})
			currPtr = nextPtr
		}

		retVal = IValue{
			value: tensors,
			kind:  TensorListVal,
			name:  "TensorList",
		}

	case 12: // GenericList []IValue
		// NOTE: atm, all these cases are unsupported.
		// 1. Len
		len := lib.AtiLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		// 2. Call with first pointer and length
		ptr1 := (*lib.Civalue)(unsafe.Pointer(C.malloc(0)))
		lib.AtiToGenericList(cval.civalue, ptr1, int(len))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get values
		var civalues []CIValue
		civalues = append(civalues, CIValue{civalue: *ptr1})
		currPtr := ptr1
		for i := 1; i < int(len); i++ {
			nextPtr := (*lib.Civalue)(unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1)))
			civalues = append(civalues, CIValue{civalue: *nextPtr})
			currPtr = nextPtr
		}

		// 4. Get Ivalue from Civalue for each tuple element
		var vals []interface{}
		var itemTyp string
		for _, civalue := range civalues {
			v, err := IValueFromC(civalue)
			if err != nil {
				return retVal, err
			}
			itemTyp = reflect.TypeOf(v.value).Kind().String()
			vals = append(vals, v.value)
		}

		switch itemTyp {
		case "string":
			var specVals []string
			for _, v := range vals {
				specVals = append(specVals, v.(string))
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericListVal,
				name:  "GenericList",
			}
		case "int":
			var specVals []int
			for _, v := range vals {
				specVals = append(specVals, v.(int))
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericListVal,
				name:  "GenericList",
			}
		case "int32":
			var specVals []int32
			for _, v := range vals {
				specVals = append(specVals, v.(int32))
			}
			retVal = IValue{
				value: vals,
				kind:  GenericListVal,
				name:  "GenericList",
			}
		case "float32":
			var specVals []float32
			for _, v := range vals {
				specVals = append(specVals, v.(float32))
			}
			retVal = IValue{
				value: vals,
				kind:  GenericListVal,
				name:  "GenericList",
			}
			return retVal, nil

		default:
			log.Fatalf("IValueFromC method call - GenericList case: Unsupported item type (%v)\n", itemTyp)
		}

	case 13: // GenericDict map[IValue]IValue
		// 1. Len
		numVals := lib.AtiLength(cval.civalue)
		if err = TorchErr(); err != nil {
			return retVal, err
		}
		// 2. Call with first pointer and length
		ptr1 := (*lib.Civalue)(unsafe.Pointer(C.malloc(0)))
		lib.AtiToGenericDict(cval.civalue, ptr1, int(numVals))
		if err = TorchErr(); err != nil {
			return retVal, err
		}

		// 3. Get values

		// TODO: Need to drill down a specific type
		var civalues []CIValue
		civalues = append(civalues, CIValue{civalue: *ptr1})
		currPtr := ptr1
		for i := 1; i < int(numVals)*2; i++ {
			nextPtr := (*lib.Civalue)(unsafe.Pointer(uintptr(unsafe.Pointer(currPtr)) + unsafe.Sizeof(ptr1)))
			civalues = append(civalues, CIValue{civalue: *nextPtr})
			currPtr = nextPtr
		}

		// 4. Get Ivalue from Civalue for each element
		var vals []interface{}
		var itemTyp string
		for _, civalue := range civalues {
			v, err := IValueFromC(civalue)
			if err != nil {
				return retVal, err
			}
			itemTyp = reflect.TypeOf(v.value).Kind().String()
			vals = append(vals, v.value)
		}

		switch itemTyp {
		case "string":
			var specVals map[string]string = make(map[string]string)
			for i := 0; i < len(vals); i += 2 {
				specVals[vals[i].(string)] = vals[i+1].(string)
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericDictVal,
				name:  "GenericDict",
			}
		case "int":
			var specVals map[int]int = make(map[int]int)
			for i := 0; i < len(vals); i += 2 {
				specVals[vals[i].(int)] = vals[i+1].(int)
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericDictVal,
				name:  "GenericDict",
			}
		case "int32":
			var specVals map[int32]int32 = make(map[int32]int32)
			for i := 0; i < len(vals); i += 2 {
				specVals[vals[i].(int32)] = vals[i+1].(int32)
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericDictVal,
				name:  "GenericDict",
			}
		case "int64":
			var specVals map[int64]int64 = make(map[int64]int64)
			for i := 0; i < len(vals); i += 2 {
				specVals[vals[i].(int64)] = vals[i+1].(int64)
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericDictVal,
				name:  "GenericDict",
			}
		case "float32":
			var specVals map[float32]float32 = make(map[float32]float32)
			for i := 0; i < len(vals); i += 2 {
				specVals[vals[i].(float32)] = vals[i+1].(float32)
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericDictVal,
				name:  "GenericDict",
			}
			return retVal, nil
		case "float64":
			var specVals map[float64]float64 = make(map[float64]float64)
			for i := 0; i < len(vals); i += 2 {
				specVals[vals[i].(float64)] = vals[i+1].(float64)
			}
			retVal = IValue{
				value: specVals,
				kind:  GenericDictVal,
				name:  "GenericDict",
			}
			return retVal, nil
		}

	default:
		log.Fatalf("IValueFromC - Unsupported type (tag value: %v)\n", tag)
	}

	return
}