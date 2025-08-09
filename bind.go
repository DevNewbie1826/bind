package bind

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
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
// binderCache - A cache for Binder field indices.
// Optimizes reflection performance by caching the indices of fields that implement the Binder interface for each struct type.
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
		return err
	}
	return binder(r, reflect.ValueOf(v))
}

// binder - 재귀적 바인딩 함수
// 리플렉션을 사용하여 값(v)의 필드를 순회하고, Binder를 구현하는 필드에 대해 재귀적으로 binder를 호출합니다.
// binder - The recursive binding function.
// Uses reflection to iterate over the fields of a value (v) and recursively calls binder for fields that implement Binder.
func binder(r *http.Request, rv reflect.Value) error {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil // 포인터가 nil이면 아무 작업도 하지 않음
		}
		rv = rv.Elem()
	}

	// Binder가 아닌 값은 처리하지 않음
	if !rv.Addr().Type().Implements(binderType) {
		return nil
	}

	// 구조체가 아닐 경우 바로 Bind 실행
	if rv.Kind() != reflect.Struct {
		return rv.Addr().Interface().(Binder).Bind(r)
	}

	// 캐시에서 Binder 필드 인덱스 조회
	rt := rv.Type()
	var binderFields []int
	if cached, ok := binderCache.Load(rt); ok {
		binderFields = cached.([]int)
	} else {
		// 캐시에 없는 경우, 리플렉션으로 찾아 캐시에 저장
		for i := 0; i < rv.NumField(); i++ {
			if rt.Field(i).Type.Implements(binderType) {
				binderFields = append(binderFields, i)
			}
		}
		binderCache.Store(rt, binderFields)
	}

	// 캐시된 인덱스를 사용하여 필드 순회 및 재귀 호출
	for _, i := range binderFields {
		if err := binder(r, rv.Field(i)); err != nil {
			return err
		}
	}

	// 바텀업 순서: 하위 필드 Binder 호출 후 최종적으로 자기 자신의 Bind 호출
	return rv.Addr().Interface().(Binder).Bind(r)
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