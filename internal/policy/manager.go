package policy

type Manager interface {
	Evaluate(query string, input interface{}) bool
}
