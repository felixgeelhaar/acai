package localcache

import (
	"encoding/json"
	"testing"
)

func TestProsemirrorToPlainText(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "simple paragraph",
			raw:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello world"}]}]}`,
			want: "Hello world",
		},
		{
			name: "heading and paragraph",
			raw:  `{"type":"doc","content":[{"type":"heading","content":[{"type":"text","text":"Title"}]},{"type":"paragraph","content":[{"type":"text","text":"Body text"}]}]}`,
			want: "Title\nBody text",
		},
		{
			name: "bullet list",
			raw:  `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"First"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Second"}]}]}]}]}`,
			want: "- First\n- Second",
		},
		{
			name: "hard break",
			raw:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Line 1"},{"type":"hardBreak"},{"type":"text","text":"Line 2"}]}]}`,
			want: "Line 1\nLine 2",
		},
		{
			name: "nested list",
			raw:  `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Outer"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Inner"}]}]}]}]}]}]}`,
			want: "- Outer\n  - Inner",
		},
		{
			name: "empty doc",
			raw:  `{"type":"doc","content":[]}`,
			want: "",
		},
		{
			name: "nil input",
			raw:  "",
			want: "",
		},
		{
			name: "invalid JSON",
			raw:  `{invalid`,
			want: "",
		},
		{
			name: "unknown node types recurse into children",
			raw:  `{"type":"doc","content":[{"type":"customBlock","content":[{"type":"paragraph","content":[{"type":"text","text":"Nested in unknown"}]}]}]}`,
			want: "Nested in unknown",
		},
		{
			name: "multiple paragraphs",
			raw:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Para 1"}]},{"type":"paragraph","content":[{"type":"text","text":"Para 2"}]}]}`,
			want: "Para 1\nPara 2",
		},
		{
			name: "ordered list treated like bullet list",
			raw:  `{"type":"doc","content":[{"type":"orderedList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Step 1"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Step 2"}]}]}]}]}`,
			want: "- Step 1\n- Step 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw json.RawMessage
			if tt.raw != "" {
				raw = json.RawMessage(tt.raw)
			}
			got := prosemirrorToPlainText(raw)
			if got != tt.want {
				t.Errorf("prosemirrorToPlainText() = %q, want %q", got, tt.want)
			}
		})
	}
}
