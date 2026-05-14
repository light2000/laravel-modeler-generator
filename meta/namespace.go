package meta

import "github.com/light2000/laravel-modeler-generator/proto"

type Namespace struct {
	*proto.Namespace
	// Add your custom fields below
	Project           *Project
	ModelOutputPath   string
	FactoryOutputPath string
	SeederOutputPath  string
	EnumOutputPath    string
}

func FromProtoNamespace(p *proto.Namespace, project *Project) *Namespace {
	if p == nil {
		return nil
	}
	return &Namespace{
		Namespace: p,
		Project:   project,
	}
}
