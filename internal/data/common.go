// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data

// This file contains data structures that are common for all telemetry types,
// such as timestamps, attributes, etc.

import (
	"sort"

	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
)

// TimestampUnixNano is a time specified as UNIX Epoch time in nanoseconds since
// 00:00:00 UTC on 1 January 1970.
type TimestampUnixNano uint64

// AttributeValueType specifies the type of value. Numerically is equal to
// otlp.AttributeKeyValue_ValueType.
type AttributeValueType int32

const (
	AttributeValueSTRING AttributeValueType = AttributeValueType(otlpcommon.AttributeKeyValue_STRING)
	AttributeValueINT    AttributeValueType = AttributeValueType(otlpcommon.AttributeKeyValue_INT)
	AttributeValueDOUBLE AttributeValueType = AttributeValueType(otlpcommon.AttributeKeyValue_DOUBLE)
	AttributeValueBOOL   AttributeValueType = AttributeValueType(otlpcommon.AttributeKeyValue_BOOL)
)

// AttributeValue represents a value of an attribute. Typically used in an Attributes map.
// Must use one of NewAttributeValue* functions below to create new instances.
// Important: zero-initialized instance is not valid for use.
//
// Intended to be passed by value since internally it is just a pointer to actual
// value representation. For the same reason passing by value and calling setters
// will modify the original, e.g.:
//
//   function f1(val AttributeValue) { val.SetInt(234) }
//   function f2() {
//   	v := NewAttributeValueString("a string")
//      f1(v)
//      _ := v.GetType() // this will return AttributeValueINT
//   }
type AttributeValue struct {
	orig *otlpcommon.AttributeKeyValue
}

func NewAttributeValueString(v string) AttributeValue {
	return AttributeValue{orig: &otlpcommon.AttributeKeyValue{Type: otlpcommon.AttributeKeyValue_STRING, StringValue: v}}
}

func NewAttributeValueInt(v int64) AttributeValue {
	return AttributeValue{orig: &otlpcommon.AttributeKeyValue{Type: otlpcommon.AttributeKeyValue_INT, IntValue: v}}
}

func NewAttributeValueDouble(v float64) AttributeValue {
	return AttributeValue{orig: &otlpcommon.AttributeKeyValue{Type: otlpcommon.AttributeKeyValue_DOUBLE, DoubleValue: v}}
}

func NewAttributeValueBool(v bool) AttributeValue {
	return AttributeValue{orig: &otlpcommon.AttributeKeyValue{Type: otlpcommon.AttributeKeyValue_BOOL, BoolValue: v}}
}

// NewAttributeValueSlice creates a slice of attributes values that are correctly initialized.
func NewAttributeValueSlice(len int) []AttributeValue {
	// Allocate 2 slices, one for AttributeValues, another for underlying OTLP structs.
	// TODO: make one allocation for both slices.
	origs := make([]otlpcommon.AttributeKeyValue, len)
	wrappers := make([]AttributeValue, len)
	for i := range origs {
		wrappers[i].orig = &origs[i]
	}
	return wrappers
}

// All AttributeValue functions bellow must be called only on instances that are created
// via NewAttributeValue* functions. Calling these functions on zero-initialized
// AttributeValue struct will cause a panic.

func (a AttributeValue) Type() AttributeValueType {
	return AttributeValueType(a.orig.Type)
}

func (a AttributeValue) StringVal() string {
	return a.orig.StringValue
}

func (a AttributeValue) IntVal() int64 {
	return a.orig.IntValue
}

func (a AttributeValue) DoubleVal() float64 {
	return a.orig.DoubleValue
}

func (a AttributeValue) BoolVal() bool {
	return a.orig.BoolValue
}

func (a AttributeValue) SetString(v string) {
	a.orig.Type = otlpcommon.AttributeKeyValue_STRING
	a.orig.StringValue = v
}

func (a AttributeValue) SetInt(v int64) {
	a.orig.Type = otlpcommon.AttributeKeyValue_INT
	a.orig.IntValue = v
}

func (a AttributeValue) SetDouble(v float64) {
	a.orig.Type = otlpcommon.AttributeKeyValue_DOUBLE
	a.orig.DoubleValue = v
}

func (a AttributeValue) SetBool(v bool) {
	a.orig.Type = otlpcommon.AttributeKeyValue_BOOL
	a.orig.BoolValue = v
}

// AttributesMap stores a map of attribute keys to values.
type AttributesMap map[string]AttributeValue

// Attributes stores the map of attributes and a number of dropped attributes.
// Typically used by translator functions to easily pass the pair.
type Attributes struct {
	attrs        AttributesMap
	droppedCount uint32
}

func NewAttributes(m AttributesMap, droppedCount uint32) Attributes {
	return Attributes{m, droppedCount}
}

// StringKeyValue stores a key and value pair.
type StringKeyValue struct {
	orig *otlpcommon.StringKeyValue
}

