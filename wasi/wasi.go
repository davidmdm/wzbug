package wasi

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/davidmdm/x/xerr"
)

type ExecParams struct {
	Name     string
	Wasm     []byte
	Stdin    io.Reader
	CacheDir string
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
	defer func() {
		err = xerr.MultiErrFrom("", err, runtime.Close(ctx))
	}()

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	moduleCfg := wazero.
		NewModuleConfig().
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithStdin(params.Stdin).
		WithRandSource(rand.Reader).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithArgs(params.Name)

	compiledModule, err := runtime.CompileModule(ctx, params.Wasm)
	if err != nil {
		return nil, fmt.Errorf("failed to compile module: %w", err)
	}

	fmt.Println("[debug] module compiled")
	fmt.Println("[debug] instantiating module...")

	if _, err := runtime.InstantiateModule(ctx, compiledModule, moduleCfg); err != nil {
		details := stderr.String()
		if details == "" {
			details = "(no output captured on stderr)"
		}
		return nil, fmt.Errorf("failed to instantiate module: %w: stderr: %s", err, details)
	}

	return stdout.Bytes(), nil
}
