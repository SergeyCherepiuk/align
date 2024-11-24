package resources

type Correction func() error

type Resource interface {
	Id() string
	Check() ([]Correction, error)
}
