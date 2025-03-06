package main

import (
	"context"
	"iter"
)

type paginatable[O, R any] interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(O)) (R, error)
}

func paginatorToSeq[O, R, V any](
	ctx context.Context, p paginatable[O, R], f func(R) iter.Seq[V], optFns ...func(O),
) iter.Seq2[V, error] {
	var empty V
	return func(yield func(V, error) bool) {
		for p.HasMorePages() {
			r, err := p.NextPage(ctx, optFns...)
			if err != nil {
				yield(empty, err)
				return
			}
			for v := range f(r) {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

// from https://github.com/golang/go/issues/61898
func filter2[K, V any](f func(K, V) bool, seq iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if f(k, v) && !yield(k, v) {
				return
			}
		}
	}
}

// from https://github.com/golang/go/issues/61898
func concat[V any](seqs ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, seq := range seqs {
			for e := range seq {
				if !yield(e) {
					return
				}
			}
		}
	}
}
