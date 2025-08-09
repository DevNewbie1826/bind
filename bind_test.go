package bind_test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/DevNewbie1826/bind"
)

// --- 테스트용 구조체 정의 ---

type TestPayload struct {
	Name  string `json:"name" xml:"name" form:"name"`
	Value int    `json:"value" xml:"value" form:"value"`
}

func (p *TestPayload) Bind(r *http.Request) error { return nil }

type FileUploadPayload struct {
	Name  string                  `form:"name"`
	File  *multipart.FileHeader   `form:"file"`
	Files []*multipart.FileHeader `form:"files"`
}

func (p *FileUploadPayload) Bind(r *http.Request) error { return nil }

type NestedPayload struct {
	OuterField string       `json:"outer_field"`
	Inner      *TestPayload `json:"inner"`
}

func (p *NestedPayload) Bind(r *http.Request) error { return nil }

type ParentBinder struct {
	Child *TestPayload `json:"child"`
}

func (pb *ParentBinder) Bind(r *http.Request) error { return nil }

type InnerBinder struct{}

func (b *InnerBinder) Bind(r *http.Request) error { return errors.New("inner error") }

type MiddleBinder struct {
	Inner *InnerBinder `json:"inner"`
}

func (b *MiddleBinder) Bind(r *http.Request) error { return nil }

type OuterBinder struct {
	Middle *MiddleBinder `json:"middle"`
}

func (b *OuterBinder) Bind(r *http.Request) error { return nil }

type DeepBinder struct {
	Child *DeepBinder `json:"child"`
}

func (b *DeepBinder) Bind(r *http.Request) error { return nil }

type EmbeddedPayload struct {
	TestPayload
	Extra string `json:"extra"`
}

func (p *EmbeddedPayload) Bind(r *http.Request) error { return nil }

type UnexportedFieldPayload struct {
	unexportedBinder *TestPayload `form:"unexported"`
	Exported         string       `form:"exported"`
}

func (p *UnexportedFieldPayload) Bind(r *http.Request) error { return nil }

// --- 테스트 함수 ---

func TestAction_JSONBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"name":"test", "value":42}`))
	req.Header.Set("Content-Type", "application/json")
	payload := &TestPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Name != "test" || payload.Value != 42 {
		t.Errorf(`expected {"test", 42}, got {"%s", %d}`, payload.Name, payload.Value)
	}
}

func TestAction_XMLBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`<TestPayload><name>test</name><value>42</value></TestPayload>`))
	req.Header.Set("Content-Type", "application/xml")
	payload := &TestPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Name != "test" || payload.Value != 42 {
		t.Errorf(`expected {"test", 42}, got {"%s", %d}`, payload.Name, payload.Value)
	}
}

func TestAction_FormBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("name=test&value=42"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	payload := &TestPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Name != "test" || payload.Value != 42 {
		t.Errorf(`expected {"test", 42}, got {"%s", %d}`, payload.Name, payload.Value)
	}
}

func TestAction_NestedBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"outer_field":"outer", "inner":{"name":"inner_test", "value":123}}`))
	req.Header.Set("Content-Type", "application/json")
	payload := &NestedPayload{Inner: &TestPayload{}}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.OuterField != "outer" || payload.Inner.Name != "inner_test" || payload.Inner.Value != 123 {
		t.Errorf("nested binding failed, got %+v", payload)
	}
}

func TestAction_UnsupportedContentType(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("data"))
	req.Header.Set("Content-Type", "application/octet-stream")
	err := bind.Action(req, &TestPayload{})
	if err == nil {
		t.Error("expected error for unsupported content type, got nil")
	}
}

func TestAction_InvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"name": "abc", "value":}`))
	req.Header.Set("Content-Type", "application/json")
	err := bind.Action(req, &TestPayload{})
	if err == nil {
		t.Error("expected JSON decode error, got nil")
	}
}

func TestAction_NilBinderField(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"child":null}`))
	req.Header.Set("Content-Type", "application/json")
	if err := bind.Action(req, &ParentBinder{}); err != nil {
		t.Errorf("unexpected error with nil binder field: %v", err)
	}
}

func TestAction_MultipartForm(t *testing.T) {
	body := new(bytes.Buffer)
	body.WriteString("--BOUNDARY\r\n")
	body.WriteString(`Content-Disposition: form-data; name="name"` + "\r\n\r\n")
	body.WriteString("multi\r\n")
	body.WriteString("--BOUNDARY\r\n")
	body.WriteString(`Content-Disposition: form-data; name="value"` + "\r\n\r\n")
	body.WriteString("123\r\n")
	body.WriteString("--BOUNDARY--\r\n")
	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=BOUNDARY")
	payload := &TestPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if payload.Name != "multi" || payload.Value != 123 {
		t.Errorf("multipart binding failed, got %+v", payload)
	}
}

func TestAction_MultipartFileUpload(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test file"))
	writer.WriteField("name", "file-test")
	writer.Close()
	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	payload := &FileUploadPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("Action failed with file upload: %v", err)
	}
	if payload.Name != "file-test" || payload.File == nil || payload.File.Filename != "test.txt" {
		t.Errorf("single file upload binding failed, got %+v", payload)
	}
}

