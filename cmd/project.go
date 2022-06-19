package cmd

import "os/exec"

type Project struct {
	ProjectName string
	ProjectDir  string
	Wd          string
}

func (p Project) Create() (err error) {

	cmd := exec.Command("go", "mod", "init", p.ProjectName)
	err = cmd.Run()
	return
}
