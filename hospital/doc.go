// Package hospital models the Hospital de los Costeños con Narcolepsia:
// patients who fall asleep anywhere, the staff who attend them, the rooms
// they are moved to, and the record of every sleep episode.
//
// The package holds the business rules only. It never prints and never
// draws: main prints the demo scenario and the game draws the screens, and
// both talk to the model through the methods of Hospital.
//
// Concurrency: Hospital is safe for concurrent use. Patient, Room and the
// setup calls (Doctor.DiagnosePatient) are not; code running in goroutines
// must change patients only through Hospital's methods, which take the
// hospital's lock, and read them while goroutines run through Hospital.Snapshot.
// The staff's own history (MyEpisodes) has its own small lock.
package hospital
