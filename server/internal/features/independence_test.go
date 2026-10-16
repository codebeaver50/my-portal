// Package features は feature 間の依存を検査するテストだけを置く。各featureの実装は
// サブパッケージ（internal/features/<feature>/...）に置く。
package features

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const featuresImportPrefix = "github.com/codebeaver50/my-portal/server/internal/features/"

// TestFeaturesDoNotImportEachOther は、feature から別の feature のパッケージを import して
// いないことを確かめる。internal/features/ 配下のディレクトリを読んで検査するため、featureを
// 追加しても設定なしで対象になる（golangci-lint の depguard ではfeatureごとにルールを
// 書き並べる必要があり、追記を忘れると黙って検査の対象から外れてしまうため、テストにしている）。
func TestFeaturesDoNotImportEachOther(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read features dir: %v", err)
	}

	var features []string
	for _, entry := range entries {
		if entry.IsDir() {
			features = append(features, entry.Name())
		}
	}
	if len(features) == 0 {
		t.Fatal("no features found")
	}

	for _, feature := range features {
		err := filepath.WalkDir(feature, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}

			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, spec := range file.Imports {
				importPath, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return err
				}
				imported, ok := strings.CutPrefix(importPath, featuresImportPrefix)
				if !ok {
					continue
				}
				if importedFeature, _, _ := strings.Cut(imported, "/"); importedFeature != feature {
					t.Errorf("%s: feature %q imports feature %q (%s)", path, feature, importedFeature, importPath)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("failed to walk feature %q: %v", feature, err)
		}
	}
}
