package buildscript

import (
	"bytes"
	"os"
	text "text/template"
)

// git update-index --chmod +x script.sh

// Param 基础配置
type Param struct {
	Name  string
	Org   string
	Main  string
	HasUI bool
}

// 构建脚本文件名称
const (
	scriptName     = "build.sh"
	dockerFileName = "Dockerfile"
)

func Generate(name, org, mainPath string, hasUI bool) (err error) {

	p := Param{
		Name:  name,
		Org:   org,
		Main:  mainPath,
		HasUI: hasUI,
	}
	scriptContent, err := rendText(p, script)
	if err != nil {
		return
	}
	err = os.WriteFile(scriptName, []byte(scriptContent), 0644)

	dockerContent, err := rendText(p, dockerFile)

	if err != nil {
		return
	}

	err = os.WriteFile(dockerFileName, []byte(dockerContent), 0644)
	return
}

func rendText(data interface{}, temp string) (content string, err error) {
	t, err := text.New("_").Parse(temp)
	if err != nil {
		return
	}
	w := new(bytes.Buffer)
	err = t.Execute(w, data)
	if err != nil {
		return
	}
	content = w.String()
	return
}

func GenerateBaseDockerfile() (err error) {

	dockerContent, err := rendText(nil, baseDockerFile)

	if err != nil {
		return
	}

	err = os.WriteFile(dockerFileName, []byte(dockerContent), 0644)
	return
}
