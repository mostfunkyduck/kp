package keepassv1
// TODO re-evaluate having this on its own after getting a better understanding of v2

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
