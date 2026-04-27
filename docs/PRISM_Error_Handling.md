# PRISM — Error Handling Guide

> Panduan penanganan error yang konsisten di backend (Go) dan frontend (Vue).

---

## 1. Format Error Response

Semua error response menggunakan format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Pesan yang aman untuk ditampilkan ke user",
    "details": [
      { "field": "bb_code", "message": "BB Code sudah digunakan" }
    ]
  }
}
```

`details` hanya ada untuk `VALIDATION_ERROR`. Untuk error lain, `details` tidak disertakan.

---

## 2. Error Codes & HTTP Status

| HTTP Status | Code | Kapan digunakan |
|-------------|------|----------------|
| 400 | `VALIDATION_ERROR` | Input tidak valid — field wajib kosong, format salah, constraint bisnis |
| 401 | `UNAUTHORIZED` | Token tidak ada, expired, atau tidak valid |
| 403 | `FORBIDDEN` | Token valid tapi tidak punya permission |
| 404 | `NOT_FOUND` | Resource tidak ditemukan |
| 409 | `CONFLICT` | Duplikat data (bb_code sudah ada, monitoring quarter sudah ada) |
| 422 | `BUSINESS_RULE_ERROR` | Validasi bisnis gagal (lender tidak valid untuk DK, LA sudah ada untuk DK ini) |
| 500 | `INTERNAL_ERROR` | Server error — jangan ekspos detail internal |

---

## 3. Backend — Go Error Handling

### 3.1 Custom Error Types

```go
// internal/errors/errors.go
package errors

import "net/http"

type AppError struct {
    Code       string
    Message    string
    StatusCode int
    Details    []FieldError
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

// Constructor helpers
func NotFound(msg string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: msg, StatusCode: http.StatusNotFound}
}

func Conflict(msg string) *AppError {
    return &AppError{Code: "CONFLICT", Message: msg, StatusCode: http.StatusConflict}
}

func BusinessRule(msg string) *AppError {
    return &AppError{Code: "BUSINESS_RULE_ERROR", Message: msg, StatusCode: http.StatusUnprocessableEntity}
}

func Validation(fields ...FieldError) *AppError {
    return &AppError{
        Code: "VALIDATION_ERROR",
        Message: "Input tidak valid",
        StatusCode: http.StatusBadRequest,
        Details: fields,
    }
}
```

### 3.2 PostgreSQL Error Mapping

```go
// internal/errors/pg.go
package errors

import (
    "github.com/jackc/pgx/v5/pgconn"
)

func FromPgError(err error) *AppError {
    var pgErr *pgconn.PgError
    if !errors.As(err, &pgErr) {
        return &AppError{Code: "INTERNAL_ERROR", Message: "Terjadi kesalahan", StatusCode: 500}
    }

    switch pgErr.Code {
    case "23505": // unique_violation
        return &AppError{Code: "CONFLICT", Message: "Data sudah ada", StatusCode: 409}
    case "23503": // foreign_key_violation
        return &AppError{Code: "VALIDATION_ERROR", Message: "Referensi data tidak valid", StatusCode: 400}
    case "23514": // check_violation
        return &AppError{Code: "VALIDATION_ERROR", Message: "Data tidak memenuhi aturan", StatusCode: 400}
    default:
        return &AppError{Code: "INTERNAL_ERROR", Message: "Terjadi kesalahan database", StatusCode: 500}
    }
}
```

### 3.3 Error Handler Middleware (Echo)

```go
// internal/middleware/error_handler.go
func ErrorHandler(err error, c echo.Context) {
    var appErr *errors.AppError
    var httpErr *echo.HTTPError

    switch {
    case errors.As(err, &appErr):
        c.JSON(appErr.StatusCode, map[string]any{"error": appErr})
    case errors.As(err, &httpErr):
        c.JSON(httpErr.Code, map[string]any{
            "error": map[string]any{
                "code":    "HTTP_ERROR",
                "message": httpErr.Message,
            },
        })
    default:
        log.Error().Err(err).Msg("unhandled error")
        c.JSON(500, map[string]any{
            "error": map[string]any{
                "code":    "INTERNAL_ERROR",
                "message": "Terjadi kesalahan, silakan coba lagi",
            },
        })
    }
}
```

### 3.4 Pola di Service Layer

```go
// Selalu wrap error dari DB dengan konteks yang jelas
func (s *BBProjectService) CreateBBProject(ctx context.Context, req model.CreateBBProjectRequest) (*model.BBProjectResponse, error) {
    // Validasi bisnis sebelum DB
    existing, _ := s.queries.GetBBProjectByCode(ctx, req.BBCode)
    if existing != nil {
        return nil, errors.Conflict("BB Code sudah digunakan")
    }

    tx, err := s.db.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    project, err := qtx.CreateBBProject(ctx, params)
    if err != nil {
        return nil, errors.FromPgError(err)  // map pg error ke AppError
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("commit tx: %w", err)
    }

    return toResponse(project), nil
}
```

### 3.5 Logging

```go
// BENAR — log error dengan konteks, jangan expose ke client
log.Error().Err(err).
    Str("module", "bb_project").
    Str("action", "create").
    Str("bb_code", req.BBCode).
    Msg("failed to create bb_project")

return echo.NewHTTPError(http.StatusInternalServerError, "gagal menyimpan data")

