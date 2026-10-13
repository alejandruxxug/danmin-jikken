# 🌙 Danmin Jikken – Night 3

Back to [[Danmin Jikken – Script]] · Before: [[Danmin Jikken – Night 2]] + ☀️ Day 3 · Ends the game

> **Hana stopped reporting. Find out why, and then the gurney again.** ~950 words with all endings (one playthrough sees ~600). Plays by `route` (set in [[Danmin Jikken – Night 2#Choice 2]]).
> **Day 3 before this:** rule *"Do not wake them."* Red leaked into the day. **Hana didn't clock in**, but her reports kept coming, **in round, happy handwriting**. The last patient was you: your Settler ID, `{player}`. You stamped your own file (`stamped_self`).

> ⚠️ **Content:** Hana's death (off-screen) and, on Hana's route, Mayu's suicide (silhouette, hard cut at the drop). See [[Danmin Jikken – Story#✍️ Handling the deaths]].

---

## Waking up

[FX: white fades to grey, faster than before]
[BG: ward C, bed 4]

*I'm not at the desk.*

*I was never at the desk.*

*There's a band on my wrist. It says 4.*

[BG: nurses' station, empty]
[SFX: monitor beeping]

*Nobody at the station. The monitor beeps by itself. Seven lines.*

*Hana didn't come in today. But somebody filed her reports.*

## Hana's room

[BG: nurses' quarters corridor. One door, open a crack.]

*Her name is on the door, in her small, neat letters.*

[SFX: the door creaks]
[BG: black. **We never see the room.**]

*Hana is there.*

*She isn't asleep. I know before I touch her hand.*

*Fourteen days awake, and now she'll never have to fight it again.*

*The floor is covered in pale cans. Dozens.*

*The newest ones are still sealed. She stopped.*

*On the desk there's an envelope. It says* **Four**.

## Choice 4

> - **Read the letter** → [[#4A Read the letter]] `→ set read_letter`
> - **Leave it** → [[#4B Leave it]]

### 4A Read the letter

`→ set read_letter`

`if route = mayu:`
**Letter** *(Hana's handwriting)*:
Four. You chose her. I understand.
I'd choose the white place too, if you were in it.
I kept myself awake for you. I don't have a reason to anymore.
Please don't close your eyes. Not for her. Not for me.
— H.

`if route = nobody:`
**Letter** *(Hana's handwriting)*:
Four. You didn't read my note. That's okay.
I wrote you another one.
I was always going to say it, and then I never got to.
Please don't close your eyes.
— H.

`if route = hana:`
**Letter**:
I'm so sorry!! I was just SO tired ♡
Don't be sad, okay? Be happy. Be happy with Mayu!
— Hana

*The loops on the letters. Round, happy loops.*

*I've been reading this handwriting on Hana's reports all day.*

→ [[#The gurney]]

### 4B Leave it

*I can't. Not yet.*

*I leave it on the desk with her.*

→ [[#The gurney]]

## The gurney

[SFX: footsteps, many, fast]
[FX: hands. The ceiling again. Moving.]
[SFX: squeaking wheels]

*They found me. I'm on my back. The ceiling slides past, the lights, one, two, three.*

*Like the first time.*

[BG: ward C, moving down the rows]
[SFX: a clock, somewhere, chimes]
[FX: every bed sits up at once. Red eyes. The clock says 3:17.]

*Every bed sits up at once.*

*Every face turns to watch me go by.*

[BG: procedure room. A man in a white coat. Smiling like the posters.]

**Dr. Barrow** *(smile)*: Good evening, Four. You've been a wonderful employee.

**Dr. Barrow** *(smile)*: S-317 isn't something we found. It's something we made. And we needed people nobody would check on.

**Dr. Barrow** *(smile)*: Who checks on a sleepy fisherman?

**Dr. Barrow** *(smile)*: And the nurses' little white drink? Ours too. We sell the cure to the problem. That's just good business.

**Dr. Barrow** *(smile)*: The dream is the cure. As long as you never leave it. Tonight we make your rest permanent.

*Someone takes my hand.*

`if route = mayu:`
**Mayu** *(soft)*: It's okay. I'll be right there with you. We'll have so many first days.

`if route = nobody:`
**Mayu** *(cold)*: Don't look at me like that. I'm just doing my job.

`if route = hana:`
**Mayu** *(calm)*: Shh. It's just us now.
*Her hand is very calm. That's the scariest thing about it.*

**Dr. Barrow** *(smile)*: Count back from ten for me.

## Choice 5

> - **Close your eyes** → `if route = hana:` [[#Forever ending]] · `else:` [[#Sleep ending]]
> - **Keep them open** → `if route = hana:` [[#Forever ending]] · `else if stamped_self = DISCHARGE and read_letter:` [[#Secret ending]] · `else:` [[#Wake ending]]

---

## Sleep ending

*Ten… nine…*

*It's easier. It's so much easier.*

[FX: eyes close]
[FX: hard cut to WHITE]
[MUSIC: Driftwell jingle]

**Orientation reel**: Welcome to your first day at Halcyon Rest!

[BG: the desk, day look]

`if route = mayu:`
*There's a new clerk at the next desk.*
**Mayu** *(day look, happy)*: Hi! I'm new too. Isn't this place *perfect*?

`if route = nobody:`
*The next desk is empty. It's always been empty.*

[FX: the lines boil. And boil. And boil.]

**SLEEP ENDING**

## Wake ending

*Ten…*

*No.*

[FX: the screen shakes. Grey, sharp, real.]

*I grab the rail. I pull the line out of my arm. It hurts. It's the first thing that's hurt in days.*

**Dr. Barrow** *(blank)*: Hold them— hold them down—

`if route = mayu:`
**Mayu** *(crying)*: No, no, no, you're supposed to stay, you *promised*—

*They take her instead. Gently. They put her to sleep, and this time someone will be watching.*

`if read_letter:`
*In my pocket, Hana's letter. "Please don't close your eyes."*
*I didn't.*

[FX: fade to grey, not white]

**WAKE ENDING**

## Forever ending

**Hana's route, whatever you chose.**

`if you kept your eyes open:`
*I won't. I won't close them.*
**Mayu** *(soft)*: I know. That's why I'm doing it for you.
*Her hand on my eyes.*

**Mayu** *(soft)*: Shh. I'm sorry. I'm sorry. Now nobody can take you.

[SFX: the drip, faster]
[FX: eyes closing, slowly. Only a thin gap left.]

*Through the gap, I can still see her.*

*She's standing on the chair.*

*She's smiling at me.*

[FX: silhouette only. She steps off.]
[FX: **HARD CUT TO BLACK** at the drop]
[MUSIC: off. SFX: off.]

**FOREVER ENDING**

## Secret ending

**Wake, but you stamped DISCHARGE on your own file and read Hana's letter (Mayu or Nobody route).**

*Ten…*

*No.*

[FX: the screen shakes. Grey. Real.]

*I sit up. I'm awake. I'm really awake.*

[FX: the game "falls asleep": the screen dims, input stops working for 5 seconds]
[FX: the text box types by itself]

**…**: …

[FX: the window closes]

*(The game quits. In the game folder, a new file:* `sleep_log_317.txt`*)*

```
SLEEP LOG — BED 4
Day 1   awake 0 h
Day 2   awake 0 h
Day 3   awake 0 h
Night   awake 3:17

If you're reading this, you're awake.
Stay that way. For me.
— H.
```

**3:17**

---

**Notes**
- **Choice 5 doesn't matter on Hana's route**, on purpose: the choice was already made on night 2. Mayu takes it away from you.
- **The secret** needs a choice from the day shift (`stamped_self`), so the day and the night really are one game.
- **Mayu's death is the only death shown, and only as a silhouette until she steps off.** Hard cut, no body, no aftermath (Alejo's staging, 2026-10-13).
