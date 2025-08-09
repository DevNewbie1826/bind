package bind_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/DevNewbie1826/bind"
)

// TestPayload - JSON/XML 바인딩 테스트를 위한 페이로드 구조체
// TestPayload - A payload struct for JSON/XML binding tests.
type TestPayload struct {
	Name  string `json:"name" xml:"name"`
	Value int    `json:"value" xml:"value"`
}

// Bind - TestPayload가 Binder 인터페이스를 구현하도록 합니다.
// Bind - Implements the Binder interface for TestPayload.
func (p *TestPayload) Bind(r *http.Request) error {
	// no-op
	return nil
}

// FormPayload - Form/Multipart 바인딩 테스트를 위한 페이로드 구조체
// FormPayload - A payload struct for Form/Multipart binding tests.
type FormPayload struct {
	Name  string `form:"name"`
	Value int    `form:"value"`
}

// Bind - FormPayload가 Binder 인터페이스를 구현하도록 합니다.
// Bind - Implements the Binder interface for FormPayload.
func (p *FormPayload) Bind(r *http.Request) error {
	// no-op
	return nil
}

// NestedPayload - 중첩 바인딩 테스트를 위한 구조체
// NestedPayload - A struct for testing nested binding.
type NestedPayload struct {
	OuterField string       `json:"outer_field"`
	Inner      *TestPayload `json:"inner"`
}

// Bind - NestedPayload가 Binder 인터페이스를 구현하도록 합니다.
// Bind - Implements the Binder interface for NestedPayload.
func (p *NestedPayload) Bind(r *http.Request) error {
	// no-op
	return nil
}

// ParentBinder - Nil 필드 바인딩 테스트를 위한 구조체
// ParentBinder - A struct for testing binding with a nil binder field.
type ParentBinder struct {
	Child *TestPayload `json:"child"`
}

// Bind - ParentBinder가 Binder 인터페이스를 구현하도록 합니다.
// Bind - Implements the Binder interface for ParentBinder.
func (pb *ParentBinder) Bind(r *http.Request) error {
	return nil
}

// TestAction_JSONBinding - 정상적인 JSON 요청에 대한 바인딩 성공을 테스트합니다.
// TestAction_JSONBinding - Tests successful binding of a valid JSON request.
func TestAction_JSONBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"name":"test", "value":42}`))
	req.Header.Set("Content-Type", "application/json")

	payload := &TestPayload{}
	err := bind.Action(req, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Name != "test" || payload.Value != 42 {
		t.Errorf(`expected {"test", 42}, got {"%s", %d}`, payload.Name, payload.Value)
	}
}

// TestAction_XMLBinding - 정상적인 XML 요청에 대한 바인딩 성공을 테스트합니다.
// TestAction_XMLBinding - Tests successful binding of a valid XML request.
func TestAction_XMLBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`<TestPayload><name>test</name><value>42</value></TestPayload>`))
	req.Header.Set("Content-Type", "application/xml")

	payload := &TestPayload{}
	err := bind.Action(req, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Name != "test" || payload.Value != 42 {
		t.Errorf(`expected {"test", 42}, got {"%s", %d}`, payload.Name, payload.Value)
	}
}

// TestAction_FormBinding - 정상적인 Form URL-encoded 요청에 대한 바인딩 성공을 테스트합니다.
// TestAction_FormBinding - Tests successful binding of a valid form URL-encoded request.
func TestAction_FormBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("name=test&value=42"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	payload := &FormPayload{}
	err := bind.Action(req, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Name != "test" || payload.Value != 42 {
		t.Errorf(`expected {"test", 42}, got {"%s", %d}`, payload.Name, payload.Value)
	}
}

// TestAction_NestedBinding - 중첩된 구조체에 대한 재귀적 바인딩이 올바르게 동작하는지 테스트합니다.
// TestAction_NestedBinding - Tests that recursive binding works correctly for nested structs.
func TestAction_NestedBinding(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"outer_field":"outer", "inner":{"name":"inner_test", "value":123}}`))
	req.Header.Set("Content-Type", "application/json")

	payload := &NestedPayload{Inner: &TestPayload{}}
	err := bind.Action(req, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.OuterField != "outer" {
		t.Errorf("expected outer field to be 'outer', got '%s'", payload.OuterField)
	}
	if payload.Inner.Name != "inner_test" || payload.Inner.Value != 123 {
		t.Errorf(`expected inner {"inner_test", 123}, got {"%s", %d}`, payload.Inner.Name, payload.Inner.Value)
	}
}

// TestAction_UnsupportedContentType - 지원하지 않는 Content-Type에 대해 에러가 반환되는지 테스트합니다.
// TestAction_UnsupportedContentType - Tests that an error is returned for an unsupported content type.
func TestAction_UnsupportedContentType(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("data"))
	req.Header.Set("Content-Type", "application/octet-stream")

	payload := &TestPayload{}
	err := bind.Action(req, payload)
	if err == nil {
		t.Error("expected error for unsupported content type, got nil")
	}
}

// TestAction_InvalidJSON - 잘못된 형식의 JSON에 대해 디코딩 에러가 발생하는지 테스트합니다.
// TestAction_InvalidJSON - Tests that a decode error occurs for malformed JSON.
func TestAction_InvalidJSON(t *testing.T) {
	badJSON := `{"name": "abc", "value":}`
	req, _ := http.NewRequest("POST", "/", strings.NewReader(badJSON))
	req.Header.Set("Content-Type", "application/json")

	payload := &TestPayload{}
	err := bind.Action(req, payload)
	if err == nil {
		t.Error("expected JSON decode error, got nil")
	}
}

// TestAction_NilBinderField - 중첩된 Binder 필드가 nil일 경우에도 바인딩이 패닉 없이 성공하는지 테스트합니다.
// TestAction_NilBinderField - Tests that binding succeeds without a panic even if a nested binder field is nil.
func TestAction_NilBinderField(t *testing.T) {
	// 요청 본문에서 child가 null이므로, 디코딩 후 ParentBinder.Child는 nil이 됩니다.
	// Action 함수는 nil 필드에 대해서는 Bind를 호출하지 않고 건너뛰어야 합니다.
	// In the request body, 'child' is null, so ParentBinder.Child will be nil after decoding.
	// The Action function should skip calling Bind on nil fields.
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"child":null}`))
	req.Header.Set("Content-Type", "application/json")

	p := &ParentBinder{}
	err := bind.Action(req, p)
	if err != nil {
		t.Errorf("unexpected error when binding to a struct with a nil binder field: %v", err)
	}
}

// TestAction_MultipartForm - 정상적인 multipart/form-data 요청에 대한 바인딩 성공을 테스트합니다.
// TestAction_MultipartForm - Tests successful binding of a valid multipart/form-data request.
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

	payload := &FormPayload{}
	err := bind.Action(req, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if payload.Name != "multi" || payload.Value != 123 {
		t.Errorf("multipart binding failed, got %+v", payload)
	}
}