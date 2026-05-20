package meta

import (
	"strings"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/proto"
)

type Module struct {
	*proto.Module
	Items []*Item
	// Add your custom fields below
	Project   *Project
	Namespace *Namespace
}

func FromProtoModule(p *proto.Module, project *Project) *Module {
	if p == nil {
		return nil
	}
	mustSnakeCode("FromProtoModule", p.Id, p.Name, p.Code)
	m := &Module{
		Module:  p,
		Project: project,
	}
	if p.NamespaceId != "" {
		fromNamespace := project.MapNamespaces[p.NamespaceId]
		m.Namespace = &Namespace{
			Namespace: &proto.Namespace{
				Id:                  fromNamespace.Id,
				Name:                fromNamespace.Name,
				ModelNamespace:      fromNamespace.ModelNamespace,
				MigrationOutputPath: fromNamespace.MigrationOutputPath,
				FactoryNamespace:    fromNamespace.FactoryNamespace,
				SeederNamespace:     fromNamespace.SeederNamespace,
				EnumNamespace:       fromNamespace.EnumNamespace,
			},
			Project: project,
		}
		m.Namespace.ModelNamespace = strings.ReplaceAll(fromNamespace.ModelNamespace, "{Module}", m.Class())
		m.Namespace.MigrationOutputPath = strings.ReplaceAll(fromNamespace.MigrationOutputPath, "{Module}", m.Class())
		m.Namespace.FactoryNamespace = strings.ReplaceAll(fromNamespace.FactoryNamespace, "{Module}", m.Class())
		m.Namespace.SeederNamespace = strings.ReplaceAll(fromNamespace.SeederNamespace, "{Module}", m.Class())
		m.Namespace.EnumNamespace = strings.ReplaceAll(fromNamespace.EnumNamespace, "{Module}", m.Class())
		m.Namespace.ModelOutputPath = project.ResolveOutputPathByNamespace(m.Namespace.ModelNamespace)
		m.Namespace.FactoryOutputPath = project.ResolveOutputPathByNamespace(m.Namespace.FactoryNamespace)
		m.Namespace.SeederOutputPath = project.ResolveOutputPathByNamespace(m.Namespace.SeederNamespace)
		m.Namespace.EnumOutputPath = project.ResolveOutputPathByNamespace(m.Namespace.EnumNamespace)
	}
	project.MapModules[p.Id] = m

	if p.Items != nil {
		m.Items = make([]*Item, 0, len(p.Items))
		for _, it := range p.Items {
			item := FromProtoItem(it, m, project, false)
			m.Items = append(m.Items, item)
		}
	}

	return m
}

func (m *Module) Snake() string {
	return m.Code
}

func (m *Module) Var() string {
	return common.ToLowerCamel(m.Snake())
}

func (m *Module) Class() string {
	return common.ToCamel(m.Snake())
}
