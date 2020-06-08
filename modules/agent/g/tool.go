// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package g

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/toolkits/file"
)

func GetCurrPluginVersion() string {
	if !Config().Plugin.Enabled {
		return "plugin not enabled"
	}

	pluginDir := Config().Plugin.Dir
	if !file.IsExist(pluginDir) {
		return "plugin dir not existent"
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = pluginDir

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error:%s", err.Error())
	}

	return strings.TrimSpace(out.String())
}

//common funcs
func IntSliceEqualBCE(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func StrSliceEqualBCE(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// 删除输出的\x00和多余的空格
func trimOutput(buffer bytes.Buffer) string {
	return strings.TrimSpace(string(bytes.TrimRight(buffer.Bytes(), "\x00")))
}

// 运行Shell命令，设定超时时间（秒）
func ShellCmdTimeout(timeout int, cmd string, args ...string) (stdout, stderr string, e error) {
	if len(cmd) == 0 {
		e = errors.New("Cannot run a empty command")
		return "", "", e
	}
	var out, err bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &out
	command.Stderr = &err
	command.Start()
	// 启动routine等待结束
	done := make(chan error)
	go func() { done <- command.Wait() }()

	// 设定超时时间，并select它
	after := time.After(time.Duration(timeout) * time.Second)
	select {
	case <-after:
		command.Process.Signal(syscall.SIGINT)
		time.Sleep(time.Second)
		command.Process.Kill()
		log.Errorln("Exe shell timeout:", cmd, strings.Join(args, " "), timeout)
	case <-done:
	}
	stdout = trimOutput(out)
	stderr = trimOutput(err)
	return stdout, stderr, nil
}

func IntArrayToStringArr(intArr []int) []string {
	var strArr []string
	for _, intpara := range intArr {
		tmp := strconv.Itoa(intpara)
		strArr = append(strArr, tmp)
	}
	return strArr
}
func CheckPathAndMkdir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	// check again
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func CheckFileExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func ExeShellCommand(cmdStr string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	opBytes, err := cmd.Output()
	return string(opBytes), err
}
