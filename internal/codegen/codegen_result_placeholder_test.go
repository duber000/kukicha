package codegen

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/parser"
)

func TestResultPlaceholderSliceMap(t *testing.T) {
	input := `petiole slice

func Map(items list of any, transform func(any) result) list of result
    out := make(list of result, len(items))
    for i, item in items
        out[i] = transform(item)
    return out
`

	p, err := parser.New(input, "stdlib/slice/slice.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/slice/slice.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify generic type parameters include R for result
	if !strings.Contains(output, "func Map[T any, R any]") {
		t.Errorf("expected generic function signature with [T any, R any], got: %s", output)
	}

	// Verify the parameter signature uses T and R
	if !strings.Contains(output, "(items []T, transform func(T) R)") {
		t.Errorf("expected correct parameter types with T and R, got: %s", output)
	}

	// Verify return type uses R
	if !strings.Contains(output, "[]R") {
		t.Errorf("expected return type []R, got: %s", output)
	}
}

func TestResultPlaceholderConcurrentMap(t *testing.T) {
	input := `petiole concurrent

import "sync"

func Map(items list of any, fn func(any) result) list of result
    results := make(list of result, len(items))
    wg := sync.WaitGroup{}
    wg.Add(len(items))
    for i, item in items
        idx := i
        it := item
        go func()
            results[idx] = fn(it)
            wg.Done()
        ()
    wg.Wait()
    return results
`

	p, err := parser.New(input, "stdlib/concurrent/concurrent.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/concurrent/concurrent.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify generic type parameters
	if !strings.Contains(output, "func Map[T any, R any]") {
		t.Errorf("expected generic function signature with [T any, R any], got: %s", output)
	}

	// Verify parameter types
	if !strings.Contains(output, "(items []T, fn func(T) R)") {
		t.Errorf("expected correct parameter types, got: %s", output)
	}

	// Verify return type
	if !strings.Contains(output, ") []R {") {
		t.Errorf("expected return type []R, got: %s", output)
	}

	// Verify body uses R for make
	if !strings.Contains(output, "make([]R, len(items))") {
		t.Errorf("expected make([]R, len(items)) in body, got: %s", output)
	}
}

func TestResultPlaceholderConcurrentMapWithLimit(t *testing.T) {
	input := `petiole concurrent

import "sync"

func MapWithLimit(items list of any, limit int, fn func(any) result) list of result
    results := make(list of result, len(items))
    wg := sync.WaitGroup{}
    semaphore := make(channel of int, limit)
    wg.Add(len(items))
    for i, item in items
        idx := i
        it := item
        send 1 to semaphore
        go func()
            results[idx] = fn(it)
            receive from semaphore
            wg.Done()
        ()
    wg.Wait()
    return results
`

	p, err := parser.New(input, "stdlib/concurrent/concurrent.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/concurrent/concurrent.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify generic type parameters
	if !strings.Contains(output, "func MapWithLimit[T any, R any]") {
		t.Errorf("expected generic function signature with [T any, R any], got: %s", output)
	}

	// Verify parameter types include limit int between T items and R func
	if !strings.Contains(output, "(items []T, limit int, fn func(T) R)") {
		t.Errorf("expected correct parameter types, got: %s", output)
	}

	// Verify return type
	if !strings.Contains(output, ") []R {") {
		t.Errorf("expected return type []R, got: %s", output)
	}
}