// NewStringKeyValue creates a new StringKeyValue with the given key and value.
func NewStringKeyValue(k string, v string) StringKeyValue {
	return StringKeyValue{&otlpcommon.StringKeyValue{Key: k, Value: v}}
}

// Key returns the key associated with this StringKeyValue.
func (akv StringKeyValue) Key() string {
	return akv.orig.Key
}

// Value returns the value associated with this StringKeyValue.
func (akv StringKeyValue) Value() string {
	return akv.orig.Value
}

// SetValue replaces the value associated with this StringKeyValue.
func (akv StringKeyValue) SetValue(v string) {
	akv.orig.Value = v
}

// StringMap stores a map of attribute keys to values.
type StringMap struct {
	orig *[]*otlpcommon.StringKeyValue
}

func newStringMap(orig *[]*otlpcommon.StringKeyValue) StringMap {
	return StringMap{orig}
}

// NewStringMap creates a new StringMap from the given map[string]string.
func NewStringMap(attrMap map[string]string) StringMap {
	if len(attrMap) == 0 {
		var orig []*otlpcommon.StringKeyValue
		return StringMap{&orig}
	}
	origs := make([]otlpcommon.StringKeyValue, len(attrMap))
	wrappers := make([]*otlpcommon.StringKeyValue, len(attrMap))

	ix := 0
	for k, v := range attrMap {
		wrappers[ix] = &origs[ix]
		wrappers[ix].Key = k
		wrappers[ix].Value = v
		ix++
	}

	return StringMap{&wrappers}
}

// Get returns the StringKeyValue associated with the key and true,
// otherwise an invalid instance of the StringKeyValue and false.
func (sm StringMap) Get(k string) (StringKeyValue, bool) {
	for _, a := range *sm.orig {
		if a.Key == k {
			return StringKeyValue{a}, true
		}
	}
	return StringKeyValue{nil}, false
}

// Delete deletes the entry associated with the key and returns true if the key
// was present in the map, otherwise returns false.
func (sm StringMap) Delete(k string) bool {
	for i, a := range *sm.orig {
		if a.Key == k {
			(*sm.orig)[i] = (*sm.orig)[len(*sm.orig)-1]
			*sm.orig = (*sm.orig)[:len(*sm.orig)-1]
			return true
		}
	}
	return false
}

// Insert adds the StringKeyValue to the map when the key does not exist.
// No action is applied to the map where the key already exists.
func (sm StringMap) Insert(k, v string) {
	if _, existing := sm.Get(k); !existing {
		*sm.orig = append(*sm.orig, NewStringKeyValue(k, v).orig)
	}
}

// Update updates an existing StringKeyValue with a value.
// No action is applied to the map where the key does not exist.
func (sm StringMap) Update(k, v string) {
	if av, existing := sm.Get(k); existing {
		av.SetValue(v)
	}
}

// Upsert performs the Insert or Update action. The StringKeyValue is
// insert to the map that did not originally have the key. The key/value is
// updated to the map where the key already existed.
func (sm StringMap) Upsert(k, v string) {
	if av, existing := sm.Get(k); existing {
		av.SetValue(v)
	} else {
		*sm.orig = append(*sm.orig, NewStringKeyValue(k, v).orig)
	}
}

// Len returns the number of StringKeyValue in the map.
func (sm StringMap) Len() int {
	return len(*sm.orig)
}

// GetStringKeyValue returns the StringKeyValue associated with the given index.
//
// This function is used mostly for iterating over all the values in the map:
// for i := 0; i < am.Len(); i++ {
//     akv := am.GetStringKeyValue(i)
//     ... // Do something with the attribute
// }
func (sm StringMap) GetStringKeyValue(ix int) StringKeyValue {
	return StringKeyValue{(*sm.orig)[ix]}
}

// Sort sorts the entries in the StringMap so two instances can be compared.
// Returns the same instance to allow nicer code like:
// assert.EqualValues(t, expected.Sort(), actual.Sort())
func (sm StringMap) Sort() StringMap {
	sort.SliceStable(*sm.orig, func(i, j int) bool { return (*sm.orig)[i].Key < (*sm.orig)[j].Key })
	return sm
}

// InstrumentationLibrary is a message representing the instrumentation library information.
//
// Must use NewResource functions to create new instances.
// Important: zero-initialized instance is not valid for use.
type InstrumentationLibrary struct {
	orig *otlpcommon.InstrumentationLibrary
}

// NewInstrumentationLibrary creates a new InstrumentationLibrary.
func NewInstrumentationLibrary() InstrumentationLibrary {
	return InstrumentationLibrary{}
}

func newInstrumentationLibrary(orig *otlpcommon.InstrumentationLibrary) InstrumentationLibrary {
	return InstrumentationLibrary{orig}
}

func (il InstrumentationLibrary) Name() string {
	return il.orig.Name
}

func (il InstrumentationLibrary) SetName(r string) {
	il.orig.Name = r
}

func (il InstrumentationLibrary) Version() string {
	return il.orig.Version
}

func (il InstrumentationLibrary) SetVersion(r string) {
	il.orig.Version = r
}
