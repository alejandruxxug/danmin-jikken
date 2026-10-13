// Package hospital models the Hospital de los Costeños con Narcolepsia:
// patients who fall asleep anywhere, the staff who attend them, the rooms
// they are moved to, and the record of every sleep episode.
//
// The package holds the business rules only. It never prints and never
// draws: main prints the demo scenario and the game draws the screens, and
// both talk to the model through the methods of Hospital.
//
// Concurrency: Hospital is safe for concurrent use. Patient, Doctor, Orderly
// and Room are not; code running in goroutines must change them only through
// Hospital's methods, which take the hospital's lock.
package hospital
