package cmd

import (
	"bull/template"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	tpl "text/template"
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

	// package cmd
	if err = p.createCmd(); err != nil {
		return
	}

	// package internal
	if err = p.createInternal(); err != nil {
		return
	}

	// package model
	if err = p.createModel(); err != nil {
		return
	}

	// init go mod
	err = p.initMod()
	return
}

func (p *Project) createApi() (err error) {
	fmt.Println("create api dir")
	return p.createByMap("api", template.ApiMap)
}

func (p *Project) createCmd() (err error) {
	fmt.Println("create cmd dir")
	return p.createByMap("cmd", template.CmdMap)
}

func (p *Project) createInternal() (err error) {
	fmt.Println("create internal dir")
	return p.createByMap("internal", template.InternalMap)
}

func (p *Project) createModel() (err error) {
	fmt.Println("create model dir")
	return p.createByMap("model", template.ModelMap)
}

func (p *Project) createByMap(dir string, m map[string]*tpl.Template) (err error) {
	dir = filepath.Join(p.ProjectDir, dir)
	fmt.Println("create dir ", dir)
	if err = os.Mkdir(dir, 0755); err != nil {
		return
	}

	for k, v := range m {
		var f *os.File
		filePath := filepath.Join(dir, fmt.Sprintf("%s.go", k))
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
	fmt.Println("init go mod")
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
