package common

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/light2000/laravel-modeler-generator/logger"
)

func IsDebug() bool {
	return os.Getenv("GENERATOR_DEBUG") == "1"
}

func IsReplaceMode() bool {
	return os.Getenv("GENERATOR_REPLACE") == "1"
}

func RunCmd(WorkspaceDir, cmd string, args []string) error {
	// 使用exec.Command执行命令，并传入参数
	c := exec.Command(cmd, args...)
	c.Dir = WorkspaceDir
	output, err := c.CombinedOutput()
	if err != nil {
		// 把子进程的 stdout/stderr 放进错误信息，便于排查
		return fmt.Errorf("command %q in workspace %q failed: %w\noutput:\n%s", cmd, WorkspaceDir, err, string(output))
	}
	logger.Infof("command %q in workspace %q completed: %s", cmd, WorkspaceDir, string(output))

	return nil
}

func PhpEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	return s
}
