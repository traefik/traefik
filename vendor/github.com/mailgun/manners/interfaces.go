package manners

type waitGroup interface {
	Add(int)
	Done()
	Wait()
}
