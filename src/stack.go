package main

type Stack[T interface{}] []T

func (stack *Stack[T]) Len() int {
	return len(*stack)
}

func (stack *Stack[T]) Push(new T) {
	*stack = append(*stack, new)
}

func (stack *Stack[T]) Pop() T {
	state := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return state
}

func (stack *Stack[T]) Peek() T {
	return (*stack)[len(*stack)-1]
}
