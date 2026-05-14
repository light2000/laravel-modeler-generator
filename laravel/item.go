package laravel

import (
	"fmt"

	"github.com/light2000/laravel-modeler-generator/meta"
	"github.com/light2000/laravel-modeler-generator/templatex"
)

func ItemBuild(project *meta.Project, engine *templatex.Engine) {
	factoryLoader := engine.NewLoader("laravel/item/Factory.tpl")
	factoryGeneratedLoader := engine.NewLoader("laravel/item/FactoryGenerated.tpl")
	seederBaseLoader := engine.NewLoader("laravel/item/SeederBase.tpl")
	seederRelationLoader := engine.NewLoader("laravel/item/SeederRelation.tpl")
	modelLoader := engine.NewLoader("laravel/item/Model.tpl")
	modelTraitLoader := engine.NewLoader("laravel/item/ModelTrait.tpl")

	data := map[string]interface{}{
		"Project": project,
	}

	for idx1, mod := range project.Modules {
		data["Module"] = project.Modules[idx1]
		for idx2, item := range mod.Items {
			data["Item"] = project.Modules[idx1].Items[idx2]
			//Model
			modelLoader.Render(fmt.Sprintf("%s/%s.php", item.ModelOutputPath(), item.Class()), data, false)
			modelTraitLoader.Render(fmt.Sprintf("%s/_Generated/_%sTrait.php", item.ModelOutputPath(), item.Class()), data, true)

			if item.HasFactory {
				// Factory
				factoryLoader.Render(fmt.Sprintf("%s/%sFactory.php", item.FactoryOutputPath(), item.Class()), data, true)
				factoryGeneratedLoader.Render(fmt.Sprintf("%s/_Generated/%sFactory.php", item.FactoryOutputPath(), item.Class()), data, true)
				if item.HasSeeder {
					//SeederBase
					seederBaseLoader.Render(fmt.Sprintf("%s/_Generated/%sBaseSeeder.php", item.SeederOutputPath(), item.Class()), data, true)
					if item.HasRelationSeeder() {
						//SeederRelation
						seederRelationLoader.Render(fmt.Sprintf("%s/_Generated/%sRelationSeeder.php", item.SeederOutputPath(), item.Class()), data, true)
					}
				}
			}
		}
	}

	for idx2, pivotItem := range project.PivotItems {
		data["Item"] = project.PivotItems[idx2]
		//Model
		modelLoader.Render(fmt.Sprintf("%s/%s.php", pivotItem.ModelOutputPath(), pivotItem.Class()), data, false)
		modelTraitLoader.Render(fmt.Sprintf("%s/_Generated/_%sTrait.php", pivotItem.ModelOutputPath(), pivotItem.Class()), data, true)
		// Factory
		factoryLoader.Render(fmt.Sprintf("%s/%sFactory.php", pivotItem.FactoryOutputPath(), pivotItem.Class()), data, true)
		factoryGeneratedLoader.Render(fmt.Sprintf("%s/_Generated/%sFactory.php", pivotItem.FactoryOutputPath(), pivotItem.Class()), data, true)
	}
}
