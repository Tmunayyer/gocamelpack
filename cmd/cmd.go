package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tmunayyer/gocamelpack/deps"
	"github.com/Tmunayyer/gocamelpack/files"
	"github.com/spf13/cobra"
)

// dirPerm represents rwxr-xr-x — used when creating target directories during moves.
const dirPerm = 0o755

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

func createCopyCmd(d *deps.AppDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy [source] [destination]",
		Short: "Copy files from source to destination",
		Long:  "Source may be a file or directory. Destination is the root directory under which files will be placed according to their metadata.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcInput := args[0]
			dstRoot := args[1] // base directory passed to DestinationFromMetadata
			// flags
			// jobs, _ := cmd.Flags().GetUint("jobs") // not yet used
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			atomic, _ := cmd.Flags().GetBool("atomic")

			// resolve source to an absolute path so tests expecting "abs/..." match
			src, err := filepath.Abs(srcInput)
			if err != nil {
				return fmt.Errorf("resolving %q: %w", srcInput, err)
			}

			sources, err := collectSources(d.Files, src)
			if err != nil {
				return err
			}

			if atomic {
				return performTransactionalCopy(d.Files, sources, dstRoot, dryRun, overwrite, cmd)
			}

			// Original non-transactional behavior
			for _, src := range sources {
				dst, err := destFromMetadata(d.Files, src, dstRoot)
				if err != nil {
					return err
				}

				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "Would move %s → %s\n", src, dst)
					continue
				}
				if !overwrite {
					if err := d.Files.ValidateCopyArgs(src, dst); err != nil {
						return err
					}
				}
				// Move is copy+delete or os.Rename; simplest:
				if err := d.Files.Copy(src, dst); err != nil {
					return err
				}
			}

			fmt.Printf("Copied %d file(s).\n", len(sources))
			return nil
		},
		// flag definitions added after struct literal
	}

	// CLI flags
	cmd.Flags().Bool("dry-run", false, "Show what would be copied without doing it")
	cmd.Flags().Bool("overwrite", false, "Allow overwriting existing files in destination")
	cmd.Flags().Bool("atomic", false, "Perform all-or-nothing copy with rollback on failure")
	cmd.Flags().Uint("jobs", 1, "Number of concurrent copy workers (currently only 1 is used)")

	return cmd
}

func createMoveCmd(d *deps.AppDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move [source] [destination]",
		Short: "Move files from source to destination (original files are renamed)",
		Long:  "Source may be a file or directory. Destination is the root directory under which files will be placed according to their metadata.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcInput := args[0]
			dstRoot := args[1]

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			atomic, _ := cmd.Flags().GetBool("atomic")

			srcAbs, err := filepath.Abs(srcInput)
			if err != nil {
				return fmt.Errorf("resolving %q: %w", srcInput, err)
			}

			sources, err := collectSources(d.Files, srcAbs)
			if err != nil {
				return err
			}

			if atomic {
				return performTransactionalMove(d.Files, sources, dstRoot, dryRun, overwrite, cmd)
			}

			// Original non-transactional behavior
			for _, src := range sources {
				dst, err := destFromMetadata(d.Files, src, dstRoot)
				if err != nil {
					return err
				}

				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "Would move %s -> %s\n", src, dst)
					continue
				}
				// Validate unless overwrite flag is set
				if !overwrite {
					if err := d.Files.ValidateCopyArgs(src, dst); err != nil {
						return err
					}
				}
				// Ensure destination directory exists
				if err := d.Files.EnsureDir(filepath.Dir(dst), dirPerm); err != nil {
					return err
				}
				// Perform the move (rename)
				if err := os.Rename(src, dst); err != nil {
					return err
				}
			}

			fmt.Printf("Moved %d file(s).\n", len(sources))
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be moved without doing it")
	cmd.Flags().Bool("overwrite", false, "Allow overwriting existing files in destination")
	cmd.Flags().Bool("atomic", false, "Perform all-or-nothing move with rollback on failure")

	return cmd
}

// performTransactionalCopy handles atomic copy operations using transactions.
func performTransactionalCopy(fs files.FilesService, sources []string, dstRoot string, dryRun, overwrite bool, cmd *cobra.Command) error {
	// Create a new transaction
	tx := fs.NewTransaction(overwrite)
	
	// Plan all operations
	for _, src := range sources {
		dst, err := destFromMetadata(fs, src, dstRoot)
		if err != nil {
			return err
		}
		
		if err := tx.AddCopy(src, dst); err != nil {
			return err
		}
	}
	
	// Validate all operations
	if err := tx.Validate(); err != nil {
		return err
	}
	
	// Handle dry-run mode
	if dryRun {
		for _, op := range tx.Operations() {
			fmt.Fprintf(cmd.OutOrStdout(), "Would copy %s → %s\n", op.Source(), op.Destination())
		}
		return nil
	}
	
	// Execute the transaction
	if err := tx.Execute(); err != nil {
		return err
	}
	
	fmt.Printf("Atomically copied %d file(s).\n", len(sources))
	return nil
}

// performTransactionalMove handles atomic move operations using transactions.
func performTransactionalMove(fs files.FilesService, sources []string, dstRoot string, dryRun, overwrite bool, cmd *cobra.Command) error {
	// Create a new transaction
	tx := fs.NewTransaction(overwrite)
	
	// Plan all operations
	for _, src := range sources {
		dst, err := destFromMetadata(fs, src, dstRoot)
		if err != nil {
			return err
		}
		
		if err := tx.AddMove(src, dst); err != nil {
			return err
		}
	}
	
	// Validate all operations
	if err := tx.Validate(); err != nil {
		return err
	}
	
	// Handle dry-run mode
	if dryRun {
		for _, op := range tx.Operations() {
			fmt.Fprintf(cmd.OutOrStdout(), "Would move %s → %s\n", op.Source(), op.Destination())
		}
		return nil
	}
	
	// Execute the transaction
	if err := tx.Execute(); err != nil {
		return err
	}
	
	fmt.Printf("Atomically moved %d file(s).\n", len(sources))
	return nil
}

func Execute(dependencies *deps.AppDeps) {
	rootCmd := createRootCmd(dependencies)

	rootCmd.AddCommand(createReadCmd(dependencies))
	rootCmd.AddCommand(createCopyCmd(dependencies))
	rootCmd.AddCommand(createMoveCmd(dependencies))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
