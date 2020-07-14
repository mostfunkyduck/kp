package keepassv1

type Value struct {
	// field name
	name	string
	// arbitrary datas
	value interface{}
}

func (v Value) Name() string {
	return v.name
}

func (v Value) Value() interface{} {
	return v.value
}
