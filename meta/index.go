package meta

import (
	"sort"
	"strings"

	"github.com/light2000/laravel-modeler-generator/proto"
)

type Index struct {
	*proto.Index
	Attrs []*Attribute
	// Add your custom fields below
	Project *Project
	Item    *Item
}

func FromProtoIndex(p *proto.Index, item *Item, project *Project) *Index {
	if p == nil {
		return nil
	}
	mustSnakeCode("FromProtoIndex", p.Id, item.Name, p.Code)
	index := &Index{
		Index:   p,
		Project: project,
		Item:    item,
	}
	project.MapIndexes[p.Id] = index
	// 处理属性列表
	if p.AttrIds != nil {
		index.Attrs = make([]*Attribute, 0, len(p.AttrIds))
		for _, attrId := range p.AttrIds {
			attr := project.GetAttributeById(attrId)
			if attr != nil {
				index.Attrs = append(index.Attrs, attr)
			}
		}
	}

	return index
}

func (idx *Index) IsUnique() bool {
	return idx.Type == proto.IndexType_INDEX_TYPE_UNIQUE
}

func (idx *Index) IsIndex() bool {
	return idx.Type == proto.IndexType_INDEX_TYPE_INDEX
}

// IndexName 返回索引名称
func (idx *Index) IndexName() string {
	return idx.Code
}

// ColumnNames 返回索引涉及的列名数组
func (idx *Index) ColumnNames() []string {
	columns := make([]string, 0, len(idx.Attrs))
	for _, attr := range idx.Attrs {
		columns = append(columns, attr.Snake())
	}
	sort.Strings(columns)

	return columns
}

func (idx *Index) ColumnNamesString() string {
	return "'" + strings.Join(idx.ColumnNames(), "', '") + "'"
}
