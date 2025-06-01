package post

type Post struct {
	ID               int
	Title            string
	URL              string
	Description      string
	CreatedAt        string
	UserID           string
	UserName         string
	NumberOFComments int
}

var PostNull = Post{
	ID: -1,
}
