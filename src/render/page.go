package render

type Page struct {
	Title string
	Data  interface{}
	User  struct{ Name string }
}
