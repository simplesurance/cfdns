package cfdns

import (
	"context"
	"errors"
)

var done = errors.New("done")

// Iterator implements an iterator algorithm from a function that fetches
// blocks of data. This allows having fixed-memory usage when reading
// arbitrary-sized structs without leaking implementation details about
// how to paginate consecutive blocks of data.
type Iterator[T any] struct {
	fetchNext         BlockFetchFn[T]
	elements          []T
	lastContinueToken any
}

type BlockFetchFn[T any] func(ctx context.Context, continueToken any) (
	[]T,
	any,
	error)

func (it *Iterator[T]) next(ctx context.Context) (retElm T, err error) {
	if len(it.elements) == 0 {
		var elements []T
		var nextContinueToken any
		elements, nextContinueToken, err = it.fetchNext(ctx, it.lastContinueToken)
		if err != nil {
			return
		}

		it.elements = elements
		it.lastContinueToken = nextContinueToken
	}

	retElm = it.elements[0]
	it.elements = it.elements[1:]
	return retElm, nil
}
