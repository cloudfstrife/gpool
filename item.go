package gpool

//Item pool item
type Item interface {
	Initial(map[string]string) error
	Destory() error
	Check() error
}
