package cmd

import (
	"bull/util"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project in current directory",
	Long:  `Create a new project in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		var projectName string
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Project name is required")
			os.Exit(1)
		} else {
			projectName = args[0]
		}
		err := initProject(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Create Project %s Success\n", projectName)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func initProject(projectName string) (err error) {
	var wd, projectDir string
	if wd, err = os.Getwd(); err != nil {
		return
	}
	projectDir = filepath.Join(wd, projectName)
	exist, err := util.PathExists(projectDir)
	if err != nil {
		return
	}
	if exist {
		return errors.New("project directory is already exist")
	}

	if err = os.Mkdir(projectDir, 0700); err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(projectDir)
		}
	}()
	p := Project{
		ProjectName: formatProjectName(projectName),
		ProjectDir:  projectDir,
		Wd:          wd,
	}

	// read params from stdin
	p.ReadParam()

	return p.Create()
}

func formatProjectName(projectName string) string {
	return strings.ReplaceAll(projectName, "-", "_")
}
