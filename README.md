# `bind` (English)

[![Go Report Card](https://goreportcard.com/badge/github.com/DevNewbie1826/bind)](https://goreportcard.com/report/github.com/DevNewbie1826/bind)
[![Go Reference](https://pkg.go.dev/badge/github.com/DevNewbie1826/bind.svg)](https://pkg.go.dev/github.com/DevNewbie1826/bind)

Inspired by [go-chi/render](https://github.com/go-chi/render).

A simple, powerful, and highly performant Go package for binding request data (JSON, XML, Form) to structs. It features recursive binding, detailed error pathing, file upload support, and is optimized for high-performance scenarios.

## Features

- **Multiple Content-Types:** Natively supports `application/json`, `application/xml`, `application/x-www-form-urlencoded`, and `multipart/form-data`.
- **Recursive Binding:** Automatically calls the `Bind` method on nested fields that implement the `Binder` interface. The binding order is bottom-up, from the innermost field to the outermost struct.
- **File Uploads:** Natively binds single (`*multipart.FileHeader`) and multiple (`[]*multipart.FileHeader`) file uploads from `multipart/form-data` requests.
- **Configurable Memory:** The maximum memory for multipart form parsing can be easily configured via `bind.SetMaxMultipartMemory()`.
- **Detailed Error Reporting:** Errors are wrapped in a `BindError` type that includes the full field path (e.g., `Parent.Child.Field`), making debugging significantly easier.
- **Security:** Includes a configurable recursion depth limit to prevent stack overflow attacks from malicious or malformed requests.
- **Extensible:** Easily register new decoders for custom content types.
- **Performance-Optimized:** Caches struct analysis results using `sync.Map` to minimize reflection overhead in hot paths, making it suitable for high-traffic services.

## Installation

```sh
go get github.com/DevNewbie1826/bind
```

## Basic Usage

See the examples below for basic JSON and file upload binding.

---

## Advanced Usage

### 1. Custom Decoders

You can extend `bind` to support custom content types by registering a new decoder.

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/DevNewbie1826/bind"
)

// 1. Define your custom Content-Type (if not already standard)
const MyCustomContentType bind.ContentType = 100

// 2. Create a custom decoder function
func decodeYAML(r *http.Request, v any) error {
	// In a real implementation, you would use a YAML library.
	// This is a simplified example.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), "name: John") {
		// Simulate decoding into a struct
		if p, ok := v.(*bind.TestPayload); ok {
			p.Name = "John"
			p.Value = 30
		}
		return nil
	}
	return fmt.Errorf("could not decode YAML")
}

func main() {
	// 3. Register the new decoder
	bind.RegisterDecoder(MyCustomContentType, decodeYAML)

	// You also need a way to map the "application/x-yaml" header to your custom type.
	// This part is left to the application logic, as GetContentType can also be customized.
	
	// ... server setup
}
```

### 2. Error Handling Best Practices

`bind.Action` returns a `bind.BindError`. You can use `errors.As` to inspect it and get detailed context about the failure.

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

type Address struct {
	City string `json:"city"`
}

func (a *Address) Bind(r *http.Request) error {
	if a.City == "" {
		return fmt.Errorf("city is a required field")
	}
	return nil
}

type User struct {
	Address *Address `json:"address"`
}

func (u *User) Bind(r *http.Request) error { return nil }

func handler(w http.ResponseWriter, r *http.Request) {
	var user User
	// Important: Initialize nested pointers
	user.Address = &Address{}

	if err := bind.Action(r, &user); err != nil {
		var bindErr bind.BindError
		// Use errors.As to check if the error is a BindError
		if errors.As(err, &bindErr) {
			// Now you can access the specific field and the underlying error
			log.Printf("Binding failed on field '%s': %v", bindErr.Field, bindErr.Unwrap())
			http.Error(w, fmt.Sprintf("Bad Request in field '%s'", bindErr.Field), http.StatusBadRequest)
		} else {
			// Handle other types of errors (e.g., malformed JSON)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	fmt.Fprintf(w, "Success!")
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Test with `curl`:**
```sh
# Send a request with a missing "city" field
curl -X POST http://localhost:8080 -d '{"address":{}}' -H "Content-Type: application/json"

# Server will log:
# Binding failed on field 'Address': city is a required field

# Client will receive:
# Bad Request in field 'Address'
```

### 3. Configuring Multipart Memory

For services that handle large file uploads, it's crucial to control memory usage. You can set the maximum memory for multipart form parsing globally.

It's best to do this once during application startup using an `init` function.

```go
package main

import (
	"github.com/DevNewbie1826/bind"
	// ... other imports
)

func init() {
    // Set max memory for multipart forms to 64MB
    bind.SetMaxMultipartMemory(64 << 20)
}

// ... your application code
```

---

# `bind` (한국어)

[go-chi/render](https://github.com/go-chi/render)에서 영감을 받아 제작되었습니다.

요청 데이터(JSON, XML, Form)를 Go 구조체에 바인딩하는 간단하고, 강력하며, 성능이 뛰어난 패키지입니다. 재귀적 바인딩, 상세한 오류 경로 추적, 파일 업로드 기능을 지원하며 고성능 시나리오에 최적화되어 있습니다.

## 주요 특징

- **다양한 Content-Type 지원:** `application/json`, `application/xml`, `application/x-www-form-urlencoded`, `multipart/form-data`를 기본 지원합니다.
- **재귀적 바인딩:** `Binder` 인터페이스를 구현하는 중첩 필드의 `Bind` 메서드를 가장 안쪽(bottom-up)부터 순서대로 자동 호출합니다.
- **파일 업로드:** `multipart/form-data` 요청으로부터 단일(`*multipart.FileHeader`) 및 다중(`[]*multipart.FileHeader`) 파일 업로드를 자동으로 바인딩합니다.
- **메모리 설정 가능:** `bind.SetMaxMultipartMemory()` 함수를 통해 멀티파트 폼 파싱 시 최대 메모리를 쉽게 설정할 수 있습니다.
- **상세한 오류 리포팅:** 오류 발생 시 전체 필드 경로(예: `Parent.Child.Field`)를 포함하는 `BindError` 타입으로 래핑하여 디버깅을 크게 용이하게 합니다.
- **보안:** 설정 가능한 재귀 깊이 제한을 두어 악의적이거나 잘못된 형식의 요청으로 인한 스택 오버플로우 공격을 방지합니다.
- **확장성:** 커스텀 Content-Type을 위한 새로운 디코더를 쉽게 등록할 수 있습니다.
- **성능 최적화:** `sync.Map`을 사용하여 구조체 분석 결과를 캐싱함으로써, 트래픽이 많은 서비스에 적합하도록 리플렉션 오버헤드를 최소화합니다.

## 설치

```sh
go get github.com/DevNewbie1826/bind
```

## 기본 사용법

기본적인 JSON 및 파일 업로드 바인딩은 아래 예제를 참고하세요.

---

## 고급 사용법

### 1. 커스텀 디코더 등록

`bind`를 확장하여 커스텀 Content-Type을 지원하도록 새로운 디코더를 등록할 수 있습니다.

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/DevNewbie1826/bind"
)

// 1. 커스텀 Content-Type 정의 (표준이 아닌 경우)
const MyCustomContentType bind.ContentType = 100

// 2. 커스텀 디코더 함수 생성
func decodeYAML(r *http.Request, v any) error {
	// 실제 구현에서는 YAML 라이브러리를 사용해야 합니다.
	// 여기서는 간단한 예시입니다.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), "name: John") {
		// 구조체로 디코딩하는 것을 시뮬레이션
		if p, ok := v.(*bind.TestPayload); ok {
			p.Name = "John"
			p.Value = 30
		}
		return nil
	}
	return fmt.Errorf("YAML을 디코딩할 수 없습니다")
}

func main() {
	// 3. 새로운 디코더 등록
	bind.RegisterDecoder(MyCustomContentType, decodeYAML)

	// "application/x-yaml" 헤더를 커스텀 타입에 매핑하는 로직도 필요합니다.
	// GetContentType 또한 커스터마이징이 가능하므로, 이 부분은 애플리케이션 로직에 맡겨집니다.
	
	// ... 서버 설정
}
```

### 2. 에러 처리 베스트 프랙티스

`bind.Action`은 `bind.BindError`를 반환합니다. `errors.As`를 사용하여 에러를 검사하고 실패에 대한 상세한 컨텍스트를 얻을 수 있습니다.

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

type Address struct {
	City string `json:"city"`
}

func (a *Address) Bind(r *http.Request) error {
	if a.City == "" {
		return fmt.Errorf("city 필드는 필수입니다")
	}
	return nil
}

type User struct {
	Address *Address `json:"address"`
}

func (u *User) Bind(r *http.Request) error { return nil }

func handler(w http.ResponseWriter, r *http.Request) {
	var user User
	// 중요: 중첩된 포인터 초기화
	user.Address = &Address{}

	if err := bind.Action(r, &user); err != nil {
		var bindErr bind.BindError
		// errors.As를 사용하여 BindError인지 확인
		if errors.As(err, &bindErr) {
			// 특정 필드와 원본 에러에 접근 가능
			log.Printf("필드 '%s'에서 바인딩 실패: %v", bindErr.Field, bindErr.Unwrap())
			http.Error(w, fmt.Sprintf("필드 '%s'에 잘못된 요청", bindErr.Field), http.StatusBadRequest)
		} else {
			// 다른 종류의 에러 처리 (예: 잘못된 JSON 형식)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	fmt.Fprintf(w, "성공!")
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**`curl`로 테스트하기:**
```sh
# "city" 필드가 없는 요청 보내기
curl -X POST http://localhost:8080 -d '{"address":{}}' -H "Content-Type: application/json"

# 서버 로그:
# 필드 'Address'에서 바인딩 실패: city 필드는 필수입니다

# 클라이언트가 받는 응답:
# 필드 'Address'에 잘못된 요청
```

### 3. 멀티파트 메모리 설정

대용량 파일 업로드를 처리하는 서비스에서는 메모리 사용량을 제어하는 것이 중요합니다. 멀티파트 폼 파싱을 위한 최대 메모리를 전역적으로 설정할 수 있습니다.

애플리케이션 시작 시 `init` 함수를 사용하여 한 번만 설정하는 것이 가장 좋습니다.

```go
package main

import (
	"github.com/DevNewbie1826/bind"
	// ... 다른 임포트
)

func init() {
    // 멀티파트 폼을 위한 최대 메모리를 64MB로 설정
    bind.SetMaxMultipartMemory(64 << 20)
}

// ... 애플리케이션 코드
```
---

## License

This project is licensed under the MIT License.

---

## 라이선스

이 프로젝트는 MIT 라이선스를 따릅니다.