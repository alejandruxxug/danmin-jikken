package game

import (
	"errors"
	"fmt"
)

// revealJumps are the stages -you-reveal can jump to (the backdoor's way to
// the storyboard frames): the stage, and how far into it.
var revealJumps = map[string]struct {
	stage RevealStage
	tick  int
}{
	"":     {RevealOff, 0},
	"sb1":  {RevealArrival, 0},
	"sb3":  {RevealBack, 0},
	"tilt": {RevealTilt, tiltTicks / 2},
	"face": {RevealFace, 0},
	"stun": {RevealStun, 0},
}

// jumpToReveal runs day 3 up to the moment you stand at the window (every
// patient before you is sent out with REFER on the way), and puts the reveal
// at the stage asked for. From sb3 on, your own folder is stamped ADMIT first,
// so stamped_self is set as in a real run.
func (e *Env) jumpToReveal(name string) error {
	if e.Day != 3 {
		return errors.New("-you-reveal needs -day 3")
	}
	if err := e.StartShift(); err != nil {
		return err
	}
	sh := e.Shift()
	for range 20000 {
		if g := sh.window(); g != nil && !g.Self && sh.atWindow(g) {
			_ = sh.Stamp(StampRefer)
		}
		e.tickClock()
		sh.Update()
		e.UpdateReveal(sh)
		if g := e.Reveal.Guest; g != nil && sh.atWindow(g) {
			break
		}
	}
	if e.Reveal.Guest == nil || !sh.atWindow(e.Reveal.Guest) {
		return fmt.Errorf("-you-reveal %s: you never reached the window", name)
	}
	jump := revealJumps[name]
	if jump.stage > RevealArrival {
		if err := sh.Stamp(StampAdmit); err != nil {
			return fmt.Errorf("-you-reveal %s: stamping your folder: %w", name, err)
		}
	}
	e.Reveal.Stage, e.Reveal.Tick = jump.stage, jump.tick
	return nil
}
