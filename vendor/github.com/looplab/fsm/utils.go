package fsm

import (
	"bytes"
	"fmt"
)

// Visualize outputs a visualization of a FSM in Graphviz format.
func Visualize(fsm *FSM) string {
	var buf bytes.Buffer

	states := make(map[string]int)

	buf.WriteString(fmt.Sprintf(`digraph fsm {`))
	buf.WriteString("\n")

	// make sure the initial state is at top
	for k, v := range fsm.transitions {
		if k.src == fsm.current {
			states[k.src]++
			states[v]++
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	for k, v := range fsm.transitions {
		if k.src != fsm.current {
			states[k.src]++
			states[v]++
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\n")

	for k := range states {
		buf.WriteString(fmt.Sprintf(`    "%s";`, k))
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintln("}"))

	return buf.String()
}
