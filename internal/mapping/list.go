package mapping

func List[IN any, OUT any](in []IN, mapping func(in IN) OUT) []OUT {
	out := make([]OUT, len(in))
	for i, el := range in {
		out[i] = mapping(el)
	}
	return out
}
