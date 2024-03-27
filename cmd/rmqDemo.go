package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xlcbingo1999/example-redis/rmq"
)

var rmqDemoCmd = &cobra.Command{
	Use:   "rmq_demo",
	Short: "Run rmq_demo",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalln("Recover err", err)
			}
		}()

		rmq.RunRmqStream()
	},
}

func init() {
	rootCmd.AddCommand(rmqDemoCmd)
}
