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

	// package pkg
	if err = p.createPkg(); err != nil {
		return
	}

	// package third_party
	if err = p.createThirdPart(); err != nil {
		return
	}

	// common file
	if err = p.createCommon(); err != nil {
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
	return p.createByMap("models", template.ModelMap)
}

func (p *Project) createPkg() (err error) {
	fmt.Println("create pkg dir")
	return p.createByMap("pkg", template.PkgMap)
}

func (p *Project) createThirdPart() (err error) {
	fmt.Println("create config dir")
	return p.createByMap("third_party", template.ThirdPartMap)
}

func (p *Project) createCommon() (err error) {
	fmt.Println("create common file")
	return p.createByMap("", template.CommonMap)
}

func (p *Project) createByMap(dir string, m map[string]*tpl.Template) (err error) {
	// create dir
	dir = filepath.Join(p.ProjectDir, dir)
	if dir != p.ProjectDir {
		fmt.Println("create dir ", dir)
		err = os.Mkdir(dir, 0755)
	}
	if err != nil {
		return
	}

	for k, v := range m {
		var f *os.File
		filePath := filepath.Join(dir, k)
		fileDir := filepath.Dir(filePath)
		if fileDir != p.ProjectDir {
			// create file dir
			err = os.MkdirAll(fileDir, 0755)
		}
		if err != nil {
			return
		}
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
