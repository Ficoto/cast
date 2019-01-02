package cast

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"os"

	"path/filepath"

	"bytes"

	"github.com/google/go-querystring/query"
)

type requestBody interface {
	ContentType() string
	Body() ([]byte, error)
}

type requestJSONBody struct {
	payload interface{}
}

func (body *requestJSONBody) ContentType() string {
	return applicaionJSON
}

func (body *requestJSONBody) Body() ([]byte, error) {
	return json.Marshal(body.payload)
}

type requestFormURLEncodedBody struct {
	payload interface{}
}

func (body *requestFormURLEncodedBody) ContentType() string {
	return formURLEncoded
}

func (body *requestFormURLEncodedBody) Body() ([]byte, error) {
	values, err := query.Values(body.payload)
	if err != nil {
		return nil, err
	}
	return []byte(values.Encode()), nil
}

type requestXMLBody struct {
	payload interface{}
}

func (body *requestXMLBody) Body() ([]byte, error) {
	return xml.Marshal(body.payload)
}

func (body *requestXMLBody) ContentType() string {
	return applicationXML
}

type requestPlainBody struct {
	payload string
}

func (body *requestPlainBody) ContentType() string {
	return textPlain
}

func (body *requestPlainBody) Body() ([]byte, error) {
	return []byte(body.payload), nil
}

// FormData represents multipart
type FormData struct {
	FieldName string
	Value     string
	FileName  string
	Path      string
	Reader    io.Reader
}

type requestMultipartFormDataBody struct {
	formData    []*FormData
	contentType string
}

func (body *requestMultipartFormDataBody) ContentType() string {
	return body.contentType
}

func (body *requestMultipartFormDataBody) Body() ([]byte, error) {
	buffer := &bytes.Buffer{}
	w := multipart.NewWriter(buffer)
	defer w.Close()

	for _, data := range body.formData {
		switch {
		case len(data.Value) != 0:
			if err := w.WriteField(data.FieldName, data.Value); err != nil {
				return nil, err
			}
		default:
			if len(data.FileName) == 0 || len(data.FileName) == 0 {
				continue
			}

			if len(data.Path) == 0 && data.Reader == nil {
				continue
			}

			fw, err := w.CreateFormFile(data.FieldName, data.FileName)
			if err != nil {
				return nil, err
			}

			switch {
			case len(data.Path) > 0:
				path, err := filepath.Abs(data.Path)
				if err != nil {
					return nil, err
				}
				f, err := os.Open(path)
				if err != nil {
					return nil, err
				}
				defer f.Close()
				_, err = io.Copy(fw, f)
				if err != nil {
					return nil, err
				}

			case data.Reader != nil:
				_, err := io.Copy(fw, data.Reader)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	body.contentType = w.FormDataContentType()
	return buffer.Bytes(), nil
}
