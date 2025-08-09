package bind

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

const (
	// maxRecursionDepth - 바인딩 시 최대 재귀 깊이
	// 무한 재귀로 인한 스택 오버플로우를 방지합니다.
	// maxRecursionDepth - The maximum recursion depth for binding.
	// Prevents stack overflow from infinite recursion.
	maxRecursionDepth = 1000
)

// Binder - 바인딩 인터페이스
// 구조체 또는 필드가 요청(r)을 기반으로 추가적인 바인딩 로직을 수행할 수 있도록 합니다.
// Binder - The binding interface.
// Allows a struct or field to perform additional binding logic based on the request (r).
type Binder interface {
	Bind(r *http.Request) error
}

// binderType - Binder 인터페이스의 reflect.Type
// 리플렉션을 통해 타입이 Binder 인터페이스를 구현하는지 확인하는 데 사용됩니다.
// binderType - The reflect.Type of the Binder interface.
// Used via reflection to check if a type implements the Binder interface.
var binderType = reflect.TypeOf(new(Binder)).Elem()

// binderCache - Binder 필드 인덱스 캐시
// 구조체 타입별로 Binder 인터페이스를 구현하는 필드의 인덱스를 캐싱하여 리플렉션 성능을 최적화합니다.
// sync.Map은 이러한 "write-once, read-many" 시나리오에 적합합니다.
// binderCache - A cache for Binder field indices.
// Optimizes reflection performance by caching the indices of fields that implement the Binder interface for each struct type.
// sync.Map is suitable for such "write-once, read-many" scenarios.
var binderCache = &sync.Map{}

// Action - 요청 바인딩 실행 함수
// 1. 등록된 디코더를 사용하여 요청 본문을 'v'에 디코딩합니다.
// 2. 'v' 내부의 모든 Binder 필드를 재귀적으로 바인딩합니다. (바텀업 순서)
// 3. 마지막으로 'v' 자체의 Bind 메서드를 호출합니다.
// Action - Executes the request binding.
// 1. Decodes the request body into 'v' using the registered decoder.
// 2. Recursively binds all Binder fields within 'v' (in bottom-up order).
// 3. Finally, calls the Bind method on 'v' itself.
func Action(r *http.Request, v Binder) error {
	if err := getDecode()(r, v); err != nil {
		return BindError{Err: err}
	}
	// 최상위 호출이므로 parentField는 비워두고, depth는 0에서 시작합니다.
	return binder(r, reflect.ValueOf(v), "", 0)
}

// binder - 재귀적 바인딩 함수 (필드 경로 및 깊이 추적 기능 추가)
// Bind 호출 순서:
// 1. 가장 깊은 중첩 수준의 필드부터 시작 (바텀업)
// 2. 점차 상위 레벨로 이동
// 3. 최종적으로 루트 구조체의 Bind 메서드 호출
// binder - A recursive binding function (with field path and depth tracking).
// Bind call order:
// 1. Starts from the most deeply nested fields (bottom-up).
// 2. Gradually moves to higher levels.
// 3. Finally, calls the Bind method of the root struct.
func binder(r *http.Request, rv reflect.Value, parentField string, depth int) error {
	if depth > maxRecursionDepth {
		return fmt.Errorf("max recursion depth (%d) exceeded", maxRecursionDepth)
	}

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	if !rv.Addr().Type().Implements(binderType) {
		return nil
	}

	if rv.Kind() != reflect.Struct {
		if err := rv.Addr().Interface().(Binder).Bind(r); err != nil {
			return BindError{Field: parentField, Err: err}
		}
		return nil
	}

	rt := rv.Type()
	var binderFields []int
	if cached, ok := binderCache.Load(rt); ok {
		binderFields = cached.([]int)
	} else {
		for i := 0; i < rv.NumField(); i++ {
			// 임베디드 구조체도 처리하기 위해 rt.Field(i)를 사용
			if rt.Field(i).Type.Implements(binderType) {
				binderFields = append(binderFields, i)
			}
		}
		binderCache.Store(rt, binderFields)
	}

	for _, i := range binderFields {
		field := rv.Field(i)
		fieldName := rt.Field(i).Name

		fullPath := fieldName
		if parentField != "" {
			fullPath = parentField + "." + fieldName
		}

		if err := binder(r, field, fullPath, depth+1); err != nil {
			var bindErr BindError
			if errors.As(err, &bindErr) {
				return err // 이미 BindError이므로 그대로 반환
			}
			// 새로운 에러인 경우에만 필드 정보를 추가하여 래핑
			return BindError{Field: fullPath, Err: err}
		}
	}

	if err := rv.Addr().Interface().(Binder).Bind(r); err != nil {
		return BindError{Field: parentField, Err: err}
	}
	return nil
}

// ErrorToJSON - 에러를 JSON 형식으로 변환
// 에러 메시지를 {"error":"..."} 형식의 JSON 바이트 슬라이스로 변환합니다. src가 nil이면 빈 에러 메시지를 반환합니다.
// ErrorToJSON - Converts an error to JSON format.
// Converts an error message to a JSON byte slice of the form {"error":"..."}. Returns an empty error message if src is nil.
func ErrorToJSON(src error) []byte {
	if src == nil {
		return []byte(`{"error":""}`)
	}
	b, err := json.Marshal(map[string]string{"error": src.Error()})
	if err != nil {
		// 이 경우는 거의 발생하지 않지만, 발생 시를 대비한 폴백
		return []byte(`{"error":"failed to marshal error message"}`)
	}
	return b
}

// ErrorToMap - 에러를 맵 형식으로 변환
// 에러 메시지를 map[string]string{"error":"..."} 형식으로 변환합니다. src가 nil이면 빈 에러 메시지를 반환합니다.
// ErrorToMap - Converts an error to a map format.
// Converts an error message to map[string]string{"error":"..."}. Returns an empty error message if src is nil.
func ErrorToMap(src error) map[string]string {
	if src == nil {
		return map[string]string{"error": ""}
	}
	return map[string]string{"error": src.Error()}
}

// decodeMu, decodeFn - 전역 디코더 함수와 뮤텍스
// SetDecode를 통해 커스텀 디코더 함수를 전역적으로 설정할 때 레이스 컨디션을 방지합니다.
// decodeMu, decodeFn - Global decoder function and its mutex.
// Prevents race conditions when setting a custom global decoder function via SetDecode.
var (
	decodeMu sync.RWMutex
	decodeFn = DefaultDecoder
)

// getDecode - 현재 설정된 디코더 함수를 안전하게 반환
// getDecode - Safely returns the currently configured decoder function.
func getDecode() func(*http.Request, any) error {
	decodeMu.RLock()
	defer decodeMu.RUnlock()
	return decodeFn
}

// SetDecode - 전역 디코더 함수를 안전하게 설정
// SetDecode - Safely sets the global decoder function.
func SetDecode(fn func(*http.Request, any) error) {
	decodeMu.Lock()
	defer decodeMu.Unlock()
	decodeFn = fn
}

// BindError - 표준 바인딩 에러 구조체
// 바인딩 실패 시 어떤 필드에서 에러가 발생했는지에 대한 추가 정보를 포함할 수 있습니다.
// BindError - A standard binding error struct.
// Can include additional information about which field caused the binding failure.
type BindError struct {
	Field string
	Err   error
}

func (e BindError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("bind failed on field '%s': %v", e.Field, e.Err)
	}
	return fmt.Sprintf("bind failed: %v", e.Err)
}

func (e BindError) Unwrap() error { return e.Err }