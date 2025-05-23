package yserx

import "github.com/yaklang/yaklang/common/javaclassparser/jarwar"

var Exports = map[string]interface{}{
	"ToJson":                   ToJson,
	"FromJson":                 FromJson,
	"ParseHexJavaObjectStream": ParseHexJavaSerialized,
	"ParseJavaObjectStream":    ParseJavaSerialized,

	"NewJavaNull":             NewJavaNull,
	"NewJavaClass":            NewJavaClass,
	"NewJavaClassFields":      NewJavaClassFields,
	"NewJavaClassField":       NewJavaClassField,
	"NewJavaClassData":        NewJavaClassData,
	"NewJavaClassDesc":        NewJavaClassDesc,
	"NewJavaClassDetails":     NewJavaClassDetails,
	"NewJavaEnum":             NewJavaEnum,
	"NewJavaArray":            NewJavaArray,
	"NewJavaString":           NewJavaString,
	"NewJavaLongString":       NewJavaLongString,
	"NewJavaFieldArrayValue":  NewJavaFieldArrayValue,
	"NewJavaFieldByteValue":   NewJavaFieldByteValue,
	"NewJavaFieldBoolValue":   NewJavaFieldBoolValue,
	"NewJavaFieldCharValue":   NewJavaFieldCharValue,
	"NewJavaFieldDoubleValue": NewJavaFieldDoubleValue,
	"NewJavaFieldFloatValue":  NewJavaFieldFloatValue,
	"NewJavaFieldIntValue":    NewJavaFieldIntValue,
	"NewJavaFieldLongValue":   NewJavaFieldLongValue,
	"NewJavaFieldObjectValue": NewJavaFieldObjectValue,
	"NewJavaFieldShortValue":  NewJavaFieldShortValue,
	"NewJavaFieldValue":       NewJavaFieldValue,
	"NewJavaEndBlockData":     NewJavaEndBlockData,
	"NewJavaBlockDataBytes":   NewJavaBlockDataBytes,
	"NewJavaObject":           NewJavaObject,
	"NewJavaReference":        NewJavaReference,
	"MarshalJavaObjects":      MarshalJavaObjects,

	"Decompile": jarwar.AutoDecompile,
}
