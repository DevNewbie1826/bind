# `bind` (English)

A simple and powerful Go package for binding request data (JSON, XML, Form) to structs. It features recursive binding for nested structures and is optimized for performance.

## Features

- **Multiple Content-Types:** Supports `application/json`, `application/xml`, `application/x-www-form-urlencoded`, and `multipart/form-data`.
- **Recursive Binding:** Automatically calls the `Bind` method on nested fields that implement the `Binder` interface in a bottom-up order.
- **Custom Logic:** By implementing the `Binder` interface, you can add custom validation or data manipulation logic that runs after decoding.
- **Extensible:** Easily register new decoders for custom content types.
- **Performance-Optimized:** Caches struct analysis results to minimize reflection overhead in hot paths.

## Installation

```sh
go get github.com/DevNewbie1826/bind
```

## Usage

### Example 1: Basic JSON Binding & Validation

Here is a simple example of binding a JSON request to a struct and running custom validation.

```go
// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

// User defines the structure for user data.
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Bind implements the bind.Binder interface.
// This method is called after the request body is decoded into the struct.
// You can add custom validation logic here.
func (u *User) Bind(r *http.Request) error {
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := bind.Action(r, &user); err != nil {
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

**Test it with `curl`:**

```sh
# Success Case
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name":"John Doe", "email":"john.doe@example.com"}'
# Expected Output: User received: {Name:John Doe Email:john.doe@example.com}

# Failure Case (missing name triggers validation)
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"email":"john.doe@example.com"}'
# Expected Output: name is required
```

### Example 2: Nested Binding

The `bind` package automatically handles nested structs that also implement the `Binder` interface. The `Bind` methods are called recursively from the innermost field to the outermost struct.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

// Address contains location details.
type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// Bind implements the bind.Binder interface for Address.
func (a *Address) Bind(r *http.Request) error {
	if a.Country == "" {
		return fmt.Errorf("country is required")
	}
	fmt.Println("Address.Bind called")
	return nil
}

// Profile contains user details and a nested Address.
type Profile struct {
	Name    string   `json:"name"`
	Address *Address `json:"address"`
}

// Bind for Profile will be called after its inner field (Address) is bound.
func (p *Profile) Bind(r *http.Request) error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	fmt.Println("Profile.Bind called")
	return nil
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Initialize nested pointers so they can be populated by the decoder.
	profile := Profile{
		Address: &Address{},
	}

	if err := bind.Action(r, &profile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Profile received: %+v\n", profile)
}

func main() {
	http.HandleFunc("/profile", profileHandler)
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Test it with `curl`:**

The server logs will show that `Address.Bind` is called before `Profile.Bind`.

```sh
# Success Case
curl -X POST http://localhost:8080/profile -H "Content-Type: application/json" -d '{"name":"Jane Doe", "address":{"city":"Seoul", "country":"Korea"}}'
# Expected Output: Profile received: &{Name:Jane Doe Address:0x...}
# Server Logs:
# Address.Bind called
# Profile.Bind called

# Failure Case (nested validation fails)
curl -X POST http://localhost:8080/profile -H "Content-Type: application/json" -d '{"name":"Jane Doe", "address":{"city":"Seoul"}}'
# Expected Output: country is required
```

---

# `bind` (한국어)

요청 데이터(JSON, XML, Form)를 Go 구조체에 바인딩하는 간단하고 강력한 패키지입니다. 중첩된 구조체에 대한 재귀적 바인딩을 지원하며 성능에 최적화되어 있습니다.

## 주요 특징

- **다양한 Content-Type 지원:** `application/json`, `application/xml`, `application/x-www-form-urlencoded`, `multipart/form-data`를 지원합니다.
- **재귀적 바인딩:** `Binder` 인터페이스를 구현하는 중첩된 필드의 `Bind` 메서드를 가장 안쪽 필드부터 바깥쪽으로 순서대로 자동 호출합니다.
- **커스텀 로직 추가:** `Binder` 인터페이스를 구현하여 디코딩이 끝난 후 실행될 커스텀 유효성 검사나 데이터 조작 로직을 추가할 수 있습니다.
- **확장성:** 커스텀 Content-Type을 위한 새로운 디코더를 쉽게 등록할 수 있습니다.
- **성능 최적화:** 구조체 분석 결과를 캐싱하여 자주 사용되는 경로에서의 리플렉션 오버헤드를 최소화합니다.

## 설치

```sh
go get github.com/DevNewbie1826/bind
```

## 사용법

### 예제 1: 기본 JSON 바인딩 및 유효성 검사

JSON 요청을 구조체에 바인딩하고 커스텀 유효성 검사를 실행하는 간단한 예제입니다.

```go
// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

