package spa


type Spa interface {
	SetEnv(arg string)
	Server()
}

func NewSPA() Spa {
	return &SPA{}
}