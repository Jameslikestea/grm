// Code generated by entc, DO NOT EDIT.

package reference

const (
	// Label holds the string label denoting the reference type in the database.
	Label = "reference"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldPackage holds the string denoting the package field in the database.
	FieldPackage = "package"
	// FieldRef holds the string denoting the ref field in the database.
	FieldRef = "ref"
	// FieldHash holds the string denoting the hash field in the database.
	FieldHash = "hash"
	// Table holds the table name of the reference in the database.
	Table = "references"
)

// Columns holds all SQL columns for reference fields.
var Columns = []string{
	FieldID,
	FieldPackage,
	FieldRef,
	FieldHash,
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
