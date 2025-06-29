package files

import (
	"fmt"
	"os"
	"path/filepath"
)

// CopyOperation represents a file copy operation.
type CopyOperation struct {
	src string
	dst string
}

// NewCopyOperation creates a new copy operation.
func NewCopyOperation(src, dst string) *CopyOperation {
	return &CopyOperation{src: src, dst: dst}
}

func (co *CopyOperation) Source() string {
	return co.src
}

func (co *CopyOperation) Destination() string {
	return co.dst
}

func (co *CopyOperation) Type() OperationType {
	return OperationCopy
}

func (co *CopyOperation) Execute(fs FilesService) error {
	return fs.Copy(co.src, co.dst)
}

func (co *CopyOperation) Rollback(fs FilesService) error {
	// For copy operations, rollback removes the destination file
	if err := os.Remove(co.dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove copied file %q: %w", co.dst, err)
	}
	return nil
}

// MoveOperation represents a file move operation.
type MoveOperation struct {
	src string
	dst string
}

// NewMoveOperation creates a new move operation.
func NewMoveOperation(src, dst string) *MoveOperation {
	return &MoveOperation{src: src, dst: dst}
}

func (mo *MoveOperation) Source() string {
	return mo.src
}

func (mo *MoveOperation) Destination() string {
	return mo.dst
}

func (mo *MoveOperation) Type() OperationType {
	return OperationMove
}

func (mo *MoveOperation) Execute(fs FilesService) error {
	// Ensure destination directory exists (similar to how move command works)
	if err := fs.EnsureDir(filepath.Dir(mo.dst), 0o755); err != nil {
		return err
	}
	
	// Perform the move (rename)
	if err := os.Rename(mo.src, mo.dst); err != nil {
		return fmt.Errorf("move %q to %q: %w", mo.src, mo.dst, err)
	}
	return nil
}

func (mo *MoveOperation) Rollback(fs FilesService) error {
	// For move operations, rollback moves the file back to its original location
	if err := os.Rename(mo.dst, mo.src); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to restore moved file %q to %q: %w", mo.dst, mo.src, err)
	}
	return nil
}