package meta

import (
	"fmt"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/proto"
)

type Relation struct {
	*proto.Relation
	// Add your custom fields below
	Project *Project
	Module  *Module
	Item    *Item
}

func FromProtoRelation(p *proto.Relation, item *Item, module *Module, project *Project) *Relation {
	if p == nil {
		return nil
	}
	if p.Type != proto.RelationType_ITEM_RELATION_TYPE_MORPHED_BY_MANY {
		mustSnakeCode("FromProtoRelation", p.Id, p.Code)
	}

	if len(p.MorphTargets) > 0 {
		for i, mt := range p.MorphTargets {
			if mt == nil {
				panic(fmt.Sprintf("meta.FromProtoRelation: MorphTargets[%d] 为 nil (relation id=%s)", i, p.Id))
			}
			mustSnakeCode("FromProtoRelation.MorphTarget", fmt.Sprintf("%s[%d]", p.Id, i), mt.Code)
		}
	}
	relation := &Relation{
		Relation: p,
		Project:  project,
		Module:   module,
		Item:     item,
	}

	project.MapRelations[p.Id] = relation

	if relation.IsInternal && relation.PivotId != "" {
		project.MapItems[relation.PivotId].Module = module
	}

	return relation
}

func (relation *Relation) IsSelfRelation() bool {
	return relation.Item.Id == relation.TargetItemId
}

// TargetItem 获取目标 Item
func (relation *Relation) TargetItem() *Item {
	if relation.TargetItemId == "" {
		panic("target item id is empty")
	}
	return relation.Project.MapItems[relation.TargetItemId]
}

// IsBelongsTo source 侧，FK 存储，1v1 或 1vN
func (relation *Relation) IsBelongsTo() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_BELONGS_TO
}

// IsHasOne target 侧，FK 存储，1v1
func (relation *Relation) IsHasOne() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_HAS_ONE
}

// IsHasMany target 侧，FK 存储，1vN
func (relation *Relation) IsHasMany() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_HAS_MANY
}

// IsBelongsToMany source 或 target 侧，PIVOT 存储，NvN
func (relation *Relation) IsBelongsToMany() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_BELONGS_TO_MANY
}

// IsMorphOne morph target 侧，POLYMORPHIC 存储，1v1
func (relation *Relation) IsMorphOne() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_MORPH_ONE
}

// IsMorphMany morph target 侧，POLYMORPHIC 存储，1vN
func (relation *Relation) IsMorphMany() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_MORPH_MANY
}

// IsMorphedByMany morph target 侧，POLYMORPHIC 存储，NvN
// 多对多多态，公共模型侧需要为每个 morph target 生成一个方法
func (relation *Relation) IsMorphToMany() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_MORPH_TO_MANY
}

// IsMorphToMany source 侧，POLYMORPHIC 存储，NvN
func (relation *Relation) IsMorphedByMany() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_MORPHED_BY_MANY
}

// IsMorphTo source 侧持有 morph 字段，POLYMORPHIC 存储，1v1 或 1vN
func (relation *Relation) IsMorphTo() bool {
	return relation.Type == proto.RelationType_ITEM_RELATION_TYPE_MORPH_TO
}

// ===== 获取关联信息 =====

// Method 获取关联方法名
func (relation *Relation) Method() string {
	return common.ToLowerCamel(relation.Snake())
}

func (relation *Relation) Snake() string {
	return relation.Code
}

// GetTargetClass 获取目标类名
func (relation *Relation) GetTargetClass() string {
	if relation.TargetItemId != "" {
		return relation.Project.MapItems[relation.TargetItemId].Class()
	}

	panic("target item id is empty")
}

// GetFKColumn 获取外键列名
func (relation *Relation) GetFKColumn() string {
	return relation.Attribute().Snake()
}

// MorphAble 获取多态字段前缀（去掉 _type）
func (relation *Relation) MorphAble() string {
	if relation.AttrId != "" {
		return relation.Attribute().MorphAble()
	}
	if relation.PivotId != "" {
		return relation.PivotItem().MorphAble()
	}

	panic(fmt.Sprintf("morph able is not supported: %s", relation.Name))
}

// GetPivotTable 获取中间表表名
func (relation *Relation) GetPivotTable() string {
	if relation.PivotId == "" {
		panic("pivot id is empty")
	}
	return relation.PivotItem().Table
}

