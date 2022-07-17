package cmd

import (
	"bull/template"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Project struct {
	ProjectName string
	ProjectDir  string
	Wd          string
}

func (p *Project) Create() (err error) {
	// package api
	if err = p.createApi(); err != nil {
		return
	}

	err = p.initMod()
	return
}

func (p *Project) createApi() (err error) {
	for k, v := range template.ApiMap {
		var f *os.File
		filePath := filepath.Join(p.ProjectDir, fmt.Sprintf("%s.go", k))
		fmt.Println("create file ", filePath)
		if f, err = os.Create(filePath); err != nil {
			return
		}
		if err = v.Execute(f, p); err != nil {
			_ = f.Close()
			return
		}
		_ = f.Close()
	}
	return
}

func (p *Project) initMod() (err error) {
	initCmd := exec.Command("go", "mod", "init", p.ProjectName)
	initCmd.Dir = p.ProjectDir
	if err = initCmd.Run(); err != nil {
		return
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = p.ProjectDir
	err = tidyCmd.Run()
	return
}
