package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/xlcbingo1999/example-redis/pipeline"
)

var pipelineDemoCmd = &cobra.Command{
	Use:   "pipeline_demo",
	Short: "Run pipeline_demo",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalln("Recover err", err)
			}
		}()

		pipeline.RunPipeline()
	},
}

func init() {
	rootCmd.AddCommand(pipelineDemoCmd)
}
