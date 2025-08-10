package bind

import (
	"strings"
)

// ContentType - HTTP Content-Type을 나타내는 열거형
// ContentType - An enumeration for HTTP Content-Types.
type ContentType int

const (
	// ContentTypeUnknown - 알 수 없거나 지원하지 않는 Content-Type
	// ContentTypeUnknown - An unknown or unsupported Content-Type.
	ContentTypeUnknown ContentType = iota
	// ContentTypePlainText - "text/plain"
	// ContentTypePlainText - "text/plain".
	ContentTypePlainText
	// ContentTypeHTML - "text/html"
	// ContentTypeHTML - "text/html".
	ContentTypeHTML
	// ContentTypeJSON - "application/json"
	// ContentTypeJSON - "application/json".
	ContentTypeJSON
	// ContentTypeXML - "application/xml"
	// ContentTypeXML - "application/xml".
	ContentTypeXML
	// ContentTypeForm - "application/x-www-form-urlencoded"
	// ContentTypeForm - "application/x-www-form-urlencoded".
	ContentTypeForm
	// ContentTypeMultipart - "multipart/form-data"
	// ContentTypeMultipart - "multipart/form-data".
	ContentTypeMultipart
	// ContentTypeEventStream - "text/event-stream"
	// ContentTypeEventStream - "text/event-stream".
	ContentTypeEventStream
)

// GetContentType - Content-Type 문자열을 파싱하여 ContentType 열거형 값으로 변환합니다.
// "; charset=..."과 같은 추가 파라미터는 무시합니다.
// GetContentType - Parses a Content-Type string and converts it to a ContentType enum value.
// It ignores additional parameters like "; charset=...".
func GetContentType(s string) ContentType {
	s = strings.TrimSpace(strings.Split(s, ";")[0])
	switch s {
	case "text/plain":
		return ContentTypePlainText
	case "text/html", "application/xhtml+xml":
		return ContentTypeHTML
	case "application/json", "text/javascript", "application/problem+json", "application/vnd.api+json":
		return ContentTypeJSON
	case "text/xml", "application/xml":
		return ContentTypeXML
	case "application/x-www-form-urlencoded":
		return ContentTypeForm
	case "multipart/form-data":
		return ContentTypeMultipart
	case "text/event-stream":
		return ContentTypeEventStream
	default:
		return ContentTypeUnknown
	}
}
