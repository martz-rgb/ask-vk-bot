package main

type StateStack []StateNode

func (stack *StateStack) Len() int {
	return len(*stack)
}

func (stack *StateStack) Push(new StateNode) {
	*stack = append(*stack, new)
}

func (stack *StateStack) Pop() StateNode {
	state := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return state
}

func (stack *StateStack) Peek() StateNode {
	return (*stack)[len(*stack)-1]
}
