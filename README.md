# copy-cat

**Остання зміна** is a tiny 3D corridor platformer in the spirit of the
classic "run down the hallway" chase. You play Norbert, a permanently tired
opossum courier for the "Остання зміна" afterlife delivery service. He
misplaced the main contract, so now he must sprint through a bureaucratic
afterlife office to his last client before the end of his shift — otherwise
he faces eternal overtime. The tone is darkly comic and decent: no blood, no
cruelty, just paperwork and a talent for dramatic fainting.

The game is written in Go on top of [raylib](https://www.raylib.com/) through
[raylib-go](https://github.com/gen2brain/raylib-go) (cgo). Everything is
drawn with raylib primitives and all sound will be synthesized in code, so the
build stays small and contains no binary assets.

## Prerequisites

- Go 1.22 or newer with cgo enabled.
- A C compiler: gcc or clang on Linux, MinGW-w64 on Windows.
- Linux: X11 and OpenGL development headers. On Debian/Ubuntu that is
  typically `libx11-dev`, `libxrandr-dev`, `libxinerama-dev`,
  `libxcursor-dev`, `libxi-dev`, `libgl1-mesa-dev` and `libasound2-dev`
  (ALSA, for audio).
- Windows: MSYS2 with MinGW-w64.

## Build and run

```sh
go build ./cmd/copycat
./copycat
```

On Windows the built binary is `copycat.exe`.

## Controls

- `A` / `D` or left / right arrows — switch lanes
- `Space` or `W` — jump
- `X` or `Enter` — attack
- `Esc` — exit

## Roadmap

Desktop is the first target. Android comes after the desktop build is solid.
