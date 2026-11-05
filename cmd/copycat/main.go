// Command copycat is the POC's entrypoint. It opens the render layer, wires it
// to the logic packages and runs the main loop from the title screen through
// Take 9 to the recap. It holds no game rules of its own: speeds, arcs,
// events, counts, lives and scene flow all stay in their packages.
package main

import (
	"fmt"
	"os"
	"time"

	"copycat/collide"
	"copycat/enemy"
	"copycat/leveldata"
	"copycat/lives"
	"copycat/move"
	"copycat/pickups"
	"copycat/render"
	"copycat/scenestate"
)

const (
	// maxFrameSeconds caps dt each frame so a pause or a slow frame never
	// teleports the hero or the crane forward.
	maxFrameSeconds = 0.05

	// dyingHoldSeconds is how long the Dying state plays the farce before the
	// respawn or game-over event.
	dyingHoldSeconds = 1.0

	// offCorridorZ is a forward position far behind the take, where the
	// defeated crane's body is parked so collide never reports a crane hit.
	offCorridorZ = -1000
)

func main() {
	if err := render.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "copycat:", err)
		return
	}
	defer render.Close()

	take := leveldata.Take9()
	hero := move.NewHero()
	crane := enemy.New(take.Track)
	lives := lives.New()
	reels := pickups.New()
	state := scenestate.Title
	line := ""

	var dyingElapsed float64
	var respawn bool
	var prevStart, prevLeave, prevJump bool

	prevFrame := time.Now()
	for !render.ShouldClose() {
		now := time.Now()
		dt := now.Sub(prevFrame).Seconds()
		prevFrame = now
		if dt > maxFrameSeconds {
			dt = maxFrameSeconds
		}

		left, right, jump, start, leave := render.PollInput()
		startPress := start && !prevStart
		leavePress := leave && !prevLeave
		jumpPress := jump && !prevJump
		prevStart = start
		prevLeave = leave
		prevJump = jump

		switch state {
		case scenestate.Title:
			if startPress {
				hero, crane, lives, reels, line = freshRun(take)
				state = scenestate.Next(state, scenestate.EventStart)
			}

		case scenestate.Playing:
			mapped := render.MapInput(left, right, jumpPress)
			if mapped.Jump {
				render.PlayJump()
			}

			hero = move.Step(hero, mapped, dt)
			crane = enemy.Step(crane, dt)

			if c, downed := enemy.Defeat(hero, crane); downed {
				crane = c
				reels.CollectDropped()
				render.PlayThud()
				render.PlayPickup()
			}

			body := crane.Body()
			if crane.Dropped() {
				body.Pos = leveldata.Point{Z: offCorridorZ}
			}

			var events []collide.Event
			hero, events = collide.Check(hero, take, body)

			died := false
			for _, ev := range events {
				switch ev.Kind {
				case collide.EventRigHit, collide.EventCraneHit:
					lives, respawn = lives.Die()
					state = scenestate.Next(state, scenestate.EventDeath)
					dyingElapsed = 0
					render.PlayThud()
					render.PlayCrackle()
					died = true

				case collide.EventReel:
					if reels.Collect(ev.Index) {
						render.PlayPickup()
						render.PlayCrackle()
						line = scenestate.ReelLine()
					}

				case collide.EventCheckpoint:
					lives = lives.AtCheckpoint(take.Checkpoints[ev.Index].Position)

				case collide.EventExit:
					state = scenestate.Next(state, scenestate.EventExit)
					line = scenestate.Line(scenestate.Recapping)
				}
				if died {
					break
				}
			}

		case scenestate.Dying:
			dyingElapsed += dt
			if dyingElapsed >= dyingHoldSeconds {
				if respawn {
					state = scenestate.Next(state, scenestate.EventRespawn)
					hero = heroAtRespawn(lives.Respawn())
					line = scenestate.RespawnLine(lives.Lives() == 1)
				} else {
					state = scenestate.Next(state, scenestate.EventGameOver)
					line = scenestate.Line(scenestate.GameOver)
				}
			}

		case scenestate.Recapping:
			if startPress {
				state = scenestate.Next(state, scenestate.EventStart)
				line = scenestate.Line(scenestate.Title)
			}

		case scenestate.GameOver:
			if startPress {
				hero, crane, lives, reels, line = freshRun(take)
				state = scenestate.Next(state, scenestate.EventStart)
			} else if leavePress {
				state = scenestate.Next(state, scenestate.EventLeave)
				line = scenestate.Line(scenestate.Title)
			}
		}

		render.DrawFrame(render.World{
			Hero:       hero,
			Take:       take,
			Crane:      crane,
			ReelsFound: reels.Count(),
			ReelsTotal: reels.Total(),
			Lives:      lives.Lives(),
			State:      state,
			Line:       line,
		})
	}
}

// freshRun resets the session's plain state for a new take: a fresh hero, nine
// lives, no reels and a fresh crane. It also starts the opening intercom line
// and its crackle, since Playing is beginning.
func freshRun(take leveldata.Take) (move.Hero, enemy.Crane, lives.Lives, pickups.Reels, string) {
	render.PlayCrackle()
	return move.NewHero(), enemy.New(take.Track), lives.New(), pickups.New(), scenestate.Line(scenestate.Playing)
}

// heroAtRespawn returns the hero standing on the floor at p's sideways and
// forward spot, where the respawn point places him. The checkpoint's Y is the
// clapperboard's board height, not a standing height, so only its X and Z
// place the hero.
func heroAtRespawn(p leveldata.Point) move.Hero {
	return move.Hero{X: p.X, Y: 0, Z: p.Z}
}
