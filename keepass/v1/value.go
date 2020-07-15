package keepassv1
// TODO re-evaluate having this on its own after getting a better understanding of v2

type Value struct {
	// field name
	name	string
	// arbitrary datas
	value []byte
}

func (v Value) Name() string {
	return v.name
}

func (v Value) Value() []byte {
	return v.value
}
