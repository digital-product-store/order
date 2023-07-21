package product

type Product struct {
	Id           string `json:"id"`
	UploadId     string `json:"upload_id"`
	ObjectId     string `json:"object_id"`
	OriginalName string `json:"original_name"`
	BookName     string `json:"book_name"`
	Author       string `json:"author"`
	Summary      string `json:"summary"`
	Price        string `json:"price"`
}
