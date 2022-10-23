package gofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const ApiUri = "https://api.gofile.io"

type Client struct {
	http.Client
	Token string
}

func RequestResponse[T any](c *Client, req *http.Request) (*T, error) {
	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	buff, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	var resp Response[T]
	err = json.Unmarshal(buff, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, errors.New(resp.Status)
	}
	return &resp.Data, err
}

func ReqResponse[T any](c *Client, method string, path string, query map[string]string, params map[string]string) (*T, error) {
	var body io.Reader
	if params != nil {
		paramsStr := url.Values{}
		for k := range params {
			paramsStr.Add(k, params[k])
		}
		body = bytes.NewBuffer([]byte(paramsStr.Encode()))
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%v/%v", ApiUri, path), body)
	if err != nil {
		return nil, err
	}
	if query != nil {
		queryStr := req.URL.Query()
		for k := range query {
			queryStr.Add(k, query[k])
		}
		req.URL.RawQuery = queryStr.Encode()
	}
	if params != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	return RequestResponse[T](c, req)
}

func (c *Client) GetServer() (*string, error) {
	serverResult, err := ReqResponse[ServerResult](c, "GET", "getServer", nil, nil)
	if err != nil {
		return nil, err
	}
	return &serverResult.Server, nil
}

func (c *Client) GetAccountDetails() (*AccountDetails, error) {
	return ReqResponse[AccountDetails](c, "GET", "getAccountDetails", map[string]string{
		"allDetails": "true",
		"token":      c.Token,
	}, nil)
}

func (c *Client) CreateFolder(parentFolderId string, folderName string) (*FolderCreated, error) {
	return ReqResponse[FolderCreated](c, "PUT", "createFolder", nil, map[string]string{
		"parentFolderId": parentFolderId,
		"folderName":     folderName,
		"token":          c.Token,
	})
}

func (c *Client) CopyContent(folderIdDest string, contentsId []string) error {
	_, err := ReqResponse[interface{}](c, "PUT", "copyContent", nil, map[string]string{
		"folderIdDest": folderIdDest,
		"contentsId":   strings.Join(contentsId, ","),
		"token":        c.Token,
	})
	return err
}

func (c *Client) DeleteContent(contentsId []string) error {
	_, err := ReqResponse[interface{}](c, "DELETE", "deleteContent", nil, map[string]string{
		"contentsId": strings.Join(contentsId, ","),
		"token":      c.Token,
	})
	return err
}

func (c *Client) UploadFile(
	server string,
	folderId string,
	fileName string,
	fileReader io.Reader,
	tempBuffer TempBuffer,
	onSend func(total int64, send int64),
) (*FileUpload, error) {
	tmp := &tmp{
		tempBuffer: tempBuffer,
		onSend:     onSend,
	}
	if err := tmp.init(); err != nil {
		return nil, err
	}
	defer tmp.close()
	url := fmt.Sprintf("https://%v.gofile.io/uploadFile", server)
	err := func() error {
		fw, err := tmp.w.CreateFormFile("file", fileName)
		if err != nil {
			return err
		}
		if _, err = io.Copy(fw, fileReader); err != nil {
			return err
		}
		fw, err = tmp.w.CreateFormField("folderId")
		if err != nil {
			return err
		}
		fw.Write([]byte(folderId))
		fw, err = tmp.w.CreateFormField("token")
		if err != nil {
			return err
		}
		fw.Write([]byte(c.Token))
		return nil
	}()
	if err != nil {
		return nil, err
	}
	if err := tmp.flip(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, tmp)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", tmp.w.FormDataContentType())
	return RequestResponse[FileUpload](c, req)
}

func (c *Client) GetContent(contentId string) (*ContentResult, error) {
	return ReqResponse[ContentResult](c, "GET", "deleteContent", map[string]string{
		"contentId": contentId,
		"token":     c.Token,
	}, nil)
}

//
// SetFolderOption
//
// Set an option on a folder
//
// Parameters
//  required token
//    The access token of an account. Can be retrieved from the profile page.
//  required folderId
//    The folder ID.
//  required option
//    Can be "public", "password", "description", "expire" or "tags".
//  required value
//    The value of the option to be defined.
//    For "public", can be "true" or "false".
//    For "password", must be the password.
//    For "description", must be the description.
//    For "expire", must be the expiration date in the form of unix timestamp.
//    For "tags", must be a comma seperated list of tags.
func (c *Client) SetFolderOption(folderId string, option string, value string) error {
	_, err := ReqResponse[interface{}](c, "PUT", "setFolderOption", nil, map[string]string{
		"folderId": folderId,
		"option":   option,
		"value":    value,
		"token":    c.Token,
	})
	return err
}

type tmp struct {
	tempBuffer TempBuffer
	ramBuffer  *bytes.Buffer
	file       *os.File
	reader     io.Reader
	w          *multipart.Writer
	onSend     func(total int64, send int64)
	total      int64
	readOver   int64
}

func (t *tmp) init() error {
	if t.tempBuffer.TempBufferType == Ram {
		t.ramBuffer = &bytes.Buffer{}
		t.w = multipart.NewWriter(t.ramBuffer)
		return nil
	} else if t.tempBuffer.TempBufferType == FileInTempDir {
		var err error
		t.file, err = os.CreateTemp(t.tempBuffer.SpecifyTmpFolder, "GO_FILE_TMP_")
		if err != nil {
			return err
		}
		t.w = multipart.NewWriter(t.file)
		return nil
	}
	return errors.New("error tmp buffer type")
}

func (t *tmp) flip() error {
	err := t.w.Close()
	if err != nil {
		return err
	}
	if t.tempBuffer.TempBufferType == Ram {
		t.total = int64(t.ramBuffer.Len())
		t.reader = t.ramBuffer
		return nil
	} else if t.tempBuffer.TempBufferType == FileInTempDir {
		err := t.file.Sync()
		if err != nil {
			return err
		}
		_, err = t.file.Seek(0, 0)
		if err != nil {
			return err
		}
		stat, err := t.file.Stat()
		if err != nil {
			return err
		}
		t.total = stat.Size()
		t.reader = t.file
		return nil
	}
	return errors.New("error tmp buffer type")
}

func (t *tmp) close() error {
	if t.tempBuffer.TempBufferType == Ram {
		return nil
	} else if t.tempBuffer.TempBufferType == FileInTempDir {
		if t.file != nil {
			err := t.file.Close()
			if err != nil {
				return err
			}
			return os.Remove(t.file.Name())
		}
		return nil
	}
	return errors.New("error tmp buffer type")
}

func (t *tmp) Read(p []byte) (n int, err error) {
	n, err = t.reader.Read(p)
	if t.onSend != nil {
		t.readOver += int64(n)
		t.onSend(t.total, t.readOver)
	}
	return
}
