package portal

import "connectrpc.com/connect"

func response[T any](message *T) *connect.Response[T] {
	return connect.NewResponse(message)
}
