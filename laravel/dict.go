package laravel

import (
	"fmt"

	"github.com/light2000/laravel-modeler-generator/meta"
	"github.com/light2000/laravel-modeler-generator/templatex"
)

func DictBuild(project *meta.Project, engine *templatex.Engine) {
	enumTemplates := engine.NewLoader("laravel/dict/Enum.tpl")
	enumExtTemplates := engine.NewLoader("laravel/dict/EnumExt.tpl")

	data := map[string]interface{}{
		"Project": project,
	}

	for _, dict := range project.Dictionaries {
		data["Dict"] = dict
		if dict.IsGlobal() {
			data["Module"] = nil
			enumTemplates.Render(fmt.Sprintf("%s/%s.php", project.GlobalEnumLayout.OutputPath, dict.Class()), data, true)
			enumExtTemplates.Render(fmt.Sprintf("%s/Concerns/%sExt.php", project.GlobalEnumLayout.OutputPath, dict.Class()), data, false)
		} else {
			data["Module"] = dict.Module()
			enumTemplates.Render(fmt.Sprintf("%s/%s.php", dict.Module().Namespace.EnumOutputPath, dict.Class()), data, true)
			enumExtTemplates.Render(fmt.Sprintf("%s/Concerns/%sExt.php", dict.Module().Namespace.EnumOutputPath, dict.Class()), data, false)
		}
	}
}
