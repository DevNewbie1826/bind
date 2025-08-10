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
	decoderMu sync.RWMutex
	decoders  = map[ContentType]func(*http.Request, any) error{
		ContentTypeJSON:      decodeJSONRequest,
		ContentTypeXML:       decodeXMLRequest,
		ContentTypeForm:      decodeFormRequest,
		ContentTypeMultipart: decodeMultipartFormRequest,
	}
	MaxMultipartMemory int64 = 32 << 20
)

// GetDecoder - 지정된 Content-Type에 대한 디코더 함수를 반환합니다.
// 테스트 또는 동적 디코더 관리에 유용합니다.
// GetDecoder returns the decoder function for the given Content-Type.
// Useful for testing or dynamic decoder management.
func GetDecoder(ct ContentType) (func(*http.Request, any) error, bool) {
	decoderMu.RLock()
	defer decoderMu.RUnlock()
	dec, ok := decoders[ct]
	return dec, ok
}

func SetMaxMultipartMemory(size int64) {
	MaxMultipartMemory = size
}

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

func RegisterDecoder(ct ContentType, fn func(*http.Request, any) error) {
	decoderMu.Lock()
	defer decoderMu.Unlock()
	decoders[ct] = fn
}

func decodeJSONRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	return json.NewDecoder(r.Body).Decode(v)
}

func decodeXMLRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	return xml.NewDecoder(r.Body).Decode(v)
}

func decodeFormRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	if err := r.ParseForm(); err != nil {
		return err
	}
	decoder := form.NewDecoder()
	return decoder.Decode(v, r.PostForm)
}

func decodeMultipartFormRequest(r *http.Request, v any) error {
	defer io.Copy(io.Discard, r.Body)
	if err := r.ParseMultipartForm(MaxMultipartMemory); err != nil {
		return err
	}
	if err := bindFiles(r, v); err != nil {
		return err
	}
	decoder := form.NewDecoder()
	return decoder.Decode(v, r.MultipartForm.Value)
}

func bindFiles(r *http.Request, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("bind: non-pointer passed to bindFiles")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil
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
