package templatex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/logger"
)

type Loader struct {
	tpl           string
	outputBaseDir string
	template      *template.Template
}

// resolveOutputFile builds the final write path. If relOrAbs is already an absolute path,
// outputBaseDir is not prepended (Psr4Map / resolved paths may be absolute when outside the project root).
func resolveOutputFile(outputBaseDir, relOrAbs string) string {
	s := strings.TrimSpace(relOrAbs)
	if s == "" {
		return filepath.Clean(outputBaseDir)
	}
	p := filepath.FromSlash(s)
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(outputBaseDir, p))
}

func (l *Loader) Render(path string, data interface{}, replace bool) {
	var err error
	outputPath := path
	path = resolveOutputFile(l.outputBaseDir, path)
	if nil == l.template {
		l.template, err = template.New(filepath.Base(l.tpl)).Option("missingkey=error").Funcs(funcMap).ParseFiles(l.tpl)
		if err != nil {
			panic(fmt.Errorf("parse template %q failed: %w", l.tpl, err))
		}
	}
	existing := common.PathExists(path)
	if existing {
		if !replace && !common.IsReplaceMode() {
			logger.Infof("%s already exists, skip", outputPath)
			return
		}
	}

	if !common.IsDir(filepath.Dir(path)) {
		err = os.MkdirAll(filepath.Dir(path), 0644)
		if err != nil {
			panic(fmt.Errorf("create code dir %q failed: %w", filepath.Dir(path), err))
		}
	}

	tmpPath := path + ".tmp"
	fp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("open temp code file %q failed: %w", tmpPath, err))
	}
	execErr := l.template.Execute(fp, data)
	closeErr := fp.Close()
	if execErr != nil {
		_ = os.Remove(tmpPath)
		panic(fmt.Errorf("flush template content %q failed: %w", l.tpl, execErr))
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		panic(fmt.Errorf("close temp code file %q failed: %w", tmpPath, closeErr))
	}
	if err = os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		panic(fmt.Errorf("rename %q -> %q failed: %w", tmpPath, path, err))
	}

	if existing {
		logger.Infof("%s replaced", outputPath)
	} else {
		logger.Infof("%s generated", outputPath)
	}
}
