package list_test

import (
	"cmd/go/internal/cfg"
	"cmd/go/internal/modload"
	"context"
	"internal/testenv"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var modulesForTest = []struct {
	name            string
	path            string
	expectedModules int8
}{
	{
		name:            "Empty",
		path:            "empty",
		expectedModules: 0,
	},
	{
		name:            "Cmd",
		path:            "cmd",
		expectedModules: 0,
	},
	{
		name:            "K8S",
		path:            "strippedk8s",
		expectedModules: 0,
	},
}

var TEST_NUMBER = 0

func BenchmarkListModules(b *testing.B) {
	testenv.MustHaveExternalNetwork(b)
	testDataDir := b.TempDir()
	for _, m := range modulesForTest {
		moduleTempDir := filepath.Join(testDataDir, m.path)
		err := os.Mkdir(moduleTempDir, 0755)
		if err != nil {
			b.Errorf("Failed to create a testdata directory: %v", err)
			return
		}
		err = CopyFileToDirectory(filepath.Join("testdata", m.path, "go.mod"), moduleTempDir)
		if err != nil {
			b.Errorf("Failed to copy go.mod: %v", err)
			return
		}
		goSum := filepath.Join("testdata", m.path, "go.sum")
		if _, err = os.Stat(goSum); err == nil {
			err = CopyFileToDirectory(goSum, moduleTempDir)
			if err != nil {
				b.Errorf("Failed to copy go.sum: %v", err)
				return
			}
		}
		b.Run(m.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				TEST_NUMBER++
				b.StopTimer()
				gopath := b.TempDir()
				os.Setenv("GOPATH", gopath)
				cfg.BuildContext.GOPATH = gopath
				cfg.GOMODCACHE = filepath.Join(gopath, "pkg/mod")
				cfg.SumdbDir = filepath.Join(gopath, "pkg/sumdb")
				modload.Reset()
				ctx := context.Background()
				modload.ForceUseModules = true
				modload.RootMode = modload.NeedRoot
				cfg.ModulesEnabled = true
				cfg.ModFile = filepath.Join(moduleTempDir, "go.mod")
				cfg.BuildMod = "readonly"
				cfg.BuildModExplicit = true
				cfg.BuildModReason = "to avoid vendoring error"
				b.StartTimer()
				//file, _ := os.Create(fmt.Sprintf("%d.pprof", TEST_NUMBER))
				//pprof.StartCPUProfile(file)
				modload.Init()
				got, err := modload.ListModules(ctx, []string{"all"}, 0, "")
				//pprof.StopCPUProfile()
				//file.Close()
				if err != nil {
					b.Errorf("ListModules() error = %v", err)
					return
				}
				if got != nil {
				}
			}
		})
	}
}

func CopyFileToDirectory(src, destDir string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	_, filename := filepath.Split(src)
	destPath := filepath.Join(destDir, filename)

	destinationFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
