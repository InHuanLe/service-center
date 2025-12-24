package registry

type Options interface {
	Endpoints() []string
}
