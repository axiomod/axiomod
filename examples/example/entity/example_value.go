package entity

// ExampleValue represents a value object for the Example entity
type ExampleValue struct {
	Type  string
	Count int
	Tags  []string
}

// NewExampleValue creates a new ExampleValue
func NewExampleValue(valueType string, count int, tags []string) ExampleValue {
	return ExampleValue{
		Type:  valueType,
		Count: count,
		Tags:  tags,
	}
}

// Equals checks if two ExampleValue objects are equal
func (v ExampleValue) Equals(other ExampleValue) bool {
	if v.Type != other.Type || v.Count != other.Count {
		return false
	}

	if len(v.Tags) != len(other.Tags) {
		return false
	}

	// Check if all tags are the same
	tagMap := make(map[string]struct{}, len(v.Tags))
	for _, tag := range v.Tags {
		tagMap[tag] = struct{}{}
	}

	for _, tag := range other.Tags {
		if _, ok := tagMap[tag]; !ok {
			return false
		}
	}

	return true
}

// HasTag checks if the value has a specific tag
func (v ExampleValue) HasTag(tag string) bool {
	for _, t := range v.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag to the value if it doesn't already exist
func (v *ExampleValue) AddTag(tag string) {
	if !v.HasTag(tag) {
		v.Tags = append(v.Tags, tag)
	}
}

// RemoveTag removes a tag from the value
func (v *ExampleValue) RemoveTag(tag string) {
	for i, t := range v.Tags {
		if t == tag {
			v.Tags = append(v.Tags[:i], v.Tags[i+1:]...)
			return
		}
	}
}
