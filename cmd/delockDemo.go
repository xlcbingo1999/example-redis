package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xlcbingo1999/example-redis/delock"
)

var delockDemoCmd = &cobra.Command{
	Use:   "delock_demo",
	Short: "Run delock_demo",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalln("Recover err", err)
			}
		}()

		delock.RunDelock()
	},
}

func init() {
	rootCmd.AddCommand(delockDemoCmd)
}
