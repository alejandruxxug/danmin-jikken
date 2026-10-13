// Command danmin runs the demonstration scenario of the assignment (Section
// 6) against the hospital package and prints the four required queries.
// It only builds the scenario and prints; every rule lives in hospital.
//
// The game, 断眠実験 DANMIN JIKKEN, is built on the same package separately.
package main

import (
	"fmt"
	"os"
	"sort"

	"danmin/hospital"
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
