// Package save persists the game's progress and settings as a single
// versioned JSON file in the user config directory. A save is written
// atomically through a temporary file and rename, and a file that is
// missing, corrupt, from another version, or out of range starts a fresh
// game while the bad file is quarantined beside it.
package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"copycat/internal/campaign"
)

// fileVersion is the only save format version this game writes and reads.
const fileVersion = 1

// maxUnlocked is the highest level ID the fixed campaign can reach: the nine
// run levels plus the Copycat boss finale.
const maxUnlocked = 10

// File is one persisted save: the progress and settings, wrapped in a
// version so later builds can migrate or quarantine old files.
type File struct {
	Version  int               `json:"version"`
	Progress campaign.Progress `json:"progress"`
	Settings campaign.Settings `json:"settings"`
}

// Defaults returns the save a fresh install starts with: version 1, fresh
// progress and default settings.
func Defaults() File {
	return File{
		Version:  fileVersion,
		Progress: campaign.NewProgress(),
		Settings: campaign.DefaultSettings(),
	}
}

// DefaultPath returns the path of the save file in the user config
// directory: <user config>/copycat/save.json.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "copycat", "save.json"), nil
}

// Load reads and validates the save file at path. A missing file returns
// Defaults without error. A file that fails to parse, carries a version
// other than 1, or holds out-of-range values is quarantined beside path and
// Defaults is returned. Any other read or quarantine error is returned.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Defaults(), nil
		}
		return File{}, err
	}

	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return quarantine(path)
	}
	if err := f.validate(); err != nil {
		return quarantine(path)
	}
	return f, nil
}

// Write serializes f and writes it to path atomically: the JSON goes to a
// temporary file in the same directory first, and only then is it renamed
// over path. A crash can never leave a half-written save behind.
func (f File) Write(path string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".copycat-save-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	keep := false
	defer func() {
		_ = tmp.Close()
		if !keep {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	keep = true
	return nil
}

// validate reports the first reason f must not be used.
func (f File) validate() error {
	if f.Version != fileVersion {
		return fmt.Errorf("unsupported save version %d", f.Version)
	}
	if f.Progress.Unlocked < 0 || f.Progress.Unlocked > maxUnlocked {
		return fmt.Errorf("unlocked level out of range: %d", f.Progress.Unlocked)
	}
	if f.Progress.BestEndless < 0 {
		return fmt.Errorf("endless best distance out of range: %v", f.Progress.BestEndless)
	}
	if f.Settings.MusicVolume < 0 || f.Settings.MusicVolume > 1 {
		return fmt.Errorf("music volume out of range: %v", f.Settings.MusicVolume)
	}
	if f.Settings.SoundVolume < 0 || f.Settings.SoundVolume > 1 {
		return fmt.Errorf("sound volume out of range: %v", f.Settings.SoundVolume)
	}
	switch f.Settings.Language {
	case campaign.LanguageEnglish, campaign.LanguageUkrainian:
	default:
		return fmt.Errorf("unknown language %d", f.Settings.Language)
	}
	return nil
}

// quarantine renames the bad file at path aside, keeping its bytes, and
// returns Defaults so the game starts fresh. It picks the first free
// "<path>.corrupt[-n]" name so an earlier quarantine is never overwritten.
func quarantine(path string) (File, error) {
	aside, err := quarantinePath(path)
	if err != nil {
		return File{}, err
	}
	if err := os.Rename(path, aside); err != nil {
		return File{}, err
	}
	return Defaults(), nil
}

// quarantinePath returns the first free quarantine name beside path.
func quarantinePath(path string) (string, error) {
	base := path + ".corrupt"
	for i := 0; ; i++ {
		name := base
		if i > 0 {
			name = base + "-" + strconv.Itoa(i)
		}
		_, err := os.Stat(name)
		if errors.Is(err, fs.ErrNotExist) {
			return name, nil
		}
		if err != nil {
			return "", err
		}
	}
}
