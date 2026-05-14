package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/conf"
	"github.com/light2000/laravel-modeler-generator/laravel"
	"github.com/light2000/laravel-modeler-generator/logger"
	"github.com/light2000/laravel-modeler-generator/meta"
	"github.com/light2000/laravel-modeler-generator/proto"
	"github.com/light2000/laravel-modeler-generator/templatex"
)

type composerFile struct {
	Autoload struct {
		Psr4 map[string]string `json:"psr-4"`
	} `json:"autoload"`
	Extra struct {
		MergePlugin struct {
			Include []string `json:"include"`
		} `json:"merge-plugin"`
	} `json:"extra"`
}

func loadComposerPsr4Map(laravelDir string) (map[string]string, error) {
	result := make(map[string]string)
	laravelAbs, err := filepath.Abs(filepath.Clean(laravelDir))
	if err != nil {
		return nil, fmt.Errorf("laravel project dir: %w", err)
	}
	rootComposer := filepath.Join(laravelDir, "composer.json")
	root, err := readComposer(rootComposer)
	if err != nil {
		return nil, err
	}
	mergePsr4(result, root.Autoload.Psr4, laravelDir, laravelAbs)
	for _, pattern := range root.Extra.MergePlugin.Include {
		absPattern := filepath.Join(laravelDir, filepath.FromSlash(pattern))
		includes, globErr := filepath.Glob(absPattern)
		if globErr != nil {
			return nil, fmt.Errorf("invalid composer include pattern %s: %w", pattern, globErr)
		}
		for _, includeFile := range includes {
			child, childErr := readComposer(includeFile)
			if childErr != nil {
				return nil, childErr
			}
			mergePsr4(result, child.Autoload.Psr4, filepath.Dir(includeFile), laravelAbs)
		}
	}
	return result, nil
}

func readComposer(file string) (*composerFile, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read composer file failed: %s: %w", file, err)
	}
	var out composerFile
	if err := json.Unmarshal(content, &out); err != nil {
		return nil, fmt.Errorf("parse composer file failed: %s: %w", file, err)
	}
	return &out, nil
}

// mergePsr4 merges autoload psr-4 entries into dst. Paths in composer.json are relative to
// that file's directory (composerDir). Values are stored relative to laravelRootAbs when the
// resolved directory lies inside the Laravel project; otherwise an absolute path is stored
// so template output can still target the correct tree (see templatex.resolveOutputFile).
func mergePsr4(dst map[string]string, src map[string]string, composerDir, laravelRootAbs string) {
	composerDir = filepath.Clean(composerDir)
	laravelRootAbs = filepath.Clean(laravelRootAbs)
	for namespace, path := range src {
		namespace = strings.TrimSpace(namespace)
		path = strings.TrimSpace(path)
		if namespace == "" || path == "" {
			continue
		}
		p := filepath.FromSlash(path)
		var resolved string
		if filepath.IsAbs(p) {
			resolved = filepath.Clean(p)
		} else {
			resolved = filepath.Clean(filepath.Join(composerDir, p))
		}
		resolvedAbs, absErr := filepath.Abs(resolved)
		if absErr != nil {
			dst[namespace] = filepath.ToSlash(resolved)
			continue
		}
		rel, relErr := filepath.Rel(laravelRootAbs, resolvedAbs)
		if relErr != nil {
			dst[namespace] = filepath.ToSlash(resolvedAbs)
			continue
		}
		rel = filepath.ToSlash(filepath.Clean(rel))
		if rel == ".." || strings.HasPrefix(rel, "../") {
			dst[namespace] = filepath.ToSlash(resolvedAbs)
			continue
		}
		dst[namespace] = rel
	}
}

func applyResolvedOutputPaths(project *meta.Project, psr4Map map[string]string) {
	project.Psr4Map = psr4Map
	for _, module := range project.Modules {
		if module == nil || module.Namespace == nil {
			continue
		}
		module.Namespace.ModelOutputPath = project.ResolveOutputPathByNamespace(module.Namespace.ModelNamespace)
		module.Namespace.FactoryOutputPath = project.ResolveOutputPathByNamespace(module.Namespace.FactoryNamespace)
		module.Namespace.SeederOutputPath = project.ResolveOutputPathByNamespace(module.Namespace.SeederNamespace)
		module.Namespace.EnumOutputPath = project.ResolveOutputPathByNamespace(module.Namespace.EnumNamespace)
	}
	project.PivotLayout.ModelOutputPath = project.ResolveOutputPathByNamespace(project.PivotLayout.ModelNamespace)
	project.PivotLayout.FactoryOutputPath = project.ResolveOutputPathByNamespace(project.PivotLayout.FactoryNamespace)
	project.GlobalEnumLayout.OutputPath = project.ResolveOutputPathByNamespace(project.GlobalEnumLayout.Namespace)
}

var (
	configPath string
)

