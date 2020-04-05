package rest_upload

import (
	"strings"
	"encoding/json"
  "mime/multipart"
	"github.com/sampila/go-utils/rest_errors"
	"errors"
  "bytes"
  "io/ioutil"
	"net/http"
)

/*const (
	headerXPublic   = "X-Public"
	headerXClientId = "X-Client-Id"
	headerXCallerId = "X-Caller-Id"

	paramAccessToken = "access_token"
)*/

type FileRequest struct {
  StoreCode string  `json:"store_code"`
  Type 			string  `json:"type"`
  File *multipart.FileHeader `json:"file"`
}

func UploadFile(p *FileRequest) (*map[string]interface{}, rest_errors.RestErr) {
  file, err := p.File.Open()
  defer file.Close()
  if err != nil {
		return nil, rest_errors.NewInternalServerError("invalid error interface when trying to upload file", err)
	}
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, rest_errors.NewInternalServerError("invalid error interface when trying to upload file", err)
	}

	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)
	fw, _ := bodyWriter.CreateFormFile("file", p.File.Filename)
	fw.Write(fileData)
  bodyWriter.WriteField("store_code", p.StoreCode)
  bodyWriter.WriteField("type", p.Type)
  bodyWriter.Close()
	request, err := http.NewRequest("POST", "http://localhost:9010/v1/file", bodyBuf)
	if err != nil {
		return nil, rest_errors.NewInternalServerError("error on assign request parameter", err)
	}
	request.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	client := &http.Client{}
	response, respErr := client.Do(request)
	if respErr != nil {
		return nil, rest_errors.NewInternalServerError("invalid response",
			errors.New("network timeout"))
	}
	defer response.Body.Close()
	bodyContent, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode > 299 {
		if strings.TrimSpace(string(bodyContent)) == "expired access token" {
			restErr := rest_errors.NewUnauthorizedError(string(bodyContent))
			return nil, restErr
		}
		restErr, err := rest_errors.NewRestErrorFromBytes(bodyContent)
		if err != nil {
			return nil, rest_errors.NewInternalServerError("invalid error interface when trying to upload file", err)
		}
		return nil, restErr
	}
	var respBody map[string]interface{}
	if err := json.Unmarshal(bodyContent, &respBody); err != nil {
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal response",
			errors.New("error processing json"))
	}
	return &respBody, nil
}
