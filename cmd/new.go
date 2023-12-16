package cmd

import (
	"embed"
	_ "embed"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

//go:embed templates
var templates embed.FS

func New(eLogger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [github account] [project name]",
		Short: "Create a new project",
		Long:  `Create a new project with the Echo, Zap and Godotenv installed`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			githubAcc := args[0]
			projectName := args[1]

			if err := initPrj(githubAcc, projectName, eLogger); err != nil {
				eLogger.Printf("Failed to initialize project: %v", err)

				return
			}

			ecs, err := cmd.Flags().GetBool("ecs")
			if err != nil {
				eLogger.Printf("Failed to get ECS flag: %v", err)

				return
			}

			gh, err := cmd.Flags().GetBool("gh")
			if err != nil {
				eLogger.Printf("Failed to get GitHub flag: %v", err)

				return
			}

			var wg sync.WaitGroup

			if ecs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					generateECS(projectName, eLogger)
				}()
			}

			if gh {
				wg.Add(1)
				go func() {
					defer wg.Done()
					generateGH(projectName, eLogger)
				}()
			}

			wg.Wait()
		},
	}

	cmd.Flags().BoolP("ecs", "e", false, "Add templates for ECS CI deployment")
	cmd.Flags().BoolP("gh", "g", false, "Add the GitHub Actions folder with a certain template")

	return cmd
}

func generateGH(projectName string, logger *log.Logger) {
	ghTmpl, err := fs.ReadFile(templates, "templates/gh.txt")
	if err != nil {
		logger.Printf("Failed to read GitHub template: %v", err)

		return
	}
	ghTmplStr := string(ghTmpl)

	ghTmplStr = strings.ReplaceAll(ghTmplStr, "PRJ-NAME", projectName)

	devTemplate := strings.ReplaceAll(ghTmplStr, "ENV", "dev")
	devFilePath := filepath.Join(".github", "workflows", "ecs-container-deployment-dev.yml")
	err = os.MkdirAll(filepath.Dir(devFilePath), os.ModePerm)
	if err != nil {
		logger.Printf("Failed to create directory: %v", err)

		return
	}
	err = os.WriteFile(devFilePath, []byte(devTemplate), 0600)
	if err != nil {
		logger.Printf("Failed to write dev template: %v", err)

		return
	}

	prodTemplate := strings.ReplaceAll(ghTmplStr, "ENV", "prod")
	prodFilePath := filepath.Join(".github", "workflows", "ecs-container-deployment-prod.yml")
	err = os.WriteFile(prodFilePath, []byte(prodTemplate), 0600)
	if err != nil {
		logger.Printf("Failed to write prod template: %v", err)

		return
	}
}

func generateECS(projectName string, logger *log.Logger) {
	ecsTmpl, err := fs.ReadFile(templates, "templates/ecs.txt")
	if err != nil {
		logger.Printf("Failed to read ECS template: %v", err)

		return
	}
	ecsTmplStr := string(ecsTmpl)

	ecsTmplStr = strings.ReplaceAll(ecsTmplStr, "PRJ-NAME", projectName)

	devTemplate := strings.ReplaceAll(ecsTmplStr, "ENV", "dev")
	devFilePath := filepath.Join(".", "task_definition-dev.json")
	err = os.WriteFile(devFilePath, []byte(devTemplate), 0600)
	if err != nil {
		logger.Printf("Failed to write dev template: %v", err)

		return
	}

	prodTemplate := strings.ReplaceAll(ecsTmplStr, "ENV", "prod")
	prodFilePath := filepath.Join(".", "task_definition-prod.json")
	err = os.WriteFile(prodFilePath, []byte(prodTemplate), 0600)
	if err != nil {
		logger.Printf("Failed to write prod template: %v", err)

		return
	}
}

func initPrj(githubAcc string, projectName string, eLogger *log.Logger) error {
	err := os.Mkdir(projectName, os.ModePerm)
	if err != nil {
		eLogger.Printf("Failed to create directory: %v", err)

		return err
	}

	cleanGithubAcc := filepath.Clean(githubAcc)
	cleanProjectName := filepath.Clean(projectName)
	command := exec.Command("go", "mod", "init", cleanGithubAcc+"/"+cleanProjectName)
	command.Dir = projectName
	err = command.Run()
	if err != nil {
		eLogger.Printf("Failed to run go mod init: %v", err)

		return err
	}

	if err := os.Chdir(projectName); err != nil {
		eLogger.Printf("Failed to change directory: %v", err)

		return err
	}

	echoCmd := exec.Command("go", "get", "-u", "github.com/labstack/echo/v4")
	err = echoCmd.Run()
	if err != nil {
		eLogger.Printf("Failed to fetch the latest version of Echo: %v", err)

		return err
	}

	zapCmd := exec.Command("go", "get", "-u", "go.uber.org/zap")
	err = zapCmd.Run()
	if err != nil {
		eLogger.Printf("Failed to fetch the latest version of Zap logger: %v", err)

		return err
	}

	godotenvCmd := exec.Command("go", "get", "-u", "github.com/joho/godotenv")
	err = godotenvCmd.Run()
	if err != nil {
		eLogger.Printf("Failed to fetch the latest version of Godotenv: %v", err)

		return err
	}

	mainTmpl, err := fs.ReadFile(templates, "templates/main.txt")
	if err != nil {
		eLogger.Printf("Failed to read main template: %v", err)

		return err
	}
	mainTmplStr := string(mainTmpl)

	mainFilePath := filepath.Join(".", "main.go")
	err = os.WriteFile(mainFilePath, []byte(mainTmplStr), 0600)
	if err != nil {
		eLogger.Printf("Failed to write main.go file: %v", err)

		return err
	}

	return nil
}
