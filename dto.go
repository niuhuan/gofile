package gofile

type Response[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type ServerResult struct {
	Server string `json:server`
}

type AccountDetails struct {
	Token                  string      `json:"token"`
	Email                  string      `json:"email"`
	Tier                   string      `json:"tier"`
	RootFolder             string      `json:"rootFolder"`
	FilesCount             int         `json:"filesCount"`
	FilesCountLimit        interface{} `json:"filesCountLimit"`
	TotalSize              int         `json:"totalSize"`
	TotalSizeLimit         interface{} `json:"totalSizeLimit"`
	Total30DDLTraffic      int         `json:"total30DDLTraffic"`
	Total30DDLTrafficLimit interface{} `json:"total30DDLTrafficLimit"`
}

type FileUpload struct {
	DownloadPage string `json:"downloadPage"`
	Code         string `json:"code"`
	ParentFolder string `json:"parentFolder"`
	FileId       string `json:"fileId"`
	FileName     string `json:"fileName"`
	Md5          string `json:"md5"`
}

type ContentResult struct {
	IsOwner            bool                   `json:"isOwner"`
	Id                 string                 `json:"id"`
	Type               string                 `json:"type"`
	Name               string                 `json:"name"`
	ParentFolder       string                 `json:"parentFolder"`
	Code               string                 `json:"code"`
	CreateTime         int                    `json:"createTime"`
	Public             bool                   `json:"public"`
	Childs             []string               `json:"childs"`
	TotalDownloadCount int                    `json:"totalDownloadCount"`
	TotalSize          int                    `json:"totalSize"`
	Contents           map[string]ContentInfo `json:"contents"`
}

type ContentInfo struct {
	Id            string `json:"id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	ParentFolder  string `json:"parentFolder"`
	CreateTime    int    `json:"createTime"`
	Size          int    `json:"size"`
	DownloadCount int    `json:"downloadCount"`
	Md5           string `json:"md5"`
	Mimetype      string `json:"mimetype"`
	ServerChoosen string `json:"serverChoosen"`
	DirectLink    string `json:"directLink"`
	Link          string `json:"link"`
}

type TempBufferType int

const (
	Ram TempBufferType = iota
	FileInTempDir
)

type TempBuffer struct {
	TempBufferType   TempBufferType
	SpecifyTmpFolder string
}