func TestAction_MultiFileUpload(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part1, _ := writer.CreateFormFile("files", "test1.txt")
	part1.Write([]byte("file1"))
	part2, _ := writer.CreateFormFile("files", "test2.txt")
	part2.Write([]byte("file2"))
	writer.Close()
	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	payload := &FileUploadPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("Action failed with multi-file upload: %v", err)
	}
	if len(payload.Files) != 2 || payload.Files[0].Filename != "test1.txt" || payload.Files[1].Filename != "test2.txt" {
		t.Error("multi-file upload binding failed")
	}
}

func TestAction_NestedErrorPropagation(t *testing.T) {
	payload := &OuterBinder{Middle: &MiddleBinder{Inner: &InnerBinder{}}}
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"middle":{"inner":{}}}`))
	req.Header.Set("Content-Type", "application/json")
	err := bind.Action(req, payload)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	expected := "bind failed on field 'Middle.Inner': inner error"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestAction_RecursionDepthLimit(t *testing.T) {
	jsonBody := strings.Repeat(`{"child":`, 1001) + "null" + strings.Repeat("}", 1001)
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	err := bind.Action(req, &DeepBinder{})
	if err == nil || !strings.Contains(err.Error(), "max recursion depth (1000) exceeded") {
		t.Errorf("Expected recursion depth error, got: %v", err)
	}
}

func TestAction_EmbeddedStruct(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"name":"embedded", "value":99, "extra":"field"}`))
	req.Header.Set("Content-Type", "application/json")
	payload := &EmbeddedPayload{}
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("unexpected error with embedded struct: %v", err)
	}
	if payload.Name != "embedded" || payload.Value != 99 || payload.Extra != "field" {
		t.Errorf("embedded struct binding failed, got %+v", payload)
	}
}

func TestErrorToJSON_Nil(t *testing.T) {
	result := bind.ErrorToJSON(nil)
	if string(result) != `{"error":""}` {
		t.Errorf(`Expected '{"error":""}', got '%s'`, string(result))
	}
}

func TestErrorToMap_Nil(t *testing.T) {
	result := bind.ErrorToMap(nil)
	if val, ok := result["error"]; !ok || val != "" {
		t.Errorf(`Expected map[error:""]', got '%v'`, result)
	}
}

func TestCustomDecoderRegistration(t *testing.T) {
	originalDecoder := bind.DefaultDecoder
	originalJSONDecoder, ok := bind.GetDecoder(bind.ContentTypeJSON)
	if !ok {
		t.Fatal("failed to get original JSON decoder")
	}

	t.Cleanup(func() {
		bind.SetDecode(originalDecoder)
		bind.RegisterDecoder(bind.ContentTypeJSON, originalJSONDecoder)
	})

	customErr := errors.New("custom decoder error")
	bind.RegisterDecoder(bind.ContentTypeJSON, func(r *http.Request, v any) error {
		return customErr
	})

	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{ }`))
	req.Header.Set("Content-Type", "application/json")
	err := bind.Action(req, &TestPayload{})

	if err == nil {
		t.Fatal("Expected custom decoder error, but got nil")
	}
	var bindErr bind.BindError
	if !errors.As(err, &bindErr) || bindErr.Unwrap() != customErr {
		t.Errorf("Expected error to wrap '%v', but got '%v'", customErr, err)
	}
}

// --- 추가된 테스트 케이스 ---

func TestGetContentType(t *testing.T) {
	testCases := []struct {
		input    string
		expected bind.ContentType
	}{
		{"text/plain; charset=utf-8", bind.ContentTypePlainText},
		{"application/json", bind.ContentTypeJSON},
		{"application/problem+json", bind.ContentTypeJSON},
		{"text/xml; charset=utf-8", bind.ContentTypeXML},
		{"application/x-www-form-urlencoded", bind.ContentTypeForm},
		{"multipart/form-data; boundary=...", bind.ContentTypeMultipart},
		{"text/html", bind.ContentTypeHTML},
		{"text/event-stream", bind.ContentTypeEventStream},
		{"application/unknown", bind.ContentTypeUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			if got := bind.GetContentType(tc.input); got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestAction_MalformedMultipartForm(t *testing.T) {
	// A body that is missing the final boundary
	body := new(bytes.Buffer)
	body.WriteString("--BOUNDARY\r\n")
	body.WriteString(`Content-Disposition: form-data; name="name"` + "\r\n\r\n")
	body.WriteString("multi\r\n")

	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=BOUNDARY")

	payload := &TestPayload{}
	err := bind.Action(req, payload)
	if err == nil {
		t.Fatal("expected error for malformed multipart form, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected EOF") {
		t.Errorf("expected multipart EOF error, got %v", err)
	}
}

func TestAction_UnexportedField(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("exported", "value")
	writer.Close()

	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	payload := &UnexportedFieldPayload{}
	// This should not panic and should bind the exported field.
	if err := bind.Action(req, payload); err != nil {
		t.Fatalf("unexpected error with unexported field: %v", err)
	}
	if payload.Exported != "value" {
		t.Errorf("expected exported field to be 'value', got '%s'", payload.Exported)
	}
}

// TestConcurrentBinding - `go test -race`를 통해 캐시의 동시성 안전성을 검증합니다.
func TestConcurrentBinding(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"name":"test", "value":42}`))
			req.Header.Set("Content-Type", "application/json")
			payload := &TestPayload{}
			if err := bind.Action(req, payload); err != nil {
				t.Errorf("concurrent binding failed: %v", err)
			}
		}()
	}
	wg.Wait()
}
