package shortener

type RedirectRepository interface {
	Find(code string) (*Redirect, error)
	Store(*Redirect) error
}
