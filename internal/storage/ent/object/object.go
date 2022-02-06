// Code generated by entc, DO NOT EDIT.

package object

const (
	// Label holds the string label denoting the object type in the database.
	Label = "object"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldPackage holds the string denoting the package field in the database.
	FieldPackage = "package"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldHash holds the string denoting the hash field in the database.
	FieldHash = "hash"
	// FieldContent holds the string denoting the content field in the database.
	FieldContent = "content"
	// Table holds the table name of the object in the database.
	Table = "objects"
)

// Columns holds all SQL columns for object fields.
var Columns = []string{
	FieldID,
	FieldPackage,
	FieldType,
	FieldHash,
	FieldContent,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}
