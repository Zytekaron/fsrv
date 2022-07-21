package types

// OperationType represents the type of operation occurring on a particular file or directory.
type OperationType int8

const (
	// OperationRead represents an attempt to read the
	// contents of a file or list the files in a directory.
	OperationRead OperationType = iota
	// OperationWrite represents an attempt to create
	// a new file and write to it.
	OperationWrite
	// OperationModify represents an attempt to
	// modify the contents of an existing file.
	OperationModify
	// OperationDelete represents an attempt to
	// delete an existing file
	OperationDelete
)

func (opType OperationType) Int() int {
	switch opType {
	case OperationRead:
		return 0
	case OperationWrite:
		return 1
	case OperationModify:
		return 2
	case OperationDelete:
		return 3
	default:
		//todo: log error
		panic("OperationType to int conversion failure")
	}
}
