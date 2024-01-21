package domain

type SearchType uint

const (
	Strong SearchType = iota
	Weak
	Both
)

type Individual struct {
	Uid       int    `json:"uid"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
