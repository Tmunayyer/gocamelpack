package files

import (
	"fmt"

	"github.com/Tmunayyer/gocamelpack/progress"
)

// FileTransaction implements the Transaction interface.
type FileTransaction struct {
	fs          FilesService
	operations  []Operation
	completed   []Operation
	overwrite   bool
}

// NewTransaction creates a new file transaction.
func NewTransaction(fs FilesService, overwrite bool) Transaction {
	return &FileTransaction{
		fs:        fs,
		overwrite: overwrite,
	}
}

func (ft *FileTransaction) AddCopy(src, dst string) error {
	op := NewCopyOperation(src, dst)
	ft.operations = append(ft.operations, op)
	return nil
}

func (ft *FileTransaction) AddMove(src, dst string) error {
	op := NewMoveOperation(src, dst)
	ft.operations = append(ft.operations, op)
	return nil
}

func (ft *FileTransaction) Validate() error {
	for _, op := range ft.operations {
		if !ft.overwrite {
			if err := ft.fs.ValidateCopyArgs(op.Source(), op.Destination()); err != nil {
				return &TransactionError{
					Phase:     "planning",
					Operation: op,
					Err:       err,
				}
			}
		} else {
			// Basic validation even with overwrite
			if op.Source() == "" || op.Destination() == "" {
				return &TransactionError{
					Phase:     "planning",
					Operation: op,
					Err:       fmt.Errorf("source and destination must be provided"),
				}
			}
			if !ft.fs.IsFile(op.Source()) {
				return &TransactionError{
					Phase:     "planning",
					Operation: op,
					Err:       fmt.Errorf("source %q is not a regular file", op.Source()),
				}
			}
		}
	}
	return nil
}

func (ft *FileTransaction) Execute() error {
	return ft.ExecuteWithProgress(progress.NewNoOpReporter())
}

func (ft *FileTransaction) ExecuteWithProgress(reporter progress.ProgressReporter) error {
	// Reset completed operations
	ft.completed = ft.completed[:0]
	
	// Set up progress tracking
	reporter.SetTotal(len(ft.operations))
	reporter.SetCurrent(0)
	
	for i, op := range ft.operations {
		// Update progress message
		reporter.SetMessage(fmt.Sprintf("%s %s", op.Type(), op.Source()))
		
		if err := op.Execute(ft.fs); err != nil {
			// Execution failed, rollback completed operations
			rollbackErr := ft.Rollback()
			if rollbackErr != nil {
				// Return both errors
				return &TransactionError{
					Phase:     "execution",
					Operation: op,
					Err:       fmt.Errorf("execution failed: %v; rollback also failed: %v", err, rollbackErr),
				}
			}
			return &TransactionError{
				Phase:     "execution",
				Operation: op,
				Err:       err,
			}
		}
		
		ft.completed = append(ft.completed, op)
		
		// Update progress
		reporter.SetCurrent(i + 1)
	}
	
	// Mark as finished
	reporter.Finish()
	return nil
}

func (ft *FileTransaction) Rollback() error {
	var rollbackErrors []error
	
	// Rollback in reverse order
	for i := len(ft.completed) - 1; i >= 0; i-- {
		op := ft.completed[i]
		if err := op.Rollback(ft.fs); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to rollback %s %s->%s: %w", 
				op.Type(), op.Source(), op.Destination(), err))
		}
	}
	
	// Clear completed operations after rollback attempt
	ft.completed = ft.completed[:0]
	
	if len(rollbackErrors) > 0 {
		return &TransactionError{
			Phase: "rollback",
			Err:   fmt.Errorf("rollback errors: %v", rollbackErrors),
		}
	}
	
	return nil
}

func (ft *FileTransaction) Operations() []Operation {
	// Return a copy to prevent external modification
	ops := make([]Operation, len(ft.operations))
	copy(ops, ft.operations)
	return ops
}

func (ft *FileTransaction) Completed() []Operation {
	// Return a copy to prevent external modification
	ops := make([]Operation, len(ft.completed))
	copy(ops, ft.completed)
	return ops
}