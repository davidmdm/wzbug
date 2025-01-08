package wasi

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"reflect"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type ExecParams struct {
	CompiledModule wazero.CompiledModule
	Name           string
	CacheDir       string
}

func Execute(ctx context.Context, params ExecParams) (output []byte, err error) {
	cfg := wazero.
		NewRuntimeConfig().
		WithCloseOnContextDone(true)

	if params.CacheDir != "" {
		cache, err := wazero.NewCompilationCacheWithDir(params.CacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate compilation cache: %w", err)
		}
		cfg = cfg.WithCompilationCache(cache)
	}

	runtime := wazero.NewRuntimeWithConfig(ctx, cfg)
	defer runtime.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	moduleCfg := wazero.
		NewModuleConfig().
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithRandSource(rand.Reader).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithArgs(params.Name)

	module, err := runtime.InstantiateModule(ctx, params.CompiledModule, moduleCfg)
	defer func() {
		if reflect.ValueOf(module).IsNil() {
			return
		}
		module.Close(ctx)
	}()

	if err != nil {
		details := stderr.String()
		if details == "" {
			details = "(no output captured on stderr)"
		}
		return nil, fmt.Errorf("failed to instantiate module: %w: stderr: %s", err, details)
	}

	fmt.Println(stdout.String())

	return stdout.Bytes(), nil
}

type CompileParams struct {
	Wasm     []byte
	CacheDir string
}

func Compile(ctx context.Context, params CompileParams) (mod wazero.CompiledModule, err error) {
	cfg := wazero.
		NewRuntimeConfig().
		WithCloseOnContextDone(true)

	if params.CacheDir != "" {
		cache, err := wazero.NewCompilationCacheWithDir(params.CacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate compilation cache: %w", err)
		}
		cfg = cfg.WithCompilationCache(cache)
	}

	runtime := wazero.NewRuntimeWithConfig(ctx, cfg)
	defer runtime.Close(ctx)

	// TODO: check if this is needed for compilation? If not remove.
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	return runtime.CompileModule(ctx, params.Wasm)
}
