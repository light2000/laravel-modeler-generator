package laravel

import (
	"path/filepath"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/meta"
	"github.com/light2000/laravel-modeler-generator/templatex"
)

func ProjectBuild(project *meta.Project, engine *templatex.Engine) {

	if common.IsReplaceMode() && common.IsDebug() {
		common.ClearDir(filepath.Join(engine.OutputBaseDir, "database/seeders"))
		common.ClearDir(filepath.Join(engine.OutputBaseDir, "database/factories"))
		common.ClearDir(filepath.Join(engine.OutputBaseDir, "app/Enums"))
		common.ClearDir(filepath.Join(engine.OutputBaseDir, "app/Models"))
		common.ClearDir(filepath.Join(engine.OutputBaseDir, "database/migrations"))
		common.ClearDir(filepath.Join(engine.OutputBaseDir, "config/_generated"))
	}
	modelerConfigGeneratedTemplates := engine.NewLoader("laravel/project/app.config.generated.modeler.tpl")
	modelerSeederTemplates := engine.NewLoader("laravel/project/app.database.seeders.ModelerSeeder.tpl")
	databaseSeederTemplates := engine.NewLoader("laravel/project/app.database.seeders.DatabaseSeeder.tpl")
	data := map[string]interface{}{
		"Project": project,
	}

	modelerConfigGeneratedTemplates.Render("config/_generated/modeler.php", data, true)
	modelerSeederTemplates.Render("database/seeders/ModelerSeeder.php", data, false)
	databaseSeederTemplates.Render("database/seeders/DatabaseSeeder.php", data, false)
}
