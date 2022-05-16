package option

type Option interface {
	Ident() interface{}
	Value() interface{}
}

type option struct {
	ident interface{}
	value interface{}
}

func New(ident, value interface{}) Option {
	return &option{ident, value}
}

func (o *option) Ident() interface{} {
	return o.ident
}

func (o *option) Value() interface{} {
	return o.value
}
