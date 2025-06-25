package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tmunayyer/gocamelpack/deps"
	"github.com/spf13/cobra"
)

func createRootCmd(dependencies *deps.AppDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "gocamelpack",
		Short: "gocamelpack is your CLI companion",
		Long:  `gocamelpack is a tool to help you move and rename large amounts of files based on file metadata.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from Cobra!")
		},
	}
}

func createCopyCmd(d *deps.AppDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "copy [source] [destination]",
		Short: "Copy files from source to destination",
		Long:  "Source can either be a file or a directory. Destination must be a directory.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			dst := args[1]

			var filesToCopy []string
			if d.Files.IsFile(src) {
				filesToCopy = append(filesToCopy, src)
			} else if d.Files.IsDirectory(src) {
				filepaths, err := d.Files.ReadDirectory(src)
				if err != nil {
					return fmt.Errorf("error reading directory: %v", err)
				}

				filesToCopy = append(filesToCopy, filepaths...)
			} else {
				return fmt.Errorf("unknown src argument")
			}

			metadata := d.Files.GetFileTags(filesToCopy)

			var destinations []string
			for _, md := range metadata {
				d, err := d.Files.DestinationFromMetadata(md, dst)
				if err != nil {
					return fmt.Errorf("error creating destination filepaths: %v", err)
				}

				destinations = append(destinations, d)
			}

			jsonBytes, err := json.MarshalIndent(destinations, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			fmt.Println(string(jsonBytes))

			fmt.Printf("Copying (%v) files from %s to %s\n", len(filesToCopy), src, dst)

			return nil
		},
	}
}

// var moveCmd = &cobra.Command{
// 	Use:   "move [source] [destination]",
// 	Short: "Move file or files from source to destination. This will rename the original files.",
// 	Long:  "Source can either be a file or a directory. Destination must be a directory.",
// 	Args:  cobra.ExactArgs(2),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		src := args[0]
// 		dst := args[1]
// 		fmt.Printf("Moving from %s to %s\n", src, dst)
// 	},
// }

func createReadCmd(d *deps.AppDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "read [source]",
		Short: "This will read a specified file and print the metadata.",
		Long:  "Source must be a filepath.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			if !d.Files.IsFile(src) {
				return fmt.Errorf("src is not a file")
			}

			metadata := d.Files.GetFileTags([]string{src})

			jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			fmt.Println(string(jsonBytes))

			return nil
		},
	}
}

func Execute(dependencies *deps.AppDeps) {
	rootCmd := createRootCmd(dependencies)

	rootCmd.AddCommand(createReadCmd(dependencies))
	rootCmd.AddCommand(createCopyCmd(dependencies))
	// rootCmd.AddCommand(moveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
