package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xlcbingo1999/example-redis/bloom"
)

var bloomDemoCmd = &cobra.Command{
	Use:   "bloom_demo",
	Short: "Run bloom_demo",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalln("Recover err", err)
			}
		}()

		bloom.RunBloom()
	},
}

func init() {
	rootCmd.AddCommand(bloomDemoCmd)
}
