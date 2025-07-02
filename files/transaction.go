package files

import (
	"fmt"

	"github.com/Tmunayyer/gocamelpack/progress"
)

// OperationType represents the type of file operation in a transaction.
type OperationType int

const (
	OperationCopy OperationType = iota
	OperationMove
)

func (ot OperationType) String() string {
	switch ot {
	case OperationCopy:
		return "copy"
	case OperationMove:
		return "move"
	default:
		return "unknown"
	}
}

// Operation represents a single file operation within a transaction.
type Operation interface {
	Source() string
	Destination() string
	Type() OperationType
	Execute(fs FilesService) error
	Rollback(fs FilesService) error
}

// TransactionError represents errors that occur during transaction operations.
type TransactionError struct {
	Phase     string      // "planning", "execution", "rollback"
	Operation Operation   // Operation that failed (may be nil for planning errors)
	Err       error       // Underlying error
}

func (te *TransactionError) Error() string {
	if te.Operation != nil {
		return fmt.Sprintf("transaction %s failed during %s %s -> %s: %v", 
			te.Phase, te.Operation.Type(), te.Operation.Source(), te.Operation.Destination(), te.Err)
	}
	return fmt.Sprintf("transaction %s failed: %v", te.Phase, te.Err)
}

func (te *TransactionError) Unwrap() error {
	return te.Err
}

// Transaction represents a collection of file operations that execute atomically.
type Transaction interface {
	// AddCopy plans a copy operation from src to dst.
	AddCopy(src, dst string) error
	
	// AddMove plans a move operation from src to dst.
	AddMove(src, dst string) error
	
	// Validate checks all planned operations for potential issues.
	// This should be called before Execute to catch problems early.
	Validate() error
	
	// Execute performs all planned operations atomically.
	// If any operation fails, all completed operations are rolled back.
	Execute() error
	
	// ExecuteWithProgress performs all planned operations atomically with progress reporting.
	// If any operation fails, all completed operations are rolled back.
	ExecuteWithProgress(reporter progress.ProgressReporter) error
	
	// Rollback undoes all completed operations in reverse order.
	// This is called automatically by Execute on failure.
	Rollback() error
	
	// Operations returns all planned operations for inspection (useful for dry-run).
	Operations() []Operation
	
	// Completed returns all operations that have been successfully executed.
	Completed() []Operation
}