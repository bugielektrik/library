package postgres

import (
	"fmt"
	"reflect"
	"strings"
)

// PrepareUpdateArgs builds SQL SET clauses and arguments for UPDATE queries
// using reflection to handle any struct with db tags.
//
// Example usage:
//
//	type Book struct {
//	    Name  *string `db:"name"`
//	    Genre *string `db:"genre"`
//	    ISBN  *string `db:"isbn"`
//	}
//
//	book := Book{Name: stringPtr("New Title")}
//	sets, args := PrepareUpdateArgs(book)
//	// sets = ["name=$1"]
//	// args = ["New Title"]
func PrepareUpdateArgs(data interface{}) ([]string, []interface{}) {
	var sets []string
	var args []interface{}

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return sets, args
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get db tag
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Handle comma-separated tags (e.g., "name,omitempty")
		if idx := strings.Index(dbTag, ","); idx != -1 {
			dbTag = dbTag[:idx]
		}

		// Skip if field is nil pointer
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Get actual value
		var value interface{}
		if field.Kind() == reflect.Ptr {
			value = field.Elem().Interface()
		} else {
			value = field.Interface()
		}

		// Skip zero values for non-pointer fields
		if !isPointer(field) && isZeroValue(field) {
			continue
		}

		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s=$%d", dbTag, len(args)))
	}

	return sets, args
}

// isPointer checks if a reflect.Value is a pointer
func isPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr
}

// isZeroValue checks if a reflect.Value is its zero value
func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	return reflect.DeepEqual(v.Interface(), zero)
}

// BuildUpdateQuery builds a complete UPDATE SQL query
func BuildUpdateQuery(tableName string, data interface{}, idColumn string, id interface{}) (string, []interface{}) {
	sets, args := PrepareUpdateArgs(data)
	if len(sets) == 0 {
		return "", nil
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		tableName,
		strings.Join(sets, ", "),
		idColumn,
		len(args),
	)

	return query, args
}

// BuildInsertQuery builds an INSERT query with RETURNING clause
func BuildInsertQuery(tableName string, data interface{}, returningColumn string) (string, []interface{}) {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get db tag
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" || dbTag == "-" || dbTag == returningColumn {
			continue
		}

		// Handle comma-separated tags
		if idx := strings.Index(dbTag, ","); idx != -1 {
			dbTag = dbTag[:idx]
		}

		// Get value
		var value interface{}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				value = nil
			} else {
				value = field.Elem().Interface()
			}
		} else {
			value = field.Interface()
		}

		columns = append(columns, dbTag)
		args = append(args, value)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	if returningColumn != "" {
		query += " RETURNING " + returningColumn
	}

	return query, args
}
