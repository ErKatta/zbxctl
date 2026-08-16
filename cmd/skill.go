package cmd

import (
	"fmt"

	"github.com/ErKatta/zbxctl/pkg/skill"
	"github.com/spf13/cobra"
)

var (
	skillGlobalFlag    bool
	skillWorkspaceFlag bool
	skillAllFlag       bool
	skillAgentFlag     string
)

var skillCmd = &cobra.Command{
	Use:     "skill",
	Aliases: []string{"skills"},
	Short:   "Manage AI agent skills for zbxctl (install, list, show, export)",
	Long: `skill manages packaged skills for AI coding agents (Antigravity, Claude, Cursor, Aider, Copilot) and human developers.
Skills provide domain knowledge, workflow instructions, and automation recipes.`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available built-in zbxctl skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		type skillItem struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		var items []skillItem
		for _, s := range skill.BuiltinSkills {
			items = append(items, skillItem{
				Name:        s.Name,
				Description: s.Description,
			})
		}
		return formatter.Print(items)
	},
}

var skillInstallCmd = &cobra.Command{
	Use:   "install [skill-name]",
	Short: "Install a zbxctl skill into AI agent customizations directory",
	Long: `install copies a skill into target AI agent customizations directories (Antigravity/Gemini, Claude, Cursor, zbxctl).
Use --agent=claude, --agent=cursor, --agent=antigravity, or --agent=all. Use --all to install all built-in skills.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !skillAllFlag && len(args) == 0 {
			return fmt.Errorf("please specify a skill name to install (e.g. 'zbxctl skill install zabbix-automation') or use '--all'")
		}

		isGlobal := !skillWorkspaceFlag
		targetDirs, err := skill.GetAgentSkillDirs(skillAgentFlag, isGlobal)
		if err != nil {
			return fmt.Errorf("failed to determine target skills directories: %w", err)
		}

		skillsToInstall := []string{}
		if skillAllFlag {
			for name := range skill.BuiltinSkills {
				skillsToInstall = append(skillsToInstall, name)
			}
		} else {
			skillsToInstall = append(skillsToInstall, args[0])
		}

		var installedPaths []string
		for _, targetDir := range targetDirs {
			for _, name := range skillsToInstall {
				path, err := skill.InstallSkill(name, targetDir)
				if err != nil {
					return err
				}
				installedPaths = append(installedPaths, path)
			}
		}

		fmt.Printf("Successfully installed skills across %d target directory location(s):\n", len(targetDirs))
		for _, p := range installedPaths {
			fmt.Printf(" - %s\n", p)
		}
		return nil
	},
}

var skillShowCmd = &cobra.Command{
	Use:   "show <skill-name>",
	Short: "Show content of a zbxctl skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s, ok := skill.BuiltinSkills[name]
		if !ok {
			return fmt.Errorf("skill %q not found. Use 'zbxctl skill list' to view available skills", name)
		}
		fmt.Println(s.Content)
		return nil
	},
}

var skillExportCmd = &cobra.Command{
	Use:   "export <skill-name>",
	Short: "Export a skill to workspace .agents/skills/ directory for repository sharing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		targetDir, err := skill.GetWorkspaceSkillsDir()
		if err != nil {
			return fmt.Errorf("failed to determine workspace skills directory: %w", err)
		}
		path, err := skill.InstallSkill(name, targetDir)
		if err != nil {
			return err
		}
		fmt.Printf("Successfully exported skill %q to workspace repository at %s\n", name, path)
		return nil
	},
}

func init() {
	skillInstallCmd.Flags().BoolVar(&skillGlobalFlag, "global", true, "install to global user skills directory (~/.gemini/config/skills)")
	skillInstallCmd.Flags().BoolVar(&skillWorkspaceFlag, "workspace", false, "install to local workspace repository (.agents/skills)")
	skillInstallCmd.Flags().BoolVar(&skillAllFlag, "all", false, "install all built-in zbxctl skills")
	skillInstallCmd.Flags().StringVar(&skillAgentFlag, "agent", "all", "target AI agent (all, claude, cursor, antigravity)")

	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillShowCmd)
	skillCmd.AddCommand(skillExportCmd)

	RootCmd.AddCommand(skillCmd)
}
