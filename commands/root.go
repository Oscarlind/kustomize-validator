package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Oscarlind/kustomize-validator/validate"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:  "kustomization-validator",
	Long: "A tool to validate Kustomization files",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No arguments provided")
			return
		}

		fmt.Println("Validating Kustomization files", args[0])

		isVerbose := cmd.Flag("verbose").Value.String() == "true"
		isErrorOnly := cmd.Flag("error-only").Value.String() == "true"

		msgChan := validate.KustomizeBuild(args[0])

		ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
		defer cf()

		totalCounter := 0
		successCounter := 0
		failureCounter := 0
	BREAK:
		for {
			select {
			case <-ctx.Done():
				break BREAK
			case msg, ok := <-msgChan:
				if !ok {
					break BREAK
				}
				totalCounter++
				if msg.Err != nil {
					failureCounter++
				} else {
					successCounter++
				}
				fmt.Print(msg.Msg(isErrorOnly, isVerbose))
			}
		}
		fmt.Println("Total: ", validate.ColorF(validate.ColorBlue, "%d", totalCounter))
		fmt.Println("Success: ", validate.ColorF(validate.ColorGreen, "%d", successCounter))
		fmt.Println("Error: ", validate.ColorF(validate.ColorRed, "%d", failureCounter))
		fmt.Println("Failed in %: ", validate.ColorF(validate.ColorRed, "%.2f%%", float64(failureCounter)/float64(totalCounter)*100))
	},
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolP("error-only", "e", false, "whether we should only log errors")
}
