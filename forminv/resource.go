package forminv

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/df-mc/dragonfly/server"
	gtresource "github.com/sandertv/gophertunnel/minecraft/resource"
)

//go:embed resource/pack/** resource/pack/ui/_ui_defs.json
var packFS embed.FS

// Pack builds and returns the bundled resource pack required for forminv menus.
func Pack() (*gtresource.Pack, error) {
	buf := bytes.NewBuffer(nil)
	zw := zip.NewWriter(buf)

	var files []string
	if err := fs.WalkDir(packFS, "resource/pack", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk embedded pack: %w", err)
	}
	sort.Strings(files)

	for _, path := range files {
		data, err := packFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read embedded pack file %s: %w", path, err)
		}
		name := strings.TrimPrefix(filepath.ToSlash(path), "resource/pack/")
		w, err := zw.Create(name)
		if err != nil {
			return nil, fmt.Errorf("create zip entry %s: %w", name, err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, fmt.Errorf("write zip entry %s: %w", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close resource pack zip: %w", err)
	}

	pack, err := gtresource.Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("read resource pack: %w", err)
	}
	return pack, nil
}

// MustPack builds and returns the bundled resource pack, panicking if the pack cannot be built.
func MustPack() *gtresource.Pack {
	pack, err := Pack()
	if err != nil {
		panic(err)
	}
	return pack
}

// AddToConfig builds the bundled resource pack and appends it to a Dragonfly server config.
func AddToConfig(conf *server.Config) error {
	pack, err := Pack()
	if err != nil {
		return err
	}
	conf.Resources = append(conf.Resources, pack)
	return nil
}

// MustAddToConfig builds the bundled resource pack and appends it to a Dragonfly server config, panicking on
// failure.
func MustAddToConfig(conf *server.Config) {
	if err := AddToConfig(conf); err != nil {
		panic(err)
	}
}
