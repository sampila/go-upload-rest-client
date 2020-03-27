package rest_upload

import (
	"strings"
	"github.com/mercadolibre/golang-restclient/rest"
	"time"
	"encoding/json"
  "mime/multipart"
	"github.com/sampila/go-utils/rest_errors"
	"errors"
  "bytes"
  "io/ioutil"
)

/*const (
	headerXPublic   = "X-Public"
	headerXClientId = "X-Client-Id"
	headerXCallerId = "X-Caller-Id"

	paramAccessToken = "access_token"
)*/

var (
	uploadRestClient = rest.RequestBuilder{
		BaseURL: "http://localhost:9010",
    ContentType: rest.MULTIPART,
		Timeout: 3000 * time.Millisecond,
	}
)

type FileRequest struct {
  StoreCode string  `json:"store_code"`
  File *multipart.FileHeader `json:"file"`
}

func UploadFile(p *FileRequest) (interface{}, rest_errors.RestErr) {
  bodyBuf := &bytes.Buffer{}
  bodyWriter := multipart.NewWriter(bodyBuf)
  file, err := p.File.Open()
  defer file.Close()
  if err != nil {
		return "", rest_errors.NewInternalServerError("invalid error interface when trying to upload file", err)
	}
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return "", rest_errors.NewInternalServerError("invalid error interface when trying to upload file", err)
	}
  bodyWriter.WriteField("store_code", p.StoreCode)
  bodyWriter.WriteField("file", string(fileData))
  bodyWriter.Close()
	response := uploadRestClient.Post("/v1/file", bodyBuf)
	if response == nil || response.Response == nil {
		return nil, rest_errors.NewInternalServerError("invalid restclient response when trying to upload file",
			errors.New("network timeout"))
	}
	if response.StatusCode > 299 {
		if strings.TrimSpace(response.String()) == "expired access token" {
			restErr := rest_errors.NewUnauthorizedError(response.String())
			return nil, restErr
		}
		restErr, err := rest_errors.NewRestErrorFromBytes(response.Bytes())
		if err != nil {
			return nil, rest_errors.NewInternalServerError("invalid error interface when trying to upload file", err)
		}
		return nil, restErr
	}

	var respBody interface{}
	if err := json.Unmarshal(response.Bytes(), &respBody); err != nil {
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal response",
			errors.New("error processing json"))
	}
	return &respBody, nil
}
