# Hair Raid

Hair Raid is a small but complete 3D corridor runner that sits between Crash Bandicoot and Subway Surfers. You play Downald D, a proboscis monkey with tremendous hair, auto-running through an auto-generated corridor world. Chasing him is Copycat, the cat barber, who wants that hair for himself.

## Requirements

- Go 1.22 or newer
- A C compiler, because raylib links through cgo
- A desktop operating system (Windows, macOS, or Linux)

## Build and run

Build everything and run the game:

```
go build ./...
go run ./cmd/copycat
```

Or build a binary and run it:

```
go build -o copycat ./cmd/copycat
./copycat
```

Run the headless tests:

```
go test ./...
```

## Controls

### Keyboard

- Arrow keys — switch lanes left and right
- ArrowUp or Space — jump
- ArrowDown — slide; press it in the air to slam down
- P — pause during a run
- Q — quit

In menus:

- Arrow keys — move the selection
- Enter — confirm
- Esc or Backspace — go back

### Gamepad

- Left stick or dpad — move
- A — jump and confirm
- B — slide and go back
- Start — pause

## What's in the game

### Menu

The main menu offers Continue, New Game, Map, Endless Run, Settings and Quit (exit).

- Continue resumes at the highest unlocked level.
- New Game starts the campaign over from level one with fresh progress.
- Map lets you pick any unlocked level, including the boss level.
- Endless Run starts a fresh, non-campaign run on a new seed.
- Settings opens music volume, sound volume, fullscreen and the language choice.
- Quit leaves the game.

### Settings

Settings cover music volume, sound volume, fullscreen, and the language choice: English or Ukrainian. The language is chosen at first run and can be changed in settings at any time.

### Campaign

The campaign holds three worlds of three levels each, followed by the boss level The Last Haircut, where you face Copycat. Every campaign level is generated from a fixed seed, so each attempt replays the same learnable corridor.

### Endless run

The endless run records your best distance. It draws a fresh seed each time, so no two runs are the same.

### Pickups and hazards

- Combs — collect 100 to gain a life.
- Hairspray — a one-hit shield.
- Mirrors — checkpoints.
- Stumble twice while Copycat chases and he snips a lock and takes a life.

### Sound and text

All music and sound effects are synthesized chip-tune, generated in code. The pixel font is drawn in code and covers Latin and Ukrainian.

### Saving

Progress and settings persist in the user config directory. If the save is corrupt, it is quarantined aside and the game starts fresh instead of crashing.

### Assets

Every visual and sound is generated in code, so no asset files are needed — and none ship with the game.
