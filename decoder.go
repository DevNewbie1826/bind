package bind

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-playground/form/v4"
)

var (
	// decoderMu - 디코더 맵 접근을 위한 뮤텍스
	// decoderMu - A mutex for accessing the decoders map.
	decoderMu sync.RWMutex

	// decoders - 디코더 레지스트리
	// decoders - A registry for decoders.
	decoders = map[ContentType]func(*http.Request, any) error{
		ContentTypeJSON:      decodeJSONRequest,
		ContentTypeXML:       decodeXMLRequest,
		ContentTypeForm:      decodeFormRequest,
		ContentTypeMultipart: decodeMultipartFormRequest,
	}

	// MaxMultipartMemory - 멀티파트 폼 파싱 시 최대 메모리 사용량 (기본값: 32MB)
	// 이 값은 SetMaxMultipartMemory 함수를 통해 런타임에 변경할 수 있습니다.
	// MaxMultipartMemory - The maximum memory to use when parsing multipart forms (default: 32MB).
	// This value can be changed at runtime via the SetMaxMultipartMemory function.
	MaxMultipartMemory int64 = 32 << 20
)

// SetMaxMultipartMemory - 멀티파트 폼 파싱 시 사용할 최대 메모리를 설정합니다.
// 이 함수는 서버가 요청을 처리하기 전에 호출하는 것이 좋습니다.
// SetMaxMultipartMemory sets the maximum memory (in bytes) to use when parsing multipart forms.
// It is recommended to call this function before processing requests.
func SetMaxMultipartMemory(size int64) {
	MaxMultipartMemory = size
}

// DefaultDecoder - 기본 디코더 로직
// 요청의 Content-Type 헤더를 기반으로 적절한 디코더를 찾아 실행합니다.
// 지원하지 않는 Content-Type일 경우 에러를 반환합니다.
// DefaultDecoder - The default decoder logic.
// It finds and executes the appropriate decoder based on the request's Content-Type header.
// Returns an error for unsupported content types.
func DefaultDecoder(r *http.Request, v any) error {
	ct := GetContentType(r.Header.Get("Content-Type"))

	decoderMu.RLock()
	fn, ok := decoders[ct]
	decoderMu.RUnlock()

	if ok {
		return fn(r, v)
	}
	return errors.New("bind: unsupported content type")
}

// RegisterDecoder - 디코더 등록 함수
// 새로운 Content-Type에 대한 디코딩 함수를 동적으로 추가합니다.
// RegisterDecoder - A function to register a decoder.
// Dynamically adds a decoding function for a new Content-Type.
func RegisterDecoder(ct ContentType, fn func(*http.Request, any) error) {
	decoderMu.Lock()
	defer decoderMu.Unlock()
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
	defer io.Copy(io.Discard, r.Body)
	if err := r.ParseForm(); err != nil {
		return err
	}
	decoder := form.NewDecoder()
	return decoder.Decode(v, r.PostForm)
}

// decodeMultipartFormRequest - Multipart Form 요청 디코딩 함수
// HTTP 요청의 본문을 multipart/form-data 형식으로 디코딩하고, 파일 필드를 바인딩합니다.
// decodeMultipartFormRequest - Decodes a multipart form request.
// Decodes the body of an HTTP request as multipart/form-data and binds file fields.
func decodeMultipartFormRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	if err := r.ParseMultipartForm(MaxMultipartMemory); err != nil {
		return err
	}

	// 리플렉션을 사용하여 파일 필드를 수동으로 바인딩
	if err := bindFiles(r, v); err != nil {
		return err
	}

	decoder := form.NewDecoder()
	return decoder.Decode(v, r.MultipartForm.Value)
}

// bindFiles - 리플렉션을 사용하여 multipart 파일들을 구조체 필드에 바인딩합니다.
func bindFiles(r *http.Request, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("bind: non-pointer passed to bindFiles")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil // 구조체가 아니면 파일 바인딩 스킵
	}
	rt := rv.Type()

	fileHeaderPtrType := reflect.TypeOf((*multipart.FileHeader)(nil))
	fileHeaderSliceType := reflect.TypeOf(([]*multipart.FileHeader)(nil))

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)
		formTag := fieldType.Tag.Get("form")
		if formTag == "" {
			continue
		}

		files, ok := r.MultipartForm.File[formTag]
		if !ok || len(files) == 0 {
			continue
		}

		if !field.CanSet() {
			continue
		}

		switch field.Type() {
		case fileHeaderPtrType:
			field.Set(reflect.ValueOf(files[0]))
		case fileHeaderSliceType:
			field.Set(reflect.ValueOf(files))
		}
	}
	return nil
}
