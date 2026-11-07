package save_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/save"
)

func savePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "save.json")
}

func writeBytes(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func quarantinedBytes(t *testing.T, dir string) []byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	var matches []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".corrupt") {
			matches = append(matches, filepath.Join(dir, e.Name()))
		}
	}
	if len(matches) != 1 {
		t.Fatalf("quarantine files = %d, want 1", len(matches))
	}
	return readBytes(t, matches[0])
}

func TestWriteLoadRoundTrip(t *testing.T) {
	path := savePath(t)
	want := save.Defaults()
	want.Progress = campaign.Progress{Unlocked: 7, Completed: true, BestEndless: 123.45}
	want.Settings = campaign.Settings{
		MusicVolume: 0.7,
		SoundVolume: 0.3,
		Fullscreen:  true,
		Language:    campaign.LanguageUkrainian,
	}

	if err := want.Write(path); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := save.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	got, err := save.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := save.Defaults()
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoadCorruptQuarantines(t *testing.T) {
	path := savePath(t)
	original := []byte(`{"version":1,"progress":{"Unlocked":`)
	writeBytes(t, path, original)

	got, err := save.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != save.Defaults() {
		t.Fatalf("Load() = %+v, want Defaults", got)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("save file %s still present after quarantine", path)
	}
	if !bytes.Equal(quarantinedBytes(t, filepath.Dir(path)), original) {
		t.Fatalf("quarantined bytes differ from original")
	}
}

func TestLoadWrongVersionQuarantines(t *testing.T) {
	path := savePath(t)
	f := save.Defaults()
	f.Version = 2
	original, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	writeBytes(t, path, original)

	got, err := save.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != save.Defaults() {
		t.Fatalf("Load() = %+v, want Defaults", got)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("save file %s still present after quarantine", path)
	}
	if !bytes.Equal(quarantinedBytes(t, filepath.Dir(path)), original) {
		t.Fatalf("quarantined bytes differ from original")
	}
}

func TestLoadOutOfRangeQuarantines(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*save.File)
	}{
		{"music volume too high", func(f *save.File) { f.Settings.MusicVolume = 1.5 }},
		{"music volume negative", func(f *save.File) { f.Settings.MusicVolume = -0.1 }},
		{"sound volume too high", func(f *save.File) { f.Settings.SoundVolume = 2 }},
		{"sound volume negative", func(f *save.File) { f.Settings.SoundVolume = -0.2 }},
		{"unknown language", func(f *save.File) { f.Settings.Language = campaign.Language(7) }},
		{"unlocked negative", func(f *save.File) { f.Progress.Unlocked = -1 }},
		{"unlocked too high", func(f *save.File) { f.Progress.Unlocked = 11 }},
		{"endless best negative", func(f *save.File) { f.Progress.BestEndless = -5 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := savePath(t)
			f := save.Defaults()
			tt.mut(&f)
			original, err := json.Marshal(f)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			writeBytes(t, path, original)

			got, err := save.Load(path)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got != save.Defaults() {
				t.Fatalf("Load() = %+v, want Defaults", got)
			}
			if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
				t.Fatalf("save file %s still present after quarantine", path)
			}
			if !bytes.Equal(quarantinedBytes(t, filepath.Dir(path)), original) {
				t.Fatalf("quarantined bytes differ from original")
			}
		})
	}
}

func TestDefaultPathEndsWithSaveJSON(t *testing.T) {
	path, err := save.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if filepath.Base(path) != "save.json" {
		t.Fatalf("DefaultPath() base = %q, want save.json", filepath.Base(path))
	}
	if filepath.Base(filepath.Dir(path)) != "copycat" {
		t.Fatalf("DefaultPath() parent = %q, want copycat", filepath.Base(filepath.Dir(path)))
	}
}

func TestWriteReplacesFile(t *testing.T) {
	path := savePath(t)

	first := save.Defaults()
	first.Progress.Unlocked = 2
	first.Settings.Language = campaign.LanguageEnglish

	second := save.Defaults()
	second.Progress.Unlocked = 9
	second.Settings.Language = campaign.LanguageUkrainian

	if err := first.Write(path); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	if err := second.Write(path); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}

	got, err := save.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != second {
		t.Fatalf("Load() = %+v, want %+v", got, second)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "save.json" {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Fatalf("directory entries = %v, want [save.json]", names)
	}
}
