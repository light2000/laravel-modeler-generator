package laravel

import (
	"fmt"

	"github.com/light2000/laravel-modeler-generator/meta"
	"github.com/light2000/laravel-modeler-generator/templatex"
)

const (
	MigrationRenameTable = iota
	MigrationDropTable
	MigrationCreateTable
	MigrationAttrRemove
	MigrationAttrAdd
	MigrationAttrUpdate
	MigrationIndexRemove
	MigrationIndexAdd
	MigrationIndexRename
)

func ItemMigrationBuild(project *meta.Project, engine *templatex.Engine) {
	migrationLoader := engine.NewLoader("laravel/item/Migration.tpl")
	migrationDropTableLoader := engine.NewLoader("laravel/item/MigrationDropTable.tpl")
	migrationAttrRemoveLoader := engine.NewLoader("laravel/item/MigrationAttrRemove.tpl")
	migrationAttrAddLoader := engine.NewLoader("laravel/item/MigrationAttrAdd.tpl")
	migrationAttrUpdateLoader := engine.NewLoader("laravel/item/MigrationAttrUpdate.tpl")
	migrationRenameLoader := engine.NewLoader("laravel/item/MigrationRename.tpl")
	migrationIndexAddLoader := engine.NewLoader("laravel/item/MigrationIndexAdd.tpl")
	migrationIndexRemoveLoader := engine.NewLoader("laravel/item/MigrationIndexRemove.tpl")
	migrationIndexRenameLoader := engine.NewLoader("laravel/item/MigrationIndexRename.tpl")

	migrationBuild(project, migrationLoader, migrationDropTableLoader, migrationAttrRemoveLoader, migrationAttrAddLoader, migrationAttrUpdateLoader, migrationRenameLoader, migrationIndexAddLoader, migrationIndexRemoveLoader, migrationIndexRenameLoader)
	for project.PrevProject != nil {
		project = project.PrevProject
		migrationBuild(project, migrationLoader, migrationDropTableLoader, migrationAttrRemoveLoader, migrationAttrAddLoader, migrationAttrUpdateLoader, migrationRenameLoader, migrationIndexAddLoader, migrationIndexRemoveLoader, migrationIndexRenameLoader)
	}
}

