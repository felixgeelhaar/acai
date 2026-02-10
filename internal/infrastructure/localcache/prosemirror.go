package localcache

import (
	"encoding/json"
	"strings"
)

// prosemirrorNode represents a node in ProseMirror's JSON document model.
type prosemirrorNode struct {
	Type    string            `json:"type"`
	Content []prosemirrorNode `json:"content,omitempty"`
	Text    string            `json:"text,omitempty"`
	Attrs   json.RawMessage   `json:"attrs,omitempty"`
}

// prosemirrorToPlainText converts a ProseMirror JSON document to plain text.
// It recursively walks the node tree, extracting text content.
// Unknown node types are silently recursed into.
func prosemirrorToPlainText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var doc prosemirrorNode
	if err := json.Unmarshal(raw, &doc); err != nil {
		return ""
	}

	var b strings.Builder
	renderNode(&b, doc, 0)
	return strings.TrimSpace(b.String())
}

// renderNode recursively renders a ProseMirror node into the string builder.
func renderNode(b *strings.Builder, node prosemirrorNode, listDepth int) {
	switch node.Type {
	case "text":
		b.WriteString(node.Text)

	case "hardBreak":
		b.WriteByte('\n')

	case "paragraph":
		renderChildren(b, node, listDepth)
		b.WriteByte('\n')

	case "heading":
		renderChildren(b, node, listDepth)
		b.WriteByte('\n')

	case "bulletList", "orderedList":
		renderChildren(b, node, listDepth+1)

	case "listItem":
		b.WriteString(strings.Repeat("  ", listDepth-1))
		b.WriteString("- ")
		renderChildren(b, node, listDepth)

	default:
		// Unknown nodes: recurse into children silently.
		renderChildren(b, node, listDepth)
	}
}

// renderChildren renders all child nodes.
func renderChildren(b *strings.Builder, node prosemirrorNode, listDepth int) {
	for _, child := range node.Content {
		renderNode(b, child, listDepth)
	}
}