// User - 사용자 데이터 구조를 정의합니다.
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Bind - bind.Binder 인터페이스를 구현합니다.
// 이 메서드는 요청 본문이 구조체로 디코딩된 후에 호출됩니다.
// 이곳에 커스텀 유효성 검사 로직을 추가할 수 있습니다.
func (u *User) Bind(r *http.Request) error {
	if u.Name == "" {
		return fmt.Errorf("이름은 필수 항목입니다")
	}
	return nil
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST 메서드만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := bind.Action(r, &user); err != nil {
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

**`curl`로 테스트하기:**

```sh
# 성공 케이스
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name":"홍길동", "email":"gildong@example.com"}'
# 예상 출력: 수신된 사용자: {Name:홍길동 Email:gildong@example.com}

# 실패 케이스 (이름이 누락되어 유효성 검사 실패)
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"email":"gildong@example.com"}'
# 예상 출력: 이름은 필수 항목입니다
```

### 예제 2: 중첩 바인딩

`bind` 패키지는 `Binder` 인터페이스를 구현하는 중첩 구조체를 자동으로 처리합니다. `Bind` 메서드는 가장 안쪽 필드부터 바깥 구조체 순서로 재귀적으로 호출됩니다.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevNewbie1826/bind"
)

// Address - 주소 정보를 담습니다.
type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// Bind - Address에 대한 bind.Binder 인터페이스를 구현합니다.
func (a *Address) Bind(r *http.Request) error {
	if a.Country == "" {
		return fmt.Errorf("국가는 필수 항목입니다")
	}
	fmt.Println("Address.Bind 호출됨")
	return nil
}

// Profile - 사용자 정보와 중첩된 Address를 포함합니다.
type Profile struct {
	Name    string   `json:"name"`
	Address *Address `json:"address"`
}

// Profile의 Bind는 내부 필드(Address)가 바인딩된 후에 호출됩니다.
func (p *Profile) Bind(r *http.Request) error {
	if p.Name == "" {
		return fmt.Errorf("프로필 이름은 필수 항목입니다")
	}
	fmt.Println("Profile.Bind 호출됨")
	return nil
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// 디코더가 값을 채울 수 있도록 중첩된 포인터를 초기화합니다.
	profile := Profile{
		Address: &Address{},
	}

	if err := bind.Action(r, &profile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "수신된 프로필: %+v\n", profile)
}

func main() {
	http.HandleFunc("/profile", profileHandler)
	log.Println("서버 시작 중 :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**`curl`로 테스트하기:**

서버 로그를 통해 `Address.Bind`가 `Profile.Bind`보다 먼저 호출되는 것을 확인할 수 있습니다.

```sh
# 성공 케이스
curl -X POST http://localhost:8080/profile -H "Content-Type: application/json" -d '{"name":"제인 도", "address":{"city":"서울", "country":"대한민국"}}'
# 예상 출력: 수신된 프로필: &{Name:제인 도 Address:0x...}
# 서버 로그:
# Address.Bind 호출됨
# Profile.Bind 호출됨

# 실패 케이스 (중첩된 필드의 유효성 검사 실패)
curl -X POST http://localhost:8080/profile -H "Content-Type: application/json" -d '{"name":"제인 도", "address":{"city":"서울"}}'
# 예상 출력: 국가는 필수 항목입니다
```

---