// GetPivotForeignKey 获取中间表中指向当前 item 的外键列名
func (relation *Relation) GetPivotForeignKey() string {
	return relation.Item.Code + "_id"
}

// GetPivotRelatedKey 获取中间表中指向关联 item 的外键列名
func (relation *Relation) GetPivotRelatedKey() string {
	return relation.TargetItem().Code + "_id"
}

// GetMorphPivotForeignKey 获取多态多对多关系中的 foreignPivotKey
func (relation *Relation) GetMorphPivotForeignKey() string {
	if relation.IsMorphToMany() {
		return relation.GetPivotMorphIdColumn()
	}
	if relation.IsMorphedByMany() {
		return relation.GetPivotFkColumn()
	}
	panic(fmt.Sprintf("GetMorphPivotForeignKey is only supported by morph many-to-many relations: %s", relation.Name))
}

// GetMorphPivotRelatedKey 获取多态多对多关系中的 relatedPivotKey
func (relation *Relation) GetMorphPivotRelatedKey() string {
	if relation.IsMorphToMany() {
		return relation.GetPivotFkColumn()
	}
	if relation.IsMorphedByMany() {
		return relation.GetPivotMorphIdColumn()
	}
	panic(fmt.Sprintf("GetMorphPivotRelatedKey is only supported by morph many-to-many relations: %s", relation.Name))
}

// GetPivotMorphIdColumn 获取中间表中的 morph id 字段名
func (relation *Relation) GetPivotMorphIdColumn() string {
	if relation.PivotId == "" {
		panic("pivot id is empty")
	}
	for _, attr := range relation.PivotItem().Attrs {
		if attr.IsMorphId() {
			return attr.Snake()
		}
	}
	panic(fmt.Sprintf("pivot morph id column not found: relation=%s pivot=%s", relation.Name, relation.PivotItem().Table))
}

// GetPivotFkColumn 获取中间表中的外键字段名（排除 morph 字段）
func (relation *Relation) GetPivotFkColumn() string {
	if relation.PivotId == "" {
		panic("pivot id is empty")
	}
	for _, attr := range relation.PivotItem().Attrs {
		if attr.IsFk() {
			return attr.Snake()
		}
	}
	panic(fmt.Sprintf("pivot fk column not found: relation=%s pivot=%s", relation.Name, relation.PivotItem().Table))
}

func (relation *Relation) RightItems() []*Item {
	items := make([]*Item, 0)
	if relation.TargetItemId != "" {
		items = append(items, relation.Project.MapItems[relation.TargetItemId])
	}
	for _, target := range relation.MorphTargets {
		items = append(items, relation.Project.MapItems[target.TargetItemId])
	}
	return items
}

func (relation *Relation) IsNotSelfRelation() bool {
	return relation.Item.Id != relation.TargetItemId
}

func (relation *Relation) IsManyToManySeederSide() bool {
	if relation.IsReverse && relation.IsBelongsToMany() {
		return true
	}

	if relation.IsMorphToMany() {
		return true
	}

	return false
}

func (relation *Relation) PivotItem() *Item {
	return relation.Project.MapItems[relation.PivotId]
}

// PivotExtraAttrs 获取中间表中除关系字段外的额外属性（排除 fk/morph_id/morph_type）
func (relation *Relation) PivotExtraAttrs() []*Attribute {
	if relation.PivotId == "" {
		return []*Attribute{}
	}

	attrs := make([]*Attribute, 0)
	for _, attr := range relation.PivotItem().Attrs {
		if attr.IsFk() || attr.IsMorphId() || attr.IsMorphType() {
			continue
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// HasPivotExtraAttrs 判断中间表是否存在除关系字段外的额外属性
func (relation *Relation) HasPivotExtraAttrs() bool {
	return len(relation.PivotExtraAttrs()) > 0
}

func (relation *Relation) ReverseRelation() *Relation {
	if relation.ReverseRelationId == "" {
		panic("reverse relation id is empty")
	}
	return relation.Project.MapRelations[relation.ReverseRelationId]
}

func (relation *Relation) Attribute() *Attribute {
	if relation.AttrId == "" {
		panic(fmt.Sprintf("attr id is empty: %s:%s", relation.Item.Name, relation.Name))
	}

	return relation.Project.GetAttributeById(relation.AttrId)
}
