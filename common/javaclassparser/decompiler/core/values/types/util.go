package types

import (
	"fmt"
	"strings"
)

func GetPrimerArrayType(id int) JavaType {
	switch id {
	case 4:
		return JavaBoolean
	case 5:
		return JavaChar
	case 6:
		return JavaFloat
	case 7:
		return JavaDouble
	case 8:
		return JavaByte
	case 9:
		return JavaShort
	case 10:
		return JavaInteger
	case 11:
		return JavaLong
	default:
		panic(fmt.Sprintf("unknow primer array type: %d", id))
	}
}
func ParseDescriptor(descriptor string) (JavaType, error) {
	returnType, _, err := ParseJavaDescription(descriptor)
	return returnType, err
}

// parseMethodDescriptor 解析 Java 方法描述符
func ParseMethodDescriptor(descriptor string) (*JavaFuncType, error) {
	if descriptor == "" {
		return nil, fmt.Errorf("descriptor is empty")
	}

	if descriptor[0] != '(' {
		return nil, fmt.Errorf("invalid descriptor format")
	}

	// 查找参数部分和返回类型部分
	endIndex := strings.Index(descriptor, ")")
	if endIndex == -1 {
		return nil, fmt.Errorf("invalid descriptor format")
	}

	paramDescriptor := descriptor[1:endIndex]
	returnTypeDescriptor := descriptor[endIndex+1:]

	// 解析参数类型
	paramTypes, err := parseTypes(paramDescriptor)
	if err != nil {
		return nil, err
	}

	// 解析返回类型
	returnType, _, err := ParseJavaDescription(returnTypeDescriptor)
	if err != nil {
		return nil, err
	}

	return NewJavaFuncType(descriptor, paramTypes, returnType), nil
}

// parseTypes 解析多个类型描述符
func parseTypes(descriptor string) ([]JavaType, error) {
	var types []JavaType
	for len(descriptor) > 0 {
		t, rest, err := ParseJavaDescription(descriptor)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
		descriptor = rest
	}
	return types, nil
}
func parseFuncType(desc string) (*JavaFuncType, string, error) {
	if desc == "" {
		return nil, "", fmt.Errorf("descriptor is empty")
	}
	if desc[0] != '(' {
		return nil, "", fmt.Errorf("invalid descriptor format")
	}
	endIndex := strings.Index(desc, ")")
	if endIndex == -1 {
		return nil, "", fmt.Errorf("invalid descriptor format")
	}
	paramDesc := desc[1:endIndex]
	returnDesc := desc[endIndex+1:]
	params, err := parseTypes(paramDesc)
	if err != nil {
		return nil, "", err
	}
	returnType, _, err := ParseJavaDescription(returnDesc)
	if err != nil {
		return nil, "", err
	}
	return NewJavaFuncType(desc, params, returnType), "", nil
}

// ParseJavaDescription 解析单个类型描述符
func ParseJavaDescription(descriptor string) (JavaType, string, error) {
	if len(descriptor) == 0 {
		return nil, "", fmt.Errorf("empty descriptor")
	}

	switch descriptor[0] {
	case 'B':
		return JavaByte, descriptor[1:], nil
	case 'C':
		return JavaChar, descriptor[1:], nil
	case 'D':
		return JavaDouble, descriptor[1:], nil
	case 'F':
		return JavaFloat, descriptor[1:], nil
	case 'I':
		return JavaInteger, descriptor[1:], nil
	case 'J':
		return JavaLong, descriptor[1:], nil
	case 'S':
		return JavaShort, descriptor[1:], nil
	case 'Z':
		return JavaBoolean, descriptor[1:], nil
	case 'V':
		return JavaVoid, descriptor[1:], nil
	case 'L':
		// 类类型，以 L 开头，以 ; 结尾
		endIndex := strings.Index(descriptor, ";")
		if endIndex == -1 {
			return nil, "", fmt.Errorf("invalid class descriptor format")
		}
		name := strings.Replace(descriptor[1:endIndex], "/", ".", -1)
		return NewJavaClass(name), descriptor[endIndex+1:], nil
	case '[':
		// 数组类型，以 [ 开头，后跟元素类型
		elemType, rest, err := ParseJavaDescription(descriptor[1:])
		if err != nil {
			return nil, "", err
		}
		switch ret := elemType.(type) {
		case *JavaArrayType:
			ret.Length = append(ret.Length)
			return ret, rest, nil
		default:
			return NewJavaArrayType(elemType), rest, nil
		}
	default:
		return nil, "", fmt.Errorf("unknown type descriptor: %c", descriptor[0])
	}
}