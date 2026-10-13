# 🌙 Danmin Jikken – Night 2

Back to [[Danmin Jikken – Script]] · Before: [[Danmin Jikken – Night 1]] + ☀️ Day 2 · Next: ☀️ Day 3, then [[Danmin Jikken – Night 3]]

> **The confession night.** Both nurses are on shift. Choice 2 decides your route, and the route decides who confesses. ~760 words.
> **Day 2 before this:** new rule *"Shoal patients → Ward C."* Every sleep log said **3:17**. Old Cecily laughed and collapsed (cataplexy).

---

## Waking up

[FX: **STROBE IN**, the night frames last longer. See [[Danmin Jikken – Art Direction#⚡ Day → night: the strobe (Alejo, 2026-10-13)]]. Fall into the grey]
[BG: ward C, dark. Bed 4.]

*I fell asleep at the desk again.*

*I have to stop doing that.*

[BG: nurses' station. Mayu and Hana, both there.]

**Mayu** *(bright)*: There they are!

`if liked_mayu:`
**Mayu** *(happy)*: You stayed with me last night. Nobody ever stays.

`if liked_hana:`
**Hana** *(flat)*: You read it.
*It isn't a question.*

**Mayu** *(sad)*: Did you hear about Cecily? She laughed so hard she just… fell. Right at your desk.

**Hana** *(flat)*: She didn't fall at a desk. She fell in bed 2.

**Mayu** *(laugh)*: Hana! Don't be weird.

*Hana looks at me for a long time. Then she picks up a clipboard.*

**Hana** *(flat)*: Rounds.

[SFX: footsteps, going]

**Mayu** *(smile)*: She looks at you a lot, you know.

**Mayu** *(blank)*: I could make her stop.

*A beat.*

**Mayu** *(laugh)*: Looking! I mean, I could *ask* her. Kidding. Your face!

## Choice 2

> - **Stay with Mayu** → [[#Mayu route]] `→ set stay_mayu`
> - **Go find Hana** → [[#Hana route]] `→ set find_hana`
> - **Call Mayu out** → [[#Call Mayu out]] `→ set confront_mayu`

### Call Mayu out

`→ set confront_mayu`

*"What did you mean, make her stop?"*

**Mayu** *(blank)*: …

**Mayu** *(smile)*: Wow. Okay. It was a *joke*.

**Mayu** *(cold)*: I'll go do something useful, then.

*She leaves. The station is very quiet.*

`if liked_hana:` → [[#Hana route]] (from **Hana heard you**)
`else:` → [[#Nobody route]]

---

### Mayu route

`→ route = mayu`

**Mayu** *(happy)*: You're staying? Again?

*She sits on the counter, legs swinging. For a minute she's just a girl on a long shift.*

**Mayu** *(soft)*: Can I tell you something? Promise not to laugh.

**Mayu** *(shy)*: I like you. A lot. Since the first night.

**Mayu** *(shy)*: Everyone here is so tired. And then you came, and you're… you keep waking up. Every night. For me.

**Mayu** *(smile)*: So I'm going to keep you.

*She takes my hand. Her grip is too tight.*

**Mayu** *(blank)*: Nobody else gets you. Not Driftwell. Not Hana. If anyone tried, I'd—

*She stops.*

**Mayu** *(laugh)*: I'd be *sad*. I'd be so sad. That's all.

*There's no "kidding" this time.*

[SFX: footsteps, stopping]

*At the end of the corridor, Hana is standing very still, holding her clipboard.*

*Then she turns around and walks away.*

→ [[#Kanashibari]]

---

### Hana route

`→ route = hana`

`if find_hana:`
[BG: supply room. One bulb. Hana drawing white liquid out of a pale can into a syringe.]

**Hana** *(flat)*: You're not supposed to be out of bed, Four.

`if confront_mayu:` *(Hana heard you)*
[SFX: footsteps]
**Hana** *(flat)*: You asked her. Nobody asks her.

**Hana** *(tired)*: Sit.

*She doesn't look up from the syringe. She finds a vein without looking. She's done this a lot.*

**Hana** *(flat)*: Pale MSTR. Driftwell's. It keeps the white place away. For a while.

**Hana** *(tired)*: I've been awake fourteen days. Do you know why?

**Hana** *(tired)*: Because when I sleep, I go where you are. The white place. And it's lovely there. And something in it wants to keep me.

**Hana** *(soft)*: You've been there since the gurney. Every night you come up for air a little less.

*She finally looks at me.*

**Hana** *(soft)*: I don't want you to go back under.

**Hana** *(blush)*: Because I'd miss you.

*That's all. Hana doesn't say things twice.*

[SFX: a floorboard]

*Behind her, the door is open a crack.*

*Mayu's face is in the gap. She's smiling.*

[FX: the door closes, softly]

→ [[#Kanashibari]]

---

### Nobody route

`→ route = nobody`

*I wait for Hana to come back from rounds.*

[SFX: footsteps]

**Hana** *(flat)*: You didn't read my note.

**Hana** *(flat)*: It's fine. Most people don't.

*She goes past me. Mayu doesn't come back.*

*The monitor beeps. Seven lines. Nobody to tell.*

→ [[#Kanashibari]]

---

## Kanashibari

[BG: ward C, bed 4. Darker than before.]
[MUSIC: off]

*Later. I'm in bed. I don't remember lying down.*

*I can't move.*

*Not my hands. Not my eyes. I can breathe, barely.*

[FX: a figure in the doorway, black on grey. No face.]

*Someone is standing in the doorway.*

*Watching me.*

[FX: the text box types by itself]

**…**: don't wake up don't wake up don't wake up

`if route = mayu:`
*Whoever it is, they're holding a clipboard.*

`if route = hana:`
*The figure tilts its head. I know that smile, even in the dark.*
**Mayu** *(whisper)*: I saw.

`if route = nobody:`
*It stands there until I stop being able to tell if it's there at all.*

[FX: the figure is gone]
[SFX: I can breathe again]

## Back to sleep

`if route = mayu:`
**Mayu** *(soft)*: Shh. You're all mine now. Sleep tight.

`if route = hana:`
**Hana** *(crying)*: I have to. If I don't, they send someone who won't come back for you.
**Hana** *(crying)*: Count back from ten.

`if route = nobody:`
*Nobody comes. The drip clicks by itself.*

[SFX: click]
[FX: eyes close]
[FX: hard cut to WHITE]

→ ☀️ **Day 3**

---

**Notes**
- **Mayu's slip #2** plays for everyone, before Choice 2. On her route her confession holds **slip #3, the one with no "kidding"**.
- **The figure in the doorway**: Hana on Mayu's route (she watched you choose), Mayu on Hana's route (she saw the confession), nobody, or maybe nothing, on the Nobody route.
- Hana's line *"You've been there since the gurney"* is the first time anyone tells you the truth. On the other routes you never hear it.
