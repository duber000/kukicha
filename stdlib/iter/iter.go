// Kukicha Standard Library - Iterator Operations
// Written in Go until Kukicha gets function type syntax
// This demonstrates the special transpilation approach

package iter

import "iter"

// Filter returns an iterator that yields only items matching the predicate
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) bool {
		for item := range seq {
			if keep(item) {
				if !yield(item) {
					return false
				}
			}
		}
		return true
	}
}

// Map transforms each item in the iterator
func Map[T, U any](seq iter.Seq[T], transform func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) bool {
		for item := range seq {
			if !yield(transform(item)) {
				return false
			}
		}
		return true
	}
}

// Take returns an iterator of the first n items
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) bool {
		count := 0
		for item := range seq {
			if count >= n {
				return true
			}
			if !yield(item) {
				return false
			}
			count++
		}
		return true
	}
}

// Skip returns an iterator that skips the first n items
func Skip[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) bool {
		count := 0
		for item := range seq {
			if count >= n {
				if !yield(item) {
					return false
				}
			}
			count++
		}
		return true
	}
}