func migrationBuild(project *meta.Project, migrationLoader *templatex.Loader,
	migrationDropTableLoader *templatex.Loader, migrationAttrRemoveLoader *templatex.Loader,
	migrationAttrAddLoader *templatex.Loader, migrationAttrUpdateLoader *templatex.Loader, migrationRenameLoader *templatex.Loader,
	migrationIndexAddLoader *templatex.Loader, migrationIndexRemoveLoader *templatex.Loader, migrationIndexRenameLoader *templatex.Loader) {

	data := map[string]interface{}{
		"Project": project,
	}
	for idx1, mod := range project.Modules {
		data["Module"] = project.Modules[idx1]
		for idx2, item := range mod.Items {
			data["Item"] = project.Modules[idx1].Items[idx2]

			//Migration
			if IsNewItem(item, project.PrevProject) {
				migrationLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_create_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationCreateTable, item.Table), data, true)
			} else if IsRenameItem(item, project.PrevProject) {
				data["PrevItem"] = project.PrevProject.MapItems[item.Id]
				migrationRenameLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_rename_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationRenameTable, item.Table), data, true)
			}
		}
	}

	for idx2, pivotItem := range project.PivotItems {
		data["Item"] = project.PivotItems[idx2]
		data["Module"] = pivotItem.Module
		if IsNewItem(pivotItem, project.PrevProject) {
			migrationLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_create_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationCreateTable, pivotItem.Table), data, true)
		} else if IsRenameItem(pivotItem, project.PrevProject) {
			data["PrevItem"] = project.PrevProject.MapItems[pivotItem.Id]
			migrationRenameLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_rename_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationRenameTable, pivotItem.Table), data, true)
		}
	}

	if nil != project.PrevProject {
		// 1) 删除被移除的 Item 表
		for idx1, mod := range project.PrevProject.Modules {
			data["Module"] = project.PrevProject.Modules[idx1]
			for idx2, item := range mod.Items {
				data["Item"] = project.PrevProject.Modules[idx1].Items[idx2]
				// Migration
				if IsDropItem(item, project) {
					migrationDropTableLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_drop_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationDropTable, item.Table), data, true)
				}
			}
		}

		for idx2, pivotItem := range project.PrevProject.PivotItems {
			data["Item"] = project.PrevProject.PivotItems[idx2]
			data["Module"] = pivotItem.Module
			if IsDropItem(pivotItem, project) {
				migrationDropTableLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_drop_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationDropTable, pivotItem.Table), data, true)
			}
		}

		// 2) 比较新旧 Item 的 attribute 并生成字段变更迁移
		prevProject := project.PrevProject

		// Modules -> normal items
		for _, mod := range project.Modules {
			for _, item := range mod.Items {
				prevItem, ok := prevProject.MapItems[item.Id]
				if !ok {
					continue // new item, handled by create migration earlier
				}

				for _, attr := range item.Attrs {
					if IsAddAttribute(attr, prevProject) {
						migrationAttrAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationAttrAdd, item.Table, attr.Snake()), map[string]interface{}{
							"Project":      project,
							"Module":       item.Module,
							"Item":         item,
							"AddAttribute": attr,
						}, true)
					} else if IsDropWithAddAttribute(attr, prevProject) {
						migrationAttrAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationAttrAdd, item.Table, attr.Snake()), map[string]interface{}{
							"Project":      project,
							"Module":       item.Module,
							"Item":         item,
							"AddAttribute": attr,
						}, true)
						oldAttr := prevProject.MapAttributes[attr.Id]
						migrationAttrRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationAttrRemove, item.Table, oldAttr.Snake()), map[string]interface{}{
							"Project":         project,
							"Module":          item.Module,
							"Item":            item,
							"RemoveAttribute": oldAttr,
						}, true)
					} else if IsUpdateAttribute(attr, prevProject) {
						migrationAttrUpdateLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_update_%s_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationAttrUpdate, item.Table, attr.Snake()), map[string]interface{}{
							"Project":      project,
							"Module":       item.Module,
							"Item":         item,
							"NewAttribute": attr,
							"OldAttribute": prevProject.MapAttributes[attr.Id],
						}, true)
					}
				}
				for _, oldAttr := range prevItem.Attrs {
					if IsRemoveAttribute(oldAttr, project) {
						migrationAttrRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationAttrRemove, item.Table, oldAttr.Snake()), map[string]interface{}{
							"Project":         project,
							"Module":          item.Module,
							"Item":            item,
							"RemoveAttribute": oldAttr,
						}, true)
					}
				}

				for _, index := range item.Indexes {
					if IsNewIndex(index, prevProject) {
						migrationIndexAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_index_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationIndexAdd, item.Table, index.IndexName()), map[string]interface{}{
							"Project":  project,
							"Module":   item.Module,
							"Item":     item,
							"NewIndex": index,
						}, true)
					} else if IsDropAndRecreateIndex(index, prevProject) {
						migrationIndexAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_index_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationIndexAdd, item.Table, index.IndexName()), map[string]interface{}{
							"Project":  project,
							"Module":   item.Module,
							"Item":     item,
							"NewIndex": index,
						}, true)
						migrationIndexRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_index_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationIndexRemove, item.Table, index.IndexName()), map[string]interface{}{
							"Project": project,
							"Module":  item.Module,
							"Item":    item,
							"Index":   index,
						}, true)
					} else if IsRenameIndex(index, prevProject) {
						migrationIndexRenameLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_rename_%s_index_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationIndexRename, item.Table, index.IndexName()), map[string]interface{}{
							"Project":   project,
							"Module":    item.Module,
							"Item":      item,
							"NewIndex":  index,
							"PrevIndex": prevProject.MapIndexes[index.Id],
						}, true)
					}
				}
				for _, oldIndex := range prevItem.Indexes {
					if IsDropIndex(oldIndex, project) {
						migrationIndexRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_index_%s.php", item.MigrationOutputPath(), item.MigrationTime(), project.BuildVersion, MigrationIndexRemove, item.Table, oldIndex.IndexName()), map[string]interface{}{
							"Project": project,
							"Module":  item.Module,
							"Item":    item,
							"Index":   oldIndex,
						}, true)
					}
				}
			}
		}

		// Pivot items
		for _, pivotItem := range project.PivotItems {
			prevItem, ok := prevProject.MapItems[pivotItem.Id]
			if !ok {
				continue // new pivot item
			}

			for _, attr := range pivotItem.Attrs {
				if IsAddAttribute(attr, prevProject) {
					migrationAttrAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationAttrAdd, pivotItem.Table, attr.Snake()), map[string]interface{}{
						"Project":      project,
						"Module":       pivotItem.Module,
						"Item":         pivotItem,
						"AddAttribute": attr,
					}, true)
				} else if IsDropWithAddAttribute(attr, prevProject) {
					oldAttr := prevProject.MapAttributes[attr.Id]
					migrationAttrAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationAttrAdd, pivotItem.Table, attr.Snake()), map[string]interface{}{
						"Project":      project,
						"Module":       pivotItem.Module,
						"Item":         pivotItem,
						"AddAttribute": attr,
					}, true)
					migrationAttrRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationAttrRemove, pivotItem.Table, oldAttr.Snake()), map[string]interface{}{
						"Project":         project,
						"Module":          pivotItem.Module,
						"Item":            pivotItem,
						"RemoveAttribute": oldAttr,
					}, true)
				} else if IsUpdateAttribute(attr, prevProject) {
					oldAttr := prevProject.MapAttributes[attr.Id]
					migrationAttrUpdateLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_update_%s_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationAttrUpdate, pivotItem.Table, attr.Snake()), map[string]interface{}{
						"Project":      project,
						"Module":       pivotItem.Module,
						"Item":         pivotItem,
						"NewAttribute": attr,
						"OldAttribute": oldAttr,
					}, true)
				}
			}
			for _, oldAttr := range prevItem.Attrs {
				if IsRemoveAttribute(oldAttr, project) {
					migrationAttrRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_%s.php", prevItem.MigrationOutputPath(), prevItem.MigrationTime(), project.BuildVersion, MigrationAttrRemove, prevItem.Table, oldAttr.Snake()), map[string]interface{}{
						"Project":         project,
						"Module":          prevItem.Module,
						"Item":            prevItem,
						"RemoveAttribute": oldAttr,
					}, true)
				}
			}

			for _, index := range pivotItem.Indexes {
				if IsNewIndex(index, prevProject) {
					migrationIndexAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_index_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationIndexAdd, pivotItem.Table, index.IndexName()), map[string]interface{}{
						"Project":  project,
						"Module":   pivotItem.Module,
						"Item":     pivotItem,
						"NewIndex": index,
					}, true)
				} else if IsDropAndRecreateIndex(index, prevProject) {
					migrationIndexAddLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_add_%s_index_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationIndexAdd, pivotItem.Table, index.IndexName()), map[string]interface{}{
						"Project":  project,
						"Module":   pivotItem.Module,
						"Item":     pivotItem,
						"NewIndex": index,
					}, true)
					migrationIndexRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_index_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationIndexRemove, pivotItem.Table, index.IndexName()), map[string]interface{}{
						"Project": project,
						"Module":  pivotItem.Module,
						"Item":    pivotItem,
						"Index":   index,
					}, true)
				} else if IsRenameIndex(index, prevProject) {
					migrationIndexRenameLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_rename_%s_index_%s.php", pivotItem.MigrationOutputPath(), pivotItem.MigrationTime(), project.BuildVersion, MigrationIndexRename, pivotItem.Table, index.IndexName()), map[string]interface{}{
						"Project":   project,
						"Module":    pivotItem.Module,
						"Item":      pivotItem,
						"NewIndex":  index,
						"PrevIndex": prevProject.MapIndexes[index.Id],
					}, true)
				}
			}
			for _, oldIndex := range prevItem.Indexes {
				if IsDropIndex(oldIndex, project) {
					migrationIndexRemoveLoader.Render(fmt.Sprintf("%s/%s_v%s_%d0_remove_%s_index_%s.php", prevItem.MigrationOutputPath(), prevItem.MigrationTime(), project.BuildVersion, MigrationIndexRemove, prevItem.Table, oldIndex.IndexName()), map[string]interface{}{
						"Project": project,
						"Module":  prevItem.Module,
						"Item":    prevItem,
						"Index":   oldIndex,
					}, true)
				}
			}
		}
	}
}

