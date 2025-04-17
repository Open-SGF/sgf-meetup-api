package customconstructs

import "github.com/aws/jsii-runtime-go"

type FunctionName struct {
	prefix string
	name   string
}

func NewFunctionName(prefix, name string) *FunctionName {
	return &FunctionName{
		prefix: prefix,
		name:   name,
	}
}

func (f *FunctionName) Name() *string {
	return jsii.String(f.name)
}

func (f *FunctionName) PrefixedName() *string {
	return jsii.String(f.prefix + f.name)
}
