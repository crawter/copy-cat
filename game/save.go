package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Save owns progress persistence. It loads totals from the user config
// directory once and folds the current run into them on exit.
type Save struct {
	Stamps       int     `json:"stamps"`
	Deaths       int     `json:"deaths"`
	BestDistance float32 `json:"best_distance"`
	Completed    bool    `json:"completed"`

	applied bool
}

// NewSave loads persisted progress from the user config directory. A missing,
// unreadable, or corrupt file means a first launch, so all values stay zero.
func NewSave() *Save {
	s := &Save{}
	s.load()
	return s
}

// load reads save.json into Save's own fields and leaves g.Run alone.
func (s *Save) load() {
	path, err := savePath()
	if err != nil {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, s); err != nil {
		s.Stamps = 0
		s.Deaths = 0
		s.BestDistance = 0
		s.Completed = false
	}
}

// Update applies the loaded totals to the run once, then does nothing more.
func (s *Save) Update(dt float32, g *Game) {
	if s.applied {
		return
	}
	s.applied = true

	g.Run.TotalStamps = s.Stamps
	g.Run.Deaths = s.Deaths
	g.Run.BestDistance = s.BestDistance
	g.Run.Completed = s.Completed
}

// Draw renders nothing; the save system has no on-screen state.
func (s *Save) Draw(g *Game) {}

// Write folds the current run into the stored totals and persists them
// atomically in the user config directory.
func (s *Save) Write(g *Game) {
	if g.Run.Distance > s.BestDistance {
		s.BestDistance = g.Run.Distance
	}
	if g.Run.Completed {
		s.Completed = true
	}

	s.Stamps = g.Run.TotalStamps
	s.Deaths = g.Run.Deaths

	if err := s.persist(); err != nil {
		// Write has a fixed void signature, so a persistence failure cannot be
		// returned; it must not crash an exiting game.
		return
	}
}

// persist writes the save JSON atomically: a temporary file in the same
// directory is renamed over save.json.
func (s *Save) persist() error {
	path, err := savePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "save-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	tmpName = ""

	return nil
}

// savePath returns the per-user save file location:
// os.UserConfigDir()/copy-cat/save.json.
func savePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "copy-cat", "save.json"), nil
}