func IsNewItem(item *meta.Item, prevProject *meta.Project) bool {
	if nil == prevProject {
		return true
	}
	if _, ok := prevProject.MapItems[item.Id]; !ok {
		return true
	}

	return false
}

func IsDropItem(item *meta.Item, currentProject *meta.Project) bool {
	if _, ok := currentProject.MapItems[item.Id]; !ok {
		return true
	}

	return false
}

func IsRenameItem(item *meta.Item, prevProject *meta.Project) bool {
	if nil == prevProject {
		return false
	}
	if _, ok := prevProject.MapItems[item.Id]; !ok {
		return false
	}

	return prevProject.MapItems[item.Id].Table != item.Table
}

func IsAddAttribute(attr *meta.Attribute, prevProject *meta.Project) bool {
	if nil == prevProject {
		return true
	}
	if _, ok := prevProject.MapAttributes[attr.Id]; !ok {
		return true
	}

	return false
}

func IsRemoveAttribute(attr *meta.Attribute, currentProject *meta.Project) bool {
	if _, ok := currentProject.MapAttributes[attr.Id]; !ok {
		return true
	}

	return false
}

func IsDropWithAddAttribute(attr *meta.Attribute, prevProject *meta.Project) bool {
	if nil == prevProject {
		return false
	}

	if _, ok := prevProject.MapAttributes[attr.Id]; !ok {
		return false
	}

	if attr.FieldType != prevProject.MapAttributes[attr.Id].FieldType {
		return true
	}

	return false
}

