package types

type Message struct {
	Content []Content
}

type Content struct {
	Text   string
	Images [][]byte
}
