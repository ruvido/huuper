package main

import (
	"hash/fnv"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
)

func startFrontendWatcher(skeletonDir string, publicDir string, hotReload bool, onRebuild func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := addSkeletonWatches(watcher, skeletonDir); err != nil {
		_ = watcher.Close()
		return err
	}

	go func() {
		defer watcher.Close()

		lastFingerprint, _ := frontendFingerprint(skeletonDir)
		debounce := time.NewTimer(time.Hour)
		debounce.Stop()
		pending := false

		triggerBuild := func() {
			currentFingerprint, err := frontendFingerprint(skeletonDir)
			if err != nil {
				log.Printf("frontend fingerprint failed: %v", err)
				return
			}
			if currentFingerprint == lastFingerprint {
				return
			}
			if err := buildFrontend(skeletonDir, publicDir, hotReload); err != nil {
				log.Printf("frontend rebuild failed: %v", err)
				return
			}
			lastFingerprint = currentFingerprint
			log.Printf("frontend rebuilt: %s -> %s", skeletonDir, publicDir)
			if onRebuild != nil {
				onRebuild()
			}
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
						_ = addSkeletonWatches(watcher, event.Name)
					}
				}

				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Chmod) != 0 {
					pending = true
					if !debounce.Stop() {
						select {
						case <-debounce.C:
						default:
						}
					}
					debounce.Reset(200 * time.Millisecond)
				}
			case <-debounce.C:
				if pending {
					pending = false
					triggerBuild()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("frontend watcher error: %v", err)
			}
		}
	}()

	return nil
}

func addSkeletonWatches(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		return watcher.Add(path)
	})
}

func frontendFingerprint(skeletonDir string) (uint64, error) {
	var entries []string

	err := filepath.WalkDir(skeletonDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(skeletonDir, path)
		if err != nil {
			return err
		}
		if shouldSkipFrontendSource(rel) {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".css" && ext != ".js" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		entries = append(entries, rel+"|"+strconv.FormatInt(info.Size(), 10)+"|"+strconv.FormatInt(info.ModTime().UnixNano(), 10))
		return nil
	})
	if err != nil {
		return 0, err
	}

	sort.Strings(entries)
	h := fnv.New64a()
	for _, line := range entries {
		_, _ = h.Write([]byte(line))
		_, _ = h.Write([]byte{'\n'})
	}
	return h.Sum64(), nil
}