func IsUpdateAttribute(attr *meta.Attribute, prevProject *meta.Project) bool {
	if nil == prevProject {
		return false
	}
	if _, ok := prevProject.MapAttributes[attr.Id]; !ok {
		return false
	}

	prevAttr := prevProject.MapAttributes[attr.Id]

	if attr.StrFieldLength() != prevAttr.StrFieldLength() && attr.StrFieldLength() != -1 && prevAttr.StrFieldLength() != -1 {
		return true
	}

	return attr.Nullable != prevAttr.Nullable || attr.DefaultValue != prevAttr.DefaultValue || attr.DefaultValueType != prevAttr.DefaultValueType
}

func IsNewIndex(index *meta.Index, prevProject *meta.Project) bool {
	if nil == prevProject {
		return true
	}
	if _, ok := prevProject.MapIndexes[index.Id]; !ok {
		return true
	}

	return false
}

func IsDropIndex(index *meta.Index, currentProject *meta.Project) bool {
	if _, ok := currentProject.MapIndexes[index.Id]; !ok {
		return true
	}

	return false
}

func IsRenameIndex(index *meta.Index, prevProject *meta.Project) bool {
	if nil == prevProject {
		return false
	}
	if _, ok := prevProject.MapIndexes[index.Id]; !ok {
		return false
	}

	prevIndex := prevProject.MapIndexes[index.Id]

	if prevIndex.ColumnNamesString() != index.ColumnNamesString() {
		return false
	}
	if prevIndex.Type != index.Type {
		return false
	}

	return prevProject.MapIndexes[index.Id].IndexName() != index.IndexName()
}

func IsDropAndRecreateIndex(index *meta.Index, prevProject *meta.Project) bool {
	if nil == prevProject {
		return false
	}

	if _, ok := prevProject.MapIndexes[index.Id]; !ok {
		return false
	}

	prevIndex := prevProject.MapIndexes[index.Id]

	if prevIndex.ColumnNamesString() != index.ColumnNamesString() {
		return true
	}

	if prevIndex.Type != index.Type {
		return true
	}

	return false
}
