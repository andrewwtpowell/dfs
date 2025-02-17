package cli

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var server = os.Getenv("DFS_SERVER_ADDR")

var rootCmd = &cobra.Command {
    Use: "dfs",
    Short: "DFS Client Command Line Interface",
    Args: cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {

    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
