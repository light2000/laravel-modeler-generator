package meta

import (
	"fmt"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/proto"
)

type Option struct {
	*proto.Option
	// Add your custom fields below
	Dict    *Dict
	Project *Project
}

func FromProtoOption(idx int, p *proto.Option, dict *Dict, project *Project) *Option {
	if p == nil {
		return nil
	}
	mustSnakeCode("FromProtoOption", fmt.Sprintf("%s[%d]", dict.Id, idx), p.Code)

	option := &Option{
		Option:  p,
		Dict:    dict,
		Project: project,
	}
	return option
}

func (option *Option) Snake() string {
	return option.Code
}

func (option *Option) Class() string {
	return common.ToCamel(option.Snake())
}
