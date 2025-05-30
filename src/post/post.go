package post

type Post struct {
	ID          int
	Title       string
	URL         string
	Description string
	CreatedAt   string
	OwnerID     string
}

var PostNull = Post{
	ID: -1,
}
