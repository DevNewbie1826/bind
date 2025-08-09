# `bind` (English)

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

## Usage

### Example 1: Basic JSON Binding & Validation

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Bind implements the bind.Binder interface for custom validation.
func (u *User) Bind(r *http.Request) error {
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := bind.Action(r, &user); err != nil {
		// The error will be a bind.BindError, which can be inspected.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User received: %+v\n", user)
}

func main() {
	http.HandleFunc("/users", userHandler)
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Example 2: File Uploads

Binding file uploads is automatic. Just define a field of type `*multipart.FileHeader` (for a single file) or `[]*multipart.FileHeader` (for multiple files) with a `form` tag.

```go
package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

// Set a custom memory limit for multipart forms (e.g., 64 MB).
// It's good practice to set this during application startup.
func init() {
	bind.SetMaxMultipartMemory(64 << 20)
}

type AvatarUpload struct {
	UserID string                `form:"userId"`
	Avatar *multipart.FileHeader `form:"avatar"`
}

func (p *AvatarUpload) Bind(r *http.Request) error {
	if p.UserID == "" {
		return fmt.Errorf("userId is required")
	}
	if p.Avatar == nil {
		return fmt.Errorf("avatar file is required")
	}
	return nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	var payload AvatarUpload
	if err := bind.Action(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Uploaded %s for user %s (Size: %d bytes)\n", payload.Avatar.Filename, payload.UserID, payload.Avatar.Size)
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Test it with `curl`:**

```sh
curl -X POST http://localhost:8080/upload \
  -F "userId=123" \
  -F "avatar=@/path/to/your/image.jpg"
```

---

# `bind` (한국어)

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

## 사용법

### 예제 1: 기본 JSON 바인딩 및 유효성 검사

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Bind - 커스텀 유효성 검사를 위해 bind.Binder 인터페이스를 구현합니다.
func (u *User) Bind(r *http.Request) error {
	if u.Name == "" {
		return fmt.Errorf("이름은 필수 항목입니다")
	}
	return nil
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := bind.Action(r, &user); err != nil {
		// 반환된 에러는 bind.BindError 타입이며, 상세 내용을 확인할 수 있습니다.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "수신된 사용자: %+v\n", user)
}

func main() {
	http.HandleFunc("/users", userHandler)
	log.Println("서버 시작 중 :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 예제 2: 파일 업로드

파일 업로드 바인딩은 자동으로 처리됩니다. `form` 태그와 함께 `*multipart.FileHeader`(단일 파일) 또는 `[]*multipart.FileHeader`(다중 파일) 타입의 필드를 선언하기만 하면 됩니다.

```go
package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

// init 함수에서 멀티파트 폼을 위한 커스텀 메모리 제한(예: 64MB)을 설정합니다.
// 애플리케이션 시작 시에 설정하는 것이 좋습니다.
func init() {
	bind.SetMaxMultipartMemory(64 << 20)
}

type AvatarUpload struct {
	UserID string                `form:"userId"`
	Avatar *multipart.FileHeader `form:"avatar"`
}

func (p *AvatarUpload) Bind(r *http.Request) error {
	if p.UserID == "" {
		return fmt.Errorf("userId는 필수입니다")
	}
	if p.Avatar == nil {
		return fmt.Errorf("avatar 파일은 필수입니다")
	}
	return nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	var payload AvatarUpload
	if err := bind.Action(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%%s 사용자를 위해 %%s 파일 업로드됨 (크기: %%d 바이트)\n", payload.UserID, payload.Avatar.Filename, payload.Avatar.Size)
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	log.Println("서버 시작 중 :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**`curl`로 테스트하기:**

```sh
curl -X POST http://localhost:8080/upload \
  -F "userId=123" \
  -F "avatar=@/path/to/your/image.jpg"
```

---