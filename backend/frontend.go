package main

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultSkeletonDir = "frontend/skeleton"
	defaultPublicDir   = "frontend/site"
)

func buildFrontend(skeletonDir string, publicDir string, hotReload bool) error {
	if err := os.RemoveAll(publicDir); err != nil {
		return err
	}
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		return err
	}

	if err := copyDir(filepath.Join(skeletonDir, "assets"), filepath.Join(publicDir, "assets")); err != nil && !os.IsNotExist(err) {
		return err
	}

	componentFiles, err := frontendComponentFiles(filepath.Join(skeletonDir, "components"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return filepath.WalkDir(skeletonDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(skeletonDir, path)
		if err != nil {
			return err
		}

		if shouldSkipFrontendSource(relPath) {
			return nil
		}
		if filepath.Ext(relPath) != ".html" {
			return copyFile(path, filepath.Join(publicDir, relPath))
		}

		outPath := filepath.Join(publicDir, relPath)
		return renderHTMLFile(path, outPath, componentFiles, hotReload)
	})
}

func shouldSkipFrontendSource(relPath string) bool {
	base := filepath.Base(relPath)
	if base == "README.md" || base == ".gitkeep" {
		return true
	}
	if strings.HasPrefix(relPath, "components"+string(filepath.Separator)) {
		return true
	}
	return false
}

func frontendComponentFiles(componentsDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(componentsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		files = append(files, path)
		return nil
	})

	return files, err
}

func renderHTMLFile(srcPath string, dstPath string, componentFiles []string, hotReload bool) error {
	data, err := renderTemplate(srcPath, componentFiles)
	if err != nil {
		return err
	}

	if hotReload {
		data = injectHotReload(data)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(dstPath, data, 0o644)
}

func renderTemplate(pagePath string, componentFiles []string) ([]byte, error) {
	funcs := template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, io.ErrUnexpectedEOF
			}

			out := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, io.ErrUnexpectedEOF
				}
				out[key] = values[i+1]
			}

			return out, nil
		},
		"list": func(values ...any) []any {
			return values
		},
	}

	tpl := template.New("page").Funcs(funcs)

	for _, componentPath := range componentFiles {
		componentData, err := os.ReadFile(componentPath)
		if err != nil {
			return nil, err
		}
		if _, err := tpl.Parse(string(componentData)); err != nil {
			return nil, err
		}
	}

	pageData, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, err
	}
	if _, err := tpl.Parse(string(pageData)); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := tpl.ExecuteTemplate(&out, "page", nil); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func injectHotReload(data []byte) []byte {
	snippet := []byte(`  <script src="/assets/js/core/hot-reload.js" defer></script>` + "\n")
	lower := strings.ToLower(string(data))
	idx := strings.LastIndex(lower, "</body>")
	if idx < 0 {
		var out bytes.Buffer
		out.Write(data)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			out.WriteByte('\n')
		}
		out.Write(snippet)
		return out.Bytes()
	}

	var out bytes.Buffer
	out.Write(data[:idx])
	if idx > 0 && data[idx-1] != '\n' {
		out.WriteByte('\n')
	}
	out.Write(snippet)
	out.Write(data[idx:])
	return out.Bytes()
}

func copyDir(srcDir string, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(srcPath string, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
