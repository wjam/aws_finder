package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	exeName := os.Args[0][strings.LastIndex(os.Args[0], string(os.PathSeparator))+1:]
	root := &cobra.Command{
		Use: exeName,
	}

	var shellCompletion = &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generates shell completion scripts",
		Long: fmt.Sprintf(`To load completion run

source <(%[1]s completion bash)
source <(%[1]s completion zsh)
`, exeName),
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"bash", "zsh"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell %s", args[0])
			}
		},
	}

	root.AddCommand(shellCompletion)
	root.AddCommand(instanceByIp)
	root.AddCommand(vpcByCidr)
	root.AddCommand(vpcEndpointService)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
