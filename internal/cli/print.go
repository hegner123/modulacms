package cli

import "github.com/charmbracelet/lipgloss"

type Stack struct {
	Data []TreeNode
}

func (s *Stack) Push(node TreeNode) {
	s.Data = append(s.Data, node)
}
func (s *Stack) Pop() TreeNode {
	toPop := len(s.Data) - 1
	out := s.Data[toPop]
	s.Data = s.Data[:toPop]
	return out
}

func (s *Stack) RecurseStack(tn TreeNode, depth int) {
	if tn.Nodes == nil {
		return
	}
	index := len(*tn.Nodes) - 1

	for i := index; i > -1; i-- {
		list := *tn.Nodes
		if list[i].Nodes != nil {
			d := depth + 1
			s.RecurseStack(*list[i], d)
		} else {
			toPush := *list[i]
			toPush.Indent = depth
			s.Push(toPush)
		}

	}
}

func LoopThrough(m Model) string {
	display := Stack{}
	display.RecurseStack(*m.Root.Root, 0)
	out := make([]string, 0, len(display.Data))
	for len(display.Data) > 0 {
		next := display.Pop()
		row := FormatRow(&next)
		out = append(out, row)

	}
	return lipgloss.JoinVertical(lipgloss.Top, out...)

}