func main() {
	defer handlePanic()
	flag.StringVar(&configPath, "config", "", "配置文件路径（必填）")
	flag.Parse()
	if configPath == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须指定 -config 配置文件路径\n")
		flag.Usage()
		os.Exit(1)
	}

	if err := conf.LoadConfig(configPath); err != nil {
		log.Fatalf("读取JSON配置失败: %v", err)
	}
	if err := run(); err != nil {
		handleError(err)
	}
}

func run() error {
	conf.InitFeatureStatus()

	laravelDir := conf.Config.ProjectDir
	versionsDir := conf.Config.DataPath
	templatesDir := conf.Config.TemplatesPath
	if err := initLogger(); err != nil {
		return fmt.Errorf("init logger failed: %w", err)
	}

	logger.Infof("generator started: command=generate config=%s", configPath)

	if err := runGenerate(laravelDir, templatesDir, versionsDir); err != nil {
		return fmt.Errorf("run generate failed: %w", err)
	}

	logger.Infof("generate finished")
	return nil
}

func runGenerate(laravelDir string, templatesDir string, versionsDir string) error {
	var pbProject *proto.Project
	var project *meta.Project
	var err error

	engine := templatex.NewEngine(laravelDir, templatesDir)

	pbProject = &proto.Project{}

	if err = common.LoadProtoFromJSONFile(filepath.Join(versionsDir, "latest.json"), pbProject); err != nil {
		return err
	}
	psr4Map, err := loadComposerPsr4Map(laravelDir)
	if err != nil {
		return err
	}
	// 读取 ${schema}/versions/ 目录下的所有文件名，放入 versions 数组
	snapshotsDir := filepath.Join(versionsDir, "snapshots")
	versions := []string{}
	if entries, err := os.ReadDir(snapshotsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				versions = append(versions, entry.Name())
			}
		}
	} else if !os.IsNotExist(err) {
		// 如果目录不存在，则忽略，否则返回错误
		return fmt.Errorf("failed to read versions directory: %w", err)
	}
	sort.Strings(versions)
	versionProjects := make([]*meta.Project, 0)
	for _, version := range versions {
		versionPbProject := &proto.Project{}
		if err = common.LoadProtoFromJSONFile(filepath.Join(snapshotsDir, version), versionPbProject); err != nil {
			return err
		}
		versionProject := meta.FromProtoProject(versionPbProject)
		applyResolvedOutputPaths(versionProject, psr4Map)
		versionProjects = append(versionProjects, versionProject)
	}

	for idx, versionProject := range versionProjects {
		if idx > 0 {
			versionProject.PrevProject = versionProjects[idx-1]
		} else {
			versionProject.PrevProject = nil
		}
	}

	// 设置pbProject.BuildVersion
	// 获取当前所有版本号，去除前置0，找到最大值
	maxVersion := 0
	for _, v := range versions {
		// 去除文件名的扩展名（如 .json）
		versionName := v
		if ext := filepath.Ext(versionName); ext != "" {
			versionName = versionName[:len(versionName)-len(ext)]
		}
		// 去除前置0
		intVal := 0
		fmt.Sscanf(versionName, "%d", &intVal)
		if intVal > maxVersion {
			maxVersion = intVal
		}
	}

	// 新的版本号+1并补零至四位
	newVersion := maxVersion + 1
	newVersionStr := fmt.Sprintf("%04d", newVersion)
	pbProject.BuildVersion = newVersionStr

	project = meta.FromProtoProject(pbProject)
	applyResolvedOutputPaths(project, psr4Map)
	if len(versionProjects) > 0 {
		project.PrevProject = versionProjects[len(versionProjects)-1]
	} else {
		project.PrevProject = nil
	}

	laravel.ProjectBuild(project, engine)
	laravel.DictBuild(project, engine)
	laravel.ItemBuild(project, engine)
	laravel.ItemMigrationBuild(project, engine)
	logger.Infof("generated files finished")
	// 保存新的versions/xxxx.json
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}
	newVersionFile := filepath.Join(snapshotsDir, newVersionStr+".json")
	if err := common.SaveProtoToJSONFile(newVersionFile, pbProject); err != nil {
		return fmt.Errorf("failed to save new version file: %w", err)
	}
	logger.Infof("new snapshot file saved: %s", newVersionFile)
	return nil
}

func initLogger() error {
	logPath := filepath.Join(conf.Config.LogPath, "generator.log")
	if err := logger.Init(logPath, logger.InfoLevel); err != nil {
		return fmt.Errorf("init logger failed: %w", err)
	}
	return nil
}

func handleError(err error) {
	if err == nil {
		return
	}

	logger.Errorf("%v", err)
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func handlePanic() {
	if r := recover(); r != nil {
		logger.Errorf("panic: %v\n%s", r, debug.Stack())

		if common.IsDebug() {
			panic(r)
		}

		fmt.Fprintln(os.Stderr, "Internal error:", r)
		os.Exit(2)
	}
}
