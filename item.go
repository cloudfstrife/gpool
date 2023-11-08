package gpool

//Item pool item
type Item interface {
	Initial(map[string]string) error
	Destory(map[string]string) error
	Check(map[string]string) error
}
