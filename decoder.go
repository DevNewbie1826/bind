package bind

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/form/v4"
)

// decoders - 디코더 레지스트리
// ContentType별로 요청을 디코딩하는 함수를 저장합니다.
// decoders - A registry for decoders.
// Stores functions that decode requests for each ContentType.
var decoders = map[ContentType]func(*http.Request, any) error{
	ContentTypeJSON:      decodeJSONRequest,
	ContentTypeXML:       decodeXMLRequest,
	ContentTypeForm:      decodeFormRequest,
	ContentTypeMultipart: decodeMultipartFormRequest,
}

// DefaultDecoder - 기본 디코더 로직
// 요청의 Content-Type 헤더를 기반으로 적절한 디코더를 찾아 실행합니다.
// 지원하지 않는 Content-Type일 경우 에러를 반환합니다.
// DefaultDecoder - The default decoder logic.
// It finds and executes the appropriate decoder based on the request's Content-Type header.
// Returns an error for unsupported content types.
func DefaultDecoder(r *http.Request, v any) error {
	ct := GetContentType(r.Header.Get("Content-Type"))
	if fn, ok := decoders[ct]; ok {
		return fn(r, v)
	}
	return errors.New("bind: unsupported content type")
}

// RegisterDecoder - 디코더 등록 함수
// 새로운 Content-Type에 대한 디코딩 함수를 동적으로 추가합니다.
// RegisterDecoder - A function to register a decoder.
// Dynamically adds a decoding function for a new Content-Type.
func RegisterDecoder(ct ContentType, fn func(*http.Request, any) error) {
	decoders[ct] = fn
}

// decodeJSONRequest - JSON 요청 디코딩 함수
// HTTP 요청의 본문을 JSON으로 디코딩합니다.
// decodeJSONRequest - Decodes a JSON request.
// Decodes the body of an HTTP request as JSON.
func decodeJSONRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	return json.NewDecoder(r.Body).Decode(v)
}

// decodeXMLRequest - XML 요청 디코딩 함수
// HTTP 요청의 본문을 XML으로 디코딩합니다.
// decodeXMLRequest - Decodes an XML request.
// Decodes the body of an HTTP request as XML.
func decodeXMLRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	return xml.NewDecoder(r.Body).Decode(v)
}

// decodeFormRequest - Form URL-encoded 요청 디코딩 함수
// HTTP 요청의 본문을 application/x-www-form-urlencoded 형식으로 디코딩합니다.
// decodeFormRequest - Decodes a form URL-encoded request.
// Decodes the body of an HTTP request as application/x-www-form-urlencoded.
func decodeFormRequest(r *http.Request, v any) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	decoder := form.NewDecoder()
	return decoder.Decode(v, r.PostForm)
}

// decodeMultipartFormRequest - Multipart Form 요청 디코딩 함수
// HTTP 요청의 본문을 multipart/form-data 형식으로 디코딩합니다.
// decodeMultipartFormRequest - Decodes a multipart form request.
// Decodes the body of an HTTP request as multipart/form-data.
func decodeMultipartFormRequest(r *http.Request, v any) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		return err
	}
	decoder := form.NewDecoder()
	return decoder.Decode(v, r.MultipartForm.Value)
}