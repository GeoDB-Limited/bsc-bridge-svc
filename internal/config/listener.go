package config

type Listener interface {
	Address() string
}

type listener struct {
	Addr string `yaml:"addr"`
}

func (l *listener) Address() string {
	return l.Addr
}
