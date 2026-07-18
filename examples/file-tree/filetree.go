package filetree

import "github.com/livetemplate/livetemplate"

// Node is one entry in the tree. A Node contains Nodes, which is what makes
// the template recursive: file-tree.tmpl defines a "node" block that invokes
// itself for each child, to whatever depth the data happens to go.
//
// Path is the stable identity used as data-key. Sibling names can repeat
// across directories (two different README.md files), so keying on Name would
// let the diff engine confuse them; the full path cannot collide.
type Node struct {
	Path     string
	Name     string
	IsDir    bool
	Expanded bool
	Starred  bool
	Children []Node
}

// State is per-session state — pure data, cloned per session. Each visitor
// expands and stars independently.
type State struct {
	Root Node
}

// Controller holds shared dependencies (none here) and exposes the action
// methods the template invokes by button name.
type Controller struct{}

// Mount seeds the tree on first render. Returning fresh data here rather than
// storing it on the Controller keeps the Controller free of per-session state.
func (c *Controller) Mount(s State, ctx *livetemplate.Context) (State, error) {
	if s.Root.Path == "" {
		s.Root = sampleTree()
	}
	return s, nil
}

// Toggle expands or collapses one directory. The button carries the node's
// path as its value, so the action knows which of the many rendered nodes was
// clicked without any per-node wiring.
func (c *Controller) Toggle(s State, ctx *livetemplate.Context) (State, error) {
	path := ctx.GetString("value")
	updateNode(&s.Root, path, func(n *Node) { n.Expanded = !n.Expanded })
	return s, nil
}

// Star marks a single file. This is the interesting one for the wire format:
// starring a leaf five levels down changes exactly one node, and the update
// sent to the browser is a nested ["u", key, ...] chain addressing that leaf —
// not a re-send of the branch containing it.
func (c *Controller) Star(s State, ctx *livetemplate.Context) (State, error) {
	path := ctx.GetString("value")
	updateNode(&s.Root, path, func(n *Node) { n.Starred = !n.Starred })
	return s, nil
}

// updateNode walks to the node with the given path and applies fn. It reports
// whether it found one, which lets the recursion stop at the first match.
func updateNode(n *Node, path string, fn func(*Node)) bool {
	if n.Path == path {
		fn(n)
		return true
	}
	for i := range n.Children {
		if updateNode(&n.Children[i], path, fn) {
			return true
		}
	}
	return false
}

// sampleTree is a fixed fixture so the recipe renders the same tree for
// everyone. Depth matters more than breadth here — the deepest file exists to
// show that a change down there stays scoped to that leaf.
func sampleTree() Node {
	return Node{
		Path: "/", Name: "project", IsDir: true, Expanded: true,
		Children: []Node{
			{
				Path: "/cmd", Name: "cmd", IsDir: true, Expanded: true,
				Children: []Node{
					{
						Path: "/cmd/server", Name: "server", IsDir: true, Expanded: true,
						Children: []Node{
							{Path: "/cmd/server/main.go", Name: "main.go"},
						},
					},
				},
			},
			{
				Path: "/internal", Name: "internal", IsDir: true,
				Children: []Node{
					{
						Path: "/internal/store", Name: "store", IsDir: true,
						Children: []Node{
							{
								Path: "/internal/store/sql", Name: "sql", IsDir: true,
								Children: []Node{
									{Path: "/internal/store/sql/query.go", Name: "query.go"},
									{Path: "/internal/store/sql/migrate.go", Name: "migrate.go"},
								},
							},
						},
					},
					{Path: "/internal/auth.go", Name: "auth.go"},
				},
			},
			{Path: "/README.md", Name: "README.md"},
		},
	}
}