// SALAH — expose internal error ke client
return c.JSON(500, err.Error())
```

---

## 4. Frontend — Vue Error Handling

### 4.1 Axios Interceptor untuk Error Global

```typescript
// services/http.ts
http.interceptors.response.use(
  response => response,
  (error: AxiosError<ApiErrorResponse>) => {
    const status = error.response?.status
    const code = error.response?.data?.error?.code

    // 401 — redirect ke login
    if (status === 401) {
      useAuthStore().logout()
      router.push('/login')
      return Promise.reject(error)
    }

    // 403 — tampilkan pesan forbidden
    if (status === 403) {
      useToast().add({
        severity: 'error',
        summary: 'Akses Ditolak',
        detail: 'Anda tidak memiliki izin untuk melakukan tindakan ini',
        life: 5000,
      })
      return Promise.reject(error)
    }

    // 500 — tampilkan pesan generic
    if (status === 500) {
      useToast().add({
        severity: 'error',
        summary: 'Terjadi Kesalahan',
        detail: 'Server mengalami masalah, silakan coba lagi',
        life: 5000,
      })
    }

    return Promise.reject(error)
  }
)
```

### 4.2 Error Handling di Service

```typescript
// Semua service function mengembalikan data langsung atau throw error
// Jangan tangkap error di service — biarkan naik ke store atau komponen

export const BBProjectService = {
  async create(bbId: string, data: CreateBBProjectRequest): Promise<BBProject> {
    const response = await http.post<ApiResponse<BBProject>>(
      `/blue-books/${bbId}/projects`, data
    )
    return response.data.data
  },
}
```

### 4.3 Error Handling di Store

```typescript
// stores/blue-book.store.ts
async function createProject(bbId: string, data: CreateBBProjectRequest) {
    loading.value = true
    error.value = null
    try {
        const project = await BBProjectService.create(bbId, data)
        projects.value.unshift(project)
        return project
    } catch (err) {
        if (isAxiosError(err)) {
            error.value = err.response?.data?.error ?? null
        }
        throw err  // re-throw agar komponen bisa handle juga
    } finally {
        loading.value = false
    }
}
```

### 4.4 Error Handling di Form (Validation Error)

```typescript
// pages/blue-book/BBProjectFormPage.vue
const onSubmit = async (values: CreateBBProjectRequest) => {
    try {
        await bbStore.createProject(route.params.bbId as string, values)
        toast.add({ severity: 'success', summary: 'Berhasil', detail: 'Proyek berhasil disimpan', life: 3000 })
        router.push(`/blue-books/${route.params.bbId}/projects`)
    } catch (err) {
        if (isAxiosError(err)) {
            const apiError = err.response?.data?.error

            // VALIDATION_ERROR — tampilkan per field
            if (apiError?.code === 'VALIDATION_ERROR' && apiError.details) {
                apiError.details.forEach((detail: FieldError) => {
                    setFieldError(detail.field, detail.message)
                })
                return
            }

            // CONFLICT — tampilkan toast
            if (apiError?.code === 'CONFLICT') {
                toast.add({ severity: 'warn', summary: 'Data Duplikat', detail: apiError.message, life: 5000 })
                return
            }

            // BUSINESS_RULE_ERROR — tampilkan toast
            if (apiError?.code === 'BUSINESS_RULE_ERROR') {
                toast.add({ severity: 'error', summary: 'Tidak Diizinkan', detail: apiError.message, life: 5000 })
                return
            }
        }
        // Error lain sudah ditangani oleh interceptor global
    }
}
```

### 4.5 Kapan Gunakan Toast vs Inline vs Dialog

| Situasi | Mekanisme |
|---------|----------|
| Sukses (create, update, delete) | Toast success — 3 detik |
| Validation error per field | Inline error di bawah field |
| Conflict / business rule error | Toast warning/error — 5 detik |
| 403 Forbidden | Toast error — 5 detik |
| 500 Server error | Toast error — 5 detik |
| Konfirmasi sebelum delete | Dialog konfirmasi (bukan toast) |
| Error saat load halaman (404) | Full page error state |

---

## 5. Edge Cases yang Wajib Ditangani

### 5.1 Network Error (Timeout, Offline)
```typescript
// Deteksi network error
if (!error.response) {
    toast.add({
        severity: 'error',
        summary: 'Koneksi Bermasalah',
        detail: 'Periksa koneksi internet Anda',
        life: 5000,
    })
}
```

### 5.2 Permission Berubah Mid-Session
- Jika ADMIN mengubah permission STAFF sementara STAFF sedang login, request berikutnya akan dapat 403.
- Interceptor global sudah menangani 403 dengan menampilkan toast.
- STAFF tidak perlu logout — cukup mendapat notifikasi bahwa tindakan tidak diizinkan.

### 5.3 Concurrent Edit
- Tidak ada optimistic locking di PRISM saat ini.
- Jika dua Staff mengedit data yang sama secara bersamaan, yang terakhir menyimpan akan menang (last-write-wins).
- Audit trail akan merekam kedua perubahan.

### 5.4 SSE Disconnect
- Jika SSE disconnect, tampilkan indikator "Tidak terhubung" di topbar.
- Auto-reconnect setelah 5 detik — sudah diimplementasikan di `useSSE()`.
- Setelah reconnect berhasil, hilangkan indikator.
