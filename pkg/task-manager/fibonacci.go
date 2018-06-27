package task_manager

// FibonacciFunc returns successive Fibonacci numbers starting from 1
func FibonacciFunc() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}

func FibonacciIndex(index int) int {
	fib := FibonacciFunc()
	num := fib()

	for i := 0; i < index; i++ {
		num = fib()
	}
	return num
}

