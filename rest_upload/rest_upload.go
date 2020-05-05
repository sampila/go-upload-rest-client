package rest_upload

import (
	"io"
	"os"
	"strings"
	"encoding/json"
  "mime/multipart"
	"github.com/sampila/go-utils/rest_errors"
	"errors"
  "bytes"
  "io/ioutil"
	"net/http"
	"time"
)

/*const (
	headerXPublic   = "X-Public"
	headerXClientId = "X-Client-Id"
	headerXCallerId = "X-Caller-Id"

	paramAccessToken = "access_token"
)*/

type UploaderFile struct{
	Url string
	Params *FileRequest
}

type FileRequest struct {
  StoreCode  string  `json:"store_code"`
  StoreName  string  `json:"store_name,omitempty"`
  TargetPath string  `json:"target_path,omitempty"`
  SourcePath string  `json:"source_path,omitempty"`
  Type 			 string  `json:"type"`
  File 			 *multipart.FileHeader `json:"file,omitempty"`
  ThemeFile  *os.File `json:"theme_file,omitempty"`
}

func NewUploader(url string, p *FileRequest) *UploaderFile{
	return &UploaderFile{Url : url, Params: p}
}

func (u *UploaderFile) Upload() (*map[string]interface{}, rest_errors.RestErr) {
	params := u.Params

	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)
	if params.File != nil {
	  file, err := params.File.Open()
	  defer file.Close()
	  if err != nil {
			return nil, rest_errors.NewInternalServerError("file error", err)
		}
		fileData, err := ioutil.ReadAll(file)
		fw, _ := bodyWriter.CreateFormFile("file", params.File.Filename)
		fw.Write(fileData)
	}

	if params.ThemeFile != nil {
		themeFileForm, err := bodyWriter.CreateFormFile("theme_file", params.ThemeFile.Name())
		if err != nil {
	    return nil, rest_errors.NewInternalServerError("error on assign request parameter", err)
		}
		_, err = io.Copy(themeFileForm, params.ThemeFile)
	}

  bodyWriter.WriteField("store_code", params.StoreCode)
  bodyWriter.WriteField("store_name", params.StoreName)
  bodyWriter.WriteField("target_path", params.TargetPath)
  bodyWriter.WriteField("source_path", params.SourcePath)
  bodyWriter.WriteField("type", params.Type)
  bodyWriter.Close()
	request, err := http.NewRequest("POST", u.Url, bodyBuf)
	if err != nil {
		return nil, rest_errors.NewInternalServerError("error on assign request parameter", err)
	}

	request.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	response, respErr := client.Do(request)
	if respErr != nil {
		return nil, rest_errors.NewInternalServerError("invalid response",
			respErr)
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

func (u *UploaderFile) UploadImg() (*map[string]interface{}, rest_errors.RestErr) {
	p := u.Params
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
	request, err := http.NewRequest("POST", u.Url, bodyBuf)
	if err != nil {
		return nil, rest_errors.NewInternalServerError("error on assign request parameter", err)
	}
	request.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	response, respErr := client.Do(request)
	if respErr != nil {
		return nil, rest_errors.NewInternalServerError("invalid response",
			respErr)
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
