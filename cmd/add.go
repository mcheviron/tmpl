package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

func Add(logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new project",
		Long:  `Add a new project with the Echo framework installed`,
		Run: func(cmd *cobra.Command, args []string) {
			projectName, err := getPrjName()
			if err != nil {
				logger.Error("Failed to get project name:", err)

				return
			}

			ecs, err := cmd.Flags().GetBool("ecs")
			if err != nil {
				logger.Error("Failed to get ECS flag:", err)

				return
			}

			gh, err := cmd.Flags().GetBool("gh")
			if err != nil {
				logger.Error("Failed to get GitHub flag:", err)

				return
			}

			var wg sync.WaitGroup

			if ecs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					generateECS(projectName, logger)
				}()
			}

			if gh {
				wg.Add(1)
				go func() {
					defer wg.Done()
					generateGH(projectName, logger)
				}()
			}

			wg.Wait()
		},
	}

	cmd.Flags().BoolP("ecs", "e", false, "Add templates for ECS CI deployment")
	cmd.Flags().BoolP("gh", "g", false, "Add the GitHub Actions folder with a certain template")

	return cmd
}

func getPrjName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := scanner.Text()
		moduleName := strings.Fields(firstLine)[1]
		projectName := path.Base(moduleName)

		return projectName, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read go.mod file: %w", err)
	}

	return "", fmt.Errorf("go.mod file is empty")
}
