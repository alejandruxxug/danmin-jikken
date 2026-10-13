// Command danmin runs the demonstration scenario of the assignment (Section
// 6) against the hospital package and prints the four required queries.
// It only builds the scenario and prints; every rule lives in hospital.
//
// The game, 断眠実験 DANMIN JIKKEN, is built on the same package separately.
package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"danmin/hospital"
	"danmin/simulation"
)

func main() {
	h := hospital.NewHospital("Halcyon Rest")

	barrow := hospital.NewDoctor("D-01", "Dr. Barrow", 52, "sleep medicine")
	sato := hospital.NewDoctor("D-02", "Dr. Sato", 38, "neurology")
	must(h.HireDoctor(barrow))
	must(h.HireDoctor(sato))
	must(h.HireStaff(hospital.NewOrderly("O-01", "Gus the orderly", 29)))

	for _, n := range []int{101, 102, 103} {
		must(h.AddRoom(hospital.NewRoom(n, 1)))
	}

	cecily := hospital.NewPatient("P-001", "Old Cecily", 71, hospital.Severe)
	matty := hospital.NewPatient("P-002", "Matty", 12, hospital.Moderate)
	wren := hospital.NewPatient("P-003", "Wren Calloway", 34, hospital.Mild)
	hank := hospital.NewPatient("P-004", "Hank Shoals", 58, hospital.Severe)
	ivy := hospital.NewPatient("P-005", "Ivy Marsh", 26, hospital.Moderate)
	for _, p := range []*hospital.Patient{cecily, matty, wren, hank, ivy} {
		must(h.AdmitPatient(p))
	}
	must(barrow.DiagnosePatient(cecily))
	must(barrow.DiagnosePatient(hank))
	must(sato.DiagnosePatient(matty))

	section("Sleep attacks")
	attacks := []struct {
		p        *hospital.Patient
		location string
	}{
		{cecily, "cafeteria"},
		{matty, "radiology queue"},
		{wren, "gift shop"},
		{hank, "hallway 2, second floor"},
	}
	for _, a := range attacks {
		must(h.RegisterEpisode(a.p, a.location))
		fmt.Printf("  %-6s %-14s fell asleep in %-24s -> %s\n", a.p.ID(), a.p.Name(), a.location, a.p.State())
	}

	section("Query 5.3: assigning a room to " + hank.ID())
	if _, err := h.AssignRoom(hank); err != nil {
		fmt.Printf("  %v\n", err)
	}

	printHallway(h, "Query 5.1: patients asleep in a hallway")

	section(matty.ID() + " wakes up, then Query 5.3 again for " + hank.ID())
	must(h.WakePatient(matty))
	fmt.Printf("  %s is %s, room 102 is free again\n", matty.Name(), matty.State())
	if room, err := h.AssignRoom(hank); err != nil {
		fmt.Printf("  %v\n", err)
	} else {
		fmt.Printf("  %s %s -> room %d\n", hank.ID(), hank.Name(), room.Number())
	}

	printHallway(h, "Query 5.1 again, after the wake-up")

	for _, d := range h.Doctors() {
		section("Query 5.2: episodes attended by " + d.Name())
		eps := d.MyEpisodes()
		if len(eps) == 0 {
			fmt.Println("  none")
		}
		for _, e := range eps {
			fmt.Println("  " + e.Summary())
		}
	}

	section("Rooms at the end of the shift")
	for _, r := range h.Rooms() {
		who := "empty"
		if occ := r.Occupants(); len(occ) > 0 {
			who = occ[0].ID() + " " + occ[0].Name()
		}
		fmt.Printf("  room %d  %-9s %s\n", r.Number(), r.State(), who)
	}

	section("Query 5.4: severity report (Severe patients)")
	report := h.SevereReport()
	severe := make([]*hospital.Patient, 0, len(report))
	for p := range report {
		severe = append(severe, p)
	}
	sort.Slice(severe, func(i, j int) bool { return severe[i].ID() < severe[j].ID() })
	for _, p := range severe {
		fmt.Printf("  %-6s %-14s %d episode(s) today\n", p.ID(), p.Name(), report[p])
	}

	runConcurrentDemo()
}

// runConcurrentDemo is the bonus: every patient of a second hospital is a
// goroutine that falls asleep at random times. The numbers it prints do not
// depend on the timing, so the output is the same on every run.
func runConcurrentDemo() {
	const attacksEach = 4
	h := hospital.NewHospital("Halcyon Rest (night ward)")
	must(h.HireDoctor(hospital.NewDoctor("D-11", "Dr. Okada", 45, "sleep medicine")))
	must(h.HireStaff(hospital.NewOrderly("O-11", "Gus the orderly", 29)))
	must(h.HireStaff(hospital.NewNurse("N-11", "Nurse Hana", 31, "east ward")))
	must(h.AddRoom(hospital.NewRoom(201, 1)))
	must(h.AddRoom(hospital.NewRoom(202, 1)))
	var patients []*hospital.Patient
	for i := 1; i <= 6; i++ {
		p := hospital.NewPatient(fmt.Sprintf("P-%03d", 100+i), fmt.Sprintf("Sleeper %d", i), 20+i, hospital.Severe)
		must(h.AdmitPatient(p))
		patients = append(patients, p)
	}

	section("Concurrent simulation: 6 patient goroutines, 2 rooms")
	err := simulation.Run(context.Background(), h, patients, simulation.Config{
		Seed:      317,
		MaxAwake:  15 * time.Millisecond,
		MaxAsleep: 15 * time.Millisecond,
		Attacks:   attacksEach,
		Locations: []string{"cafeteria", "gift shop", "radiology queue", "hallway 2"},
	})
	if err != nil {
		fmt.Printf("  simulation failed: %v\n", err)
		return
	}
	snap := h.Snapshot()
	awake := 0
	for _, p := range snap.Patients {
		if p.State == hospital.Awake {
			awake++
		}
	}
	fmt.Printf("  episodes recorded: %d (6 patients x %d attacks)\n", len(h.History()), attacksEach)
	fmt.Printf("  patients awake at the end: %d of %d\n", awake, len(snap.Patients))
	for _, a := range h.Staff() {
		fmt.Printf("  %-16s %T\n", a.Name(), a)
	}
}

func printHallway(h *hospital.Hospital, title string) {
	section(title)
	hallway := h.PatientsInHallway()
	if len(hallway) == 0 {
		fmt.Println("  none")
	}
	for _, p := range hallway {
		fmt.Printf("  %-6s %-14s (%s)\n", p.ID(), p.Name(), p.Location())
	}
}

func section(title string) {
	fmt.Printf("\n== %s ==\n", title)
}

// must stops the demo if a step that should never fail does.
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "demo setup failed:", err)
		os.Exit(1)
	}
}
