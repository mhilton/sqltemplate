package sqltemplate

import "text/template/parse"

// escapeTree adds additional "sqlliteral" function calls to the end of all
// pipelines. This ensures that inserted variables are formatted as
// appropriate SQL literals. This function is idempotent so an "sqlliteral"
// function call is only added to the end of pipelines where there isn't
// one already.
func escapeTree(t *parse.Tree) *parse.Tree {
	if t.Root == nil {
		return t
	}
	escapeNode(t, t.Root)
	return t
}

// escapeNode processes the given node in the given tree for adding
// sqlliteral function calls to the end of pipelines.
func escapeNode(t *parse.Tree, n parse.Node) {
	switch v := n.(type) {
	case *parse.ActionNode:
		escapeNode(t, v.Pipe)
	case *parse.IfNode:
		escapeNode(t, v.List)
		escapeNode(t, v.ElseList)
	case *parse.ListNode:
		if v == nil {
			return
		}
		for _, n := range v.Nodes {
			escapeNode(t, n)
		}
	case *parse.PipeNode:
		if len(v.Decl) > 0 {
			// If the pipe sets variables then don't escape it.
			return
		}
		if len(v.Cmds) < 1 {
			return
		}
		cmd := v.Cmds[len(v.Cmds)-1]
		if len(cmd.Args) == 1 && cmd.Args[0].Type() == parse.NodeIdentifier && cmd.Args[0].(*parse.IdentifierNode).Ident == "sqlliteral" {
			return
		}
		v.Cmds = append(v.Cmds, &parse.CommandNode{
			NodeType: parse.NodeCommand,
			Args:     []parse.Node{parse.NewIdentifier("sqlliteral").SetTree(t).SetPos(cmd.Pos)},
		})
	case *parse.RangeNode:
		escapeNode(t, v.List)
		escapeNode(t, v.ElseList)
	case *parse.WithNode:
		escapeNode(t, v.List)
		escapeNode(t, v.ElseList)
	}
}
