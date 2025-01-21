package utils

type ProtoFromModel[M any, P any] interface {
	FromModel(model M) *P
	*P
}

// ProtoFromOptionalModel converts a model to a proto message. The proto Message in question must
// implement the ProtoFromModel interface which essentially while complicated to read above is
// simply a method `FromModel` that takes any structure and returns a filled protobuf message.
//
// Usage is as follows:
//
//	var out = utils.ProtoFromOptionalModel[pbquote.Address](address)
//
// Where out here will be a pointer to a pbquote.Address if address is not nil, otherwise it will be nil.
// It is essentially a helper function to avoid boilerplate code that gives:
//
//	if address == nil {
//	    return nil
//	}
//
//	return new(pbquote.Address).FromModel(address)
func ProtoFromOptionalModel[T any, M comparable, PT ProtoFromModel[M, T]](model M) *T {
	var mZero M
	if model == mZero {
		return nil
	}

	var out T
	p := PT(&out)

	return p.FromModel(model)
}
