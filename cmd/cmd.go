package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tmunayyer/gocamelpack/deps"
	"github.com/Tmunayyer/gocamelpack/files"
	"github.com/Tmunayyer/gocamelpack/progress"
	"github.com/spf13/cobra"
)

// dirPerm represents rwxr-xr-x — used when creating target directories during moves.
const dirPerm = 0o755

func createRootCmd(dependencies *deps.AppDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gocamelpack",
		Version: Version(),
		Short:   "gocamelpack is your CLI companion",
		Long:    fmt.Sprintf(`gocamelpack is a tool to help you move and rename large amounts of files based on file metadata.

Version: %s`, Version()),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from Cobra!")
		},
	}
	
	// Add custom version template that shows detailed build info
	cmd.SetVersionTemplate(BuildInfo() + "\n")
	
	return cmd
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
			showProgress, _ := cmd.Flags().GetBool("progress")

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
				return performTransactionalCopy(d.Files, sources, dstRoot, dryRun, overwrite, showProgress, cmd)
			}

			// Original non-transactional behavior with progress
			return performNonTransactionalCopy(d.Files, sources, dstRoot, dryRun, overwrite, showProgress, cmd)
		},
		// flag definitions added after struct literal
	}

	// CLI flags
	cmd.Flags().Bool("dry-run", false, "Show what would be copied without doing it")
	cmd.Flags().Bool("overwrite", false, "Allow overwriting existing files in destination")
	cmd.Flags().Bool("atomic", false, "Perform all-or-nothing copy with rollback on failure")
	cmd.Flags().Bool("progress", false, "Show progress bar during copy operations")
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
			showProgress, _ := cmd.Flags().GetBool("progress")

			srcAbs, err := filepath.Abs(srcInput)
			if err != nil {
				return fmt.Errorf("resolving %q: %w", srcInput, err)
			}

			sources, err := collectSources(d.Files, srcAbs)
			if err != nil {
				return err
			}

			if atomic {
				return performTransactionalMove(d.Files, sources, dstRoot, dryRun, overwrite, showProgress, cmd)
			}

			// Original non-transactional behavior with progress
			return performNonTransactionalMove(d.Files, sources, dstRoot, dryRun, overwrite, showProgress, cmd)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be moved without doing it")
	cmd.Flags().Bool("overwrite", false, "Allow overwriting existing files in destination")
	cmd.Flags().Bool("atomic", false, "Perform all-or-nothing move with rollback on failure")
	cmd.Flags().Bool("progress", false, "Show progress bar during move operations")

	return cmd
}

// performTransactionalCopy handles atomic copy operations using transactions.
func performTransactionalCopy(fs files.FilesService, sources []string, dstRoot string, dryRun, overwrite, showProgress bool, cmd *cobra.Command) error {
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

	// Execute the transaction with progress if requested
	if showProgress {
		reporter := progress.NewSimpleProgressBar(cmd.ErrOrStderr())
		if err := tx.ExecuteWithProgress(reporter); err != nil {
			return err
		}
	} else {
		if err := tx.Execute(); err != nil {
			return err
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Atomically copied %d file(s).\n", len(sources))
	return nil
}

// performTransactionalMove handles atomic move operations using transactions.
func performTransactionalMove(fs files.FilesService, sources []string, dstRoot string, dryRun, overwrite, showProgress bool, cmd *cobra.Command) error {
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

	// Execute the transaction with progress if requested
	if showProgress {
		reporter := progress.NewSimpleProgressBar(cmd.ErrOrStderr())
		if err := tx.ExecuteWithProgress(reporter); err != nil {
			return err
		}
	} else {
		if err := tx.Execute(); err != nil {
			return err
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Atomically moved %d file(s).\n", len(sources))
	return nil
}

// performNonTransactionalCopy handles non-atomic copy operations with progress reporting.
func performNonTransactionalCopy(fs files.FilesService, sources []string, dstRoot string, dryRun, overwrite, showProgress bool, cmd *cobra.Command) error {
	// Create progress reporter based on flag
	var reporter progress.ProgressReporter
	if showProgress {
		reporter = progress.NewSimpleProgressBar(cmd.ErrOrStderr())
	} else {
		reporter = progress.NewNoOpReporter()
	}
	reporter.SetTotal(len(sources))
	
	for i, src := range sources {
		dst, err := destFromMetadata(fs, src, dstRoot)
		if err != nil {
			return err
		}
		
		reporter.SetMessage(fmt.Sprintf("copy %s", src))
		
		if dryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "Would copy %s → %s\n", src, dst)
			reporter.Increment()
			continue
		}
		
		if !overwrite {
			if err := fs.ValidateCopyArgs(src, dst); err != nil {
				return err
			}
		}
		
		if err := fs.Copy(src, dst); err != nil {
			return err
		}
		
		reporter.SetCurrent(i + 1)
	}
	
	reporter.Finish()
	fmt.Fprintf(cmd.OutOrStdout(), "Copied %d file(s).\n", len(sources))
	return nil
}

// performNonTransactionalMove handles non-atomic move operations with progress reporting.
func performNonTransactionalMove(fs files.FilesService, sources []string, dstRoot string, dryRun, overwrite, showProgress bool, cmd *cobra.Command) error {
	// Create progress reporter based on flag
	var reporter progress.ProgressReporter
	if showProgress {
		reporter = progress.NewSimpleProgressBar(cmd.ErrOrStderr())
	} else {
		reporter = progress.NewNoOpReporter()
	}
	reporter.SetTotal(len(sources))
	
	for i, src := range sources {
		dst, err := destFromMetadata(fs, src, dstRoot)
		if err != nil {
			return err
		}
		
		reporter.SetMessage(fmt.Sprintf("move %s", src))
		
		if dryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "Would move %s → %s\n", src, dst)
			reporter.Increment()
			continue
		}
		
		// Validate unless overwrite flag is set
		if !overwrite {
			if err := fs.ValidateCopyArgs(src, dst); err != nil {
				return err
			}
		}
		
		// Ensure destination directory exists
		if err := fs.EnsureDir(filepath.Dir(dst), dirPerm); err != nil {
			return err
		}
		
		// Perform the move (rename)
		if err := os.Rename(src, dst); err != nil {
			return err
		}
		
		reporter.SetCurrent(i + 1)
	}
	
	reporter.Finish()
	fmt.Fprintf(cmd.OutOrStdout(), "Moved %d file(s).\n", len(sources))
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
