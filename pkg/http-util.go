package pkg

type Response[T any] struct {
	Data  T      `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func Success[T any](data T) Response[T] {
	return Response[T]{Data: data}
}

func Fail[T any](msg string) Response[T] {
	return Response[T]{Error: msg}
}