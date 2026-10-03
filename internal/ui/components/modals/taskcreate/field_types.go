package taskcreate

// FieldType represents which field is currently focused in the task creation form
type FieldType int

const (
	FieldTitle       FieldType = 0 // Title input field (required)
	FieldDescription FieldType = 1 // Description input field (optional)
	FieldFeature     FieldType = 2 // Feature selection field (optional)
	FieldPriority    FieldType = 3 // Priority input field (optional)
)
