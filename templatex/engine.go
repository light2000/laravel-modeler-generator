package templatex

import (
	"path/filepath"
)

func NewEngine(baseDir string, templatesDir string) *Engine {
	return &Engine{
		OutputBaseDir: baseDir,
		TemplatesDir:  templatesDir,
	}
}

type Engine struct {
	OutputBaseDir string
	TemplatesDir  string
}

func (e *Engine) NewLoader(tpl string) *Loader {
	return &Loader{
		tpl:           filepath.Join(e.TemplatesDir, tpl),
		outputBaseDir: e.OutputBaseDir,
		template:      nil,
	}
}
