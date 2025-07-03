     Overview

     Implement a progress bar (not spinner) for file operations in gocamelpack, integrating with both transactional and 
     non-transactional copy/move operations.

     Key Integration Points

     - Transaction execution loops (files/transaction_manager.go:64-90)
     - Non-transactional copy/move loops (cmd/cmd.go:92-114, 157-182)
     - File collection phase (determines total work)

     TODOs (Small, Achievable Steps)

     TODO 1: Create Progress Interface & Tests

     - Create progress/progress.go with ProgressReporter interface
     - Write comprehensive unit tests for progress reporting logic
     - Test: progress calculation, display formatting, completion detection
     - Commit: "Add progress reporter interface with tests"

     TODO 2: Implement Simple Progress Bar

     - Implement basic progress bar using only standard library
     - ASCII-based bar with percentage and operation counts
     - Write tests for bar rendering and state management
     - Commit: "Implement ASCII progress bar with tests"

     TODO 3: Add Progress Hooks to Transaction System

     - Modify FileTransaction.Execute() to accept optional progress reporter
     - Add progress updates between operations
     - Write tests to verify progress reporting during transactions
     - Commit: "Add progress reporting to transaction system"

     TODO 4: Add Progress to Non-Transactional Operations

     - Modify copy/move command loops to use progress reporter
     - Add progress updates in performTransactionalCopy/Move functions
     - Test both atomic and non-atomic operation progress
     - Commit: "Add progress reporting to non-transactional operations"

     TODO 5: Integrate with CLI Commands

     - Add --progress flag to copy/move commands
     - Wire progress reporter to command execution
     - Write integration tests for CLI progress display
     - Commit: "Add progress flag to CLI commands"

     TODO 6: Add Progress During File Collection

     - Show progress while scanning/collecting source files
     - Update progress bar during metadata extraction
     - Test progress reporting for large directory scans
     - Commit: "Add progress reporting during file collection"

     TODO 7: Handle Edge Cases & Polish

     - Progress bar behavior with single files vs many files
     - Error handling with progress display
     - Clean progress bar cleanup on completion/error
     - Commit: "Polish progress bar edge cases and error handling"

     Technical Approach

     - No external dependencies - use only Go standard library
     - Interface-driven design - ProgressReporter interface for testability
     - Backward compatible - progress is opt-in via flags
     - Clean integration - minimal changes to existing code paths
     - TDD focused - tests written before implementation

     Validation

     - Run make test after each TODO
     - Run make lint to ensure code quality
     - Manual testing with small and large file sets