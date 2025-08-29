package cmd

type Result[T any] struct {
	Value T
	Err   *XbeeError
}

// Ok crée un Result porteur d'une valeur valide.
func Ok[T any](v T) Result[T] {
	return Result[T]{Value: v, Err: nil}
}

// Err crée un Result porteur d'une erreur.
func Err[T any](e *XbeeError) Result[T] {
	// la valeur Zero[T] reste inutilisée
	var zero T
	return Result[T]{Value: zero, Err: e}
}
