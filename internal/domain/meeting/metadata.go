package meeting

// Metadata is an immutable value object for extensible meeting metadata.
// All fields are defensively copied on construction and access.
type Metadata struct {
	tags         []string
	links        []string
	externalRefs map[string]string
}

func NewMetadata(tags, links []string, externalRefs map[string]string) Metadata {
	return Metadata{
		tags:         copyStrings(tags),
		links:        copyStrings(links),
		externalRefs: copyMap(externalRefs),
	}
}

func (m Metadata) Tags() []string              { return copyStrings(m.tags) }
func (m Metadata) Links() []string             { return copyStrings(m.links) }
func (m Metadata) ExternalRefs() map[string]string { return copyMap(m.externalRefs) }

func copyStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	copied := make([]string, len(s))
	copy(copied, s)
	return copied
}

func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	copied := make(map[string]string, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}
