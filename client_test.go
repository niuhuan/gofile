package gofile

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func client() *Client {
	token := os.Getenv("GO_FILE_TOKEN")
	if token == "" {
		panic("GO_FILE_TOKEN NOT SET")
	}
	return &Client{
		Token: token,
	}
}

func root() string {
	root := os.Getenv("GO_FILE_ROOT_ID")
	if root == "" {
		panic("GO_FILE_ROOT_ID NOT SET")
	}
	return root
}

func debugPrint(msg interface{}, err error) {
	if err != nil {
		panic(err)
	}
	buff, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	println(string(buff))
}

func TestClient_GetServer(t *testing.T) {
	debugPrint(client().GetServer())
}

func TestClient_GetAccountDetails(t *testing.T) {
	debugPrint(client().GetAccountDetails())
}

func TestClient_CreateFolder(t *testing.T) {
	debugPrint(client().CreateFolder(root(), "aaa666"))
}

func TestClient_UploadFile(t *testing.T) {
	file, err := os.Open("2B.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	debugPrint(client().UploadFile(
		"store3",
		root(),
		"2B.png",
		file, TempBuffer{
			TempBufferType:   FileInTempDir,
			SpecifyTmpFolder: "/Volumes/DATA/Temp",
		},
		func(total int64, send int64) {
			println(fmt.Sprintf("%v/%v", send, total))
		},
	))
}

func TestClient_SetFolderOption(t *testing.T) {
	debugPrint("OK", client().SetFolderOption(
		root(),
		"public",
		"true",
	))
}
