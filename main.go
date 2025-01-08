package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/davidmdm/wzbug/wasi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	info, _ := debug.ReadBuildInfo()

	for _, dep := range info.Deps {
		if dep.Path != "github.com/tetratelabs/wazero" {
			continue
		}
		fmt.Println("Using:", dep.Path, dep.Version)
		break
	}

	build := exec.Command("go", "build", "-o", "example.wasm", "./example")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to compile example to wasm: %s: %s", err, out)
	}

	wasm, err := os.ReadFile("./example.wasm")
	if err != nil {
		return fmt.Errorf("failed to read wasm file: %w", err)
	}

	mod, err := wasi.Compile(context.Background(), wasi.CompileParams{
		Wasm:     wasm,
		CacheDir: "./cache",
	})
	if err != nil {
		return fmt.Errorf("failed to compile module: %w", err)
	}

	_, err = wasi.Execute(context.Background(), wasi.ExecParams{
		CompiledModule: mod,
		Name:           "foo",
		CacheDir:       "./cache",
	})
	if err != nil {
		return fmt.Errorf("failed to execute wasm: %w", err)
	}

	return nil
}
