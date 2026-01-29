package main

import (
	"context"
	"log"
	"os"

	"github.com/duber000/kukicha/internal/lsp"
)

func main() {
	// Set up logging to stderr (stdout is used for LSP communication)
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Starting kukicha-lsp server...")

	ctx := context.Background()
	server := lsp.NewServer(os.Stdin, os.Stdout)

	if err := server.Run(ctx); err != nil {
		log.Fatalf("LSP server error: %v", err)
	}
}
