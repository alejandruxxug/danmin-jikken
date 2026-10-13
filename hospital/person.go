package hospital

// Person is the identity shared by patients and staff. It is embedded by
// value in Patient, Doctor, Orderly and Nurse, so its methods are promoted to them
// (patient.Name() works) and the three fields are never written twice.
type Person struct {
	id   string
	name string
	age  int
}

// NewPerson builds the identity part of a patient or staff member.
func NewPerson(id, name string, age int) Person {
	return Person{id: id, name: name, age: age}
}

func (p Person) ID() string   { return p.id }
func (p Person) Name() string { return p.name }
func (p Person) Age() int     { return p.age }
