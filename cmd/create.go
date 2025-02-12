package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/jasonuc/gignr/internal/templates"
	"github.com/jasonuc/gignr/internal/utils"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create <template> [templates]...",
	Example: "gignr create gh:Go tt:clion my-template",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Generate a .gitignore file using one or more templates",
	Long: `The create command generates a .gitignore file based on one or more templates of your choice.

Available templates are identified by prefixes:
  - tt: TopTal templates
  - gh: GitHub templates
  - ghg: GitHub Global templates
  - ghc: GitHub Community templates
  - (no prefix) → Fetch from local saved templates
`,
	Run: func(cmd *cobra.Command, args []string) {
		var mergedContent strings.Builder

		templates.InitGitHubClient("")

		// Load user-added repositories
		repos := templates.LoadCustomRepositories()

		for _, arg := range args {
			var content []byte
			var err error
			var source string

			if strings.Contains(arg, ":") {

				req := strings.SplitAfter(arg, ":")
				reqPrefix := strings.TrimSpace(req[0][:len(req[0])-1])
				templateName := strings.TrimSpace(req[1])

				var owner, repo, path string
				switch reqPrefix {
				case "tt":
					owner, repo, path = "toptal", "gitignore", "templates"
				case "gh":
					owner, repo, path = "github", "gitignore", ""
				case "ghc":
					owner, repo, path = "github", "gitignore", "community"
				case "ghg":
					owner, repo, path = "github", "gitignore", "Global"
				default:
					// If the prefix is a user-defined repo, resolve its URL
					if repoURL, exists := repos[reqPrefix]; exists {
						owner, repo = utils.ExtractRepoDetails(repoURL)
						path = ""
					} else {
						fmt.Printf("Unknown template prefix or missing repository: %s\n", reqPrefix)
						continue
					}
				}

				templateList, err := templates.FetchTemplates(owner, repo, path)
				if err != nil {
					utils.PrintError(fmt.Sprintf("Unable to fetch templates from %s: %v\n", reqPrefix, err))
					utils.PrintAlert("No .gitignore file created.")
					return
				}

				var downloadURL string
				for _, tmpl := range templateList {
					if strings.EqualFold(tmpl.Name, templateName+".gitignore") {
						downloadURL = tmpl.DownloadURL
						break
					}
				}

				if downloadURL == "" {
					utils.PrintError(fmt.Sprintf("Template %s not found in %s.\n", templateName, reqPrefix))
					continue
				}

				content, err = templates.GetTemplateContent(downloadURL)
				source = reqPrefix
			} else {
				// Attemt to fetch from local storage if no prefix is provided
				content, err = templates.GetLocalTemplate(arg)
				source = "local"
			}

			if err != nil {
				utils.PrintError(fmt.Sprintf("Unable to fetch content for %s: %v\n", arg, err))
				continue
			}

			config := box.Config{Px: 1, Py: 1, Type: "", TitlePos: "Inside"}
			boxNew := box.Box{TopRight: "*", TopLeft: "*", BottomRight: "*", BottomLeft: "*", Horizontal: "-", Vertical: "|", Config: config}

			mergedContent.WriteString(boxNew.String("", fmt.Sprintf(" %s Template (%s)", strings.ToUpper(arg), strings.ToUpper(source))))
			mergedContent.Write(content)
			mergedContent.WriteString("\n\n")
		}

		err := os.WriteFile(".gitignore", []byte(mergedContent.String()), 0644)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to write .gitignore file: %v\n", err))
			return
		}

		utils.PrintSuccess("Created .gitignore!")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
