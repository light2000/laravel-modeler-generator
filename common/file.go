package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/light2000/laravel-modeler-generator/logger"
	protopkg "github.com/light2000/laravel-modeler-generator/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// PathExists 判断所给路径文件/文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false
	}
	logger.Fatalf("check exists failed: %s", err)

	return false
}

// IsDir 判断目录是否存在
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func LoadProtoFromJSONFile(path string, msg proto.Message) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read schema file %q: %w", path, err)
	}
	if err := UnmarshalProtoJSON(data, msg); err != nil {
		return fmt.Errorf("parse schema file %q: %w", path, err)
	}
	return nil
}

func UnmarshalProtoJSON(data []byte, msg proto.Message) error {
	opts := protojson.UnmarshalOptions{
		DiscardUnknown: false, // true: 忽略未知字段；false: 遇到未知字段直接报错
		AllowPartial:   false, // true: 允许缺少 required(旧proto2) / 不完整消息；一般用不到
	}
	if err := opts.Unmarshal(data, msg); err != nil {
		return err
	}
	return nil
}

func FlushFile(data interface{}, path string, template *template.Template) {
	var err error

	if PathExists(path) {
		return
	}

	if !IsDir(filepath.Dir(path)) {
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			logger.Errorf("create code dir failed: %s", err)
			panic(err)
		}
	}

	fi, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		logger.Errorf("open target code file failed: %s", err)
		panic(err)
	}
	defer func() {
		err = fi.Close()
		if err != nil {
			logger.Errorf("close target code failed: %s", err)
			panic(err)
		}
	}()

	err = template.Execute(fi, data)
	if err != nil {
		logger.Errorf("flush template content failed: %s", err)
		panic(err)
	}

	logger.Infof("file %s created", path)
}

func SaveProtoToJSONFile(path string, proj *protopkg.Project) error {
	jsonBytes, err := protojson.Marshal(proj)
	if err != nil {
		return fmt.Errorf("marshal proto to json failed: %w", err)
	}

	prettyJsonBytes, err := PrettyJSON(jsonBytes)
	if err != nil {
		return fmt.Errorf("pretty json failed: %w", err)
	}

	if err := os.WriteFile(path, prettyJsonBytes, 0644); err != nil {
		return fmt.Errorf("save schema file failed: %w", err)
	}
	return nil
}

func PrettyJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "    "); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ClearDir(path string) {
	logger.Infof("%s clear start", path)
	var err error
	_, err = os.Stat(path)
	if err != nil && !os.IsExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			logger.Errorf("create dir failed: %s", err)
			panic(err)
		}
	}

	files, err := os.ReadDir(path)
	if err != nil {
		logger.Errorf("read dir failed: %s", err)
		panic(err)
	}
	for _, file := range files {
		logger.Debugf("file %s found", file.Name())
		if file.IsDir() {
			logger.Debugf("dir %s found", fmt.Sprintf("%s/%s", path, file.Name()))
			ClearDir(fmt.Sprintf("%s/%s", path, file.Name()))
		} else {
			if !strings.HasPrefix(file.Name(), ".") {
				err = os.Remove(filepath.Join(path, file.Name()))
				if err != nil {
					logger.Errorf("remove file failed: %s", filepath.Join(path, file.Name()))
					panic(err)
				}
				logger.Debugf("file %s deleted", file.Name())
			}
		}

	}
	logger.Infof("%s cleaned", path)
}
