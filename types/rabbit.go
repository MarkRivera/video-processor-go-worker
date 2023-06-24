package types

type RabbitTask struct {
	ChunkName   string `json:"chunkName"`
	ChunkNumber int    `json:"chunkNumber"`
	Filename    string `json:"filename"`
	Data        string `json:"data"`
	TotalChunks int    `json:"totalChunks"`
}
