package forms

// Define a new errrors type, which we will use to hold the validation error message for forms. The name of the form will be used as the key in the map.
type errors map[string][]string

// Implement an Add() method to add error messages for a given field to the map.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Implement a Get() method to retrieve the first error mesage for a given field from the map.
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}

	return es[0]
}
