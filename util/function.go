package util

import (
	"context"
	"github.com/iodasolutions/xbee-common/cmd"
)

type mapFunc[E any, F any] func(E) F

func Map[E any, F any](s []E, f mapFunc[E, F]) []F {
	result := make([]F, len(s))
	for i := range s {
		result[i] = f(s[i])
	}
	return result
}

type keepFunc[E any] func(E) bool

func Filter[E any](s []E, f keepFunc[E]) []E {
	result := []E{}
	for _, v := range s {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

type keepFuncWithCtx[E any] func(context.Context, E) (bool, *cmd.XbeeError)

func FilterWithCtx[E any](ctx context.Context, s []E, f keepFuncWithCtx[E]) ([]E, *cmd.XbeeError) {
	result := []E{}
	for _, v := range s {
		ok, err := f(ctx, v)
		if err != nil {
			return nil, err
		}
		if ok {
			result = append(result, v)
		}
	}
	return result, nil
}

type consumeFuncWithCtx[E any] func(context.Context, E) *cmd.XbeeError

func ConsumeWithCtx[E any](ctx context.Context, s []E, f consumeFuncWithCtx[E]) *cmd.XbeeError {
	for _, v := range s {
		if err := f(ctx, v); err != nil {
			return err
		}
	}
	return nil
}

type consumeFunc[E any] func(E) *cmd.XbeeError

func Consume[E any](s []E, f consumeFunc[E]) *cmd.XbeeError {
	for _, v := range s {
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}

type mapFuncWithError[E any, F any] func(E) (F, *cmd.XbeeError)

func MapWithError[E any, F any](s []E, f mapFuncWithError[E, F]) ([]F, *cmd.XbeeError) {
	result := make([]F, len(s))
	for i := range s {
		var err *cmd.XbeeError
		result[i], err = f(s[i])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
