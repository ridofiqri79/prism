# PLAN AI-00 — DeepSeek Read-Only Chat

> **Scope:** Integrasi DeepSeek V4 Pro untuk halaman chat yang dapat bertanya tentang data PRISM secara read-only.
> **Deliverable:** Backend AI gateway yang aman, tool read-only berbasis service/query PRISM, halaman chat frontend, permission, audit/log operasional, dan dokumentasi kontrak.
> **Posisi fase:** Dikerjakan setelah baseline backend/frontend dan revision versioning utama stabil: BE-11 -> FE-10 -> AI-00.
> **Referensi wajib:** `AGENTS.md`, `docs/PRISM_Business_Rules.md`, `docs/PRISM_API_Contract.md`, `docs/PRISM_Backend_Structure.md`, `docs/PRISM_Frontend_Structure.md`, `docs/PRISM_Error_Handling.md`, `docs/PRISM_Dev_Workflow.md`.

---

## Prinsip Wajib

- DeepSeek tidak boleh menerima API key dari frontend.
- DeepSeek tidak boleh mengakses database langsung.
- Model tidak boleh membuat atau mengeksekusi raw SQL.
- Semua akses data harus lewat backend PRISM, `sqlc`, service layer, dan permission middleware.
- Semua tool AI bersifat read-only. Tidak ada create/update/delete/import/export mutasi data.
- Jawaban chat harus mengikuti permission user aktif. STAFF default deny tetap berlaku.
- Data yang dikirim ke DeepSeek harus seminimal mungkin: hanya data yang relevan untuk menjawab pertanyaan.
- Jangan mulai dengan `pgvector`. Gunakan query/service deterministik dulu; pgvector hanya fase lanjutan jika ada kebutuhan semantic search/RAG.

---

## Phase 0 — Scope, Risiko, dan Kontrak Data

**Tujuan:** Menentukan batas pertanyaan dan data yang boleh dikirim ke layanan eksternal.

### Task

1. Tentukan kategori pertanyaan MVP:
   - Ringkasan dashboard.
   - Pencarian proyek.
   - Status pipeline proyek.
   - Perjalanan proyek BB -> GB -> DK -> LA -> Monitoring.
   - Ringkasan monitoring dan penyerapan.
   - Ringkasan lender, instansi, wilayah, dan mata uang.
2. Tentukan data yang tidak boleh diekspos ke AI:
   - `audit_log` untuk STAFF.
   - Password hash, token, konfigurasi rahasia, dan API key.
   - Data user management kecuali ADMIN dan hanya jika eksplisit disetujui.
3. Tentukan module permission baru:
   - Rekomendasi: `ai_chat` dengan `can_read`.
   - AI chat juga wajib mengecek permission module sumber sebelum memanggil tool.
4. Update dokumen:
   - `docs/PRISM_Business_Rules.md` bagian Permission/AI Chat.
   - `docs/PRISM_API_Contract.md` bagian AI Chat.
   - `docs/PRISM_Backend_Structure.md` untuk AI gateway/tool layer.
   - `docs/PRISM_Frontend_Structure.md` untuk page/service/store AI chat.

### Checklist

- [ ] Scope pertanyaan MVP disepakati.
- [ ] Kebijakan data eksternal disepakati.
- [ ] Module permission `ai_chat` disepakati.
- [ ] Kontrak endpoint AI Chat ditulis di docs.

---

## Phase 1 — Backend Foundation

**Tujuan:** Menambah gateway backend DeepSeek yang aman, tanpa membuka akses database bebas.

### File yang Dibuat/Diubah

- `prism-backend/internal/config/config.go`
- `prism-backend/.env.example`
- `prism-backend/internal/model/ai_chat.go`
- `prism-backend/internal/service/ai_chat_service.go`
- `prism-backend/internal/service/ai_tool_service.go`
- `prism-backend/internal/handler/ai_chat_handler.go`
- `prism-backend/sql/queries/ai_chat.sql`
- `prism-backend/cmd/api/main.go`

### Env Backend

```env
DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-pro
DEEPSEEK_TIMEOUT_SECONDS=45
DEEPSEEK_MAX_TOKENS=2000
AI_CHAT_ENABLED=false
```

### Endpoint MVP

| Method | Endpoint | Permission | Keterangan |
|--------|----------|------------|------------|
| `POST` | `/api/v1/ai-chat/messages` | read: `ai_chat` | Kirim pertanyaan, backend menjalankan tool read-only, lalu mengembalikan jawaban |

**Request:**

```json
{
  "message": "Tampilkan proyek pinjaman JICA yang sudah masuk Monitoring",
  "conversation_id": "optional-client-generated-id"
}
```

**Response:**

```json
{
  "data": {
    "answer": "Ada 3 proyek ...",
    "conversation_id": "uuid",
    "sources": [
      {
        "tool": "search_projects",
        "label": "Project Master",
        "record_count": 3
      }
    ]
  }
}
```

### DeepSeek Client

- Pakai `net/http` Go standar dulu. Jangan tambah dependency baru tanpa konfirmasi.
- Gunakan Chat Completions API yang OpenAI-compatible.
- Timeout wajib.
- Jangan log API key.
- Error dari DeepSeek dikonversi menjadi error aman ke client.

### Tool Read-Only MVP

Backend mendefinisikan tool schema untuk model, lalu mengeksekusi tool secara internal.

| Tool | Sumber Data | Permission Sumber |
|------|-------------|-------------------|
| `search_projects` | `GET /projects` logic / `project.sql` | read: `bb_project` |
| `get_project_journey` | journey service | read: `bb_project` |
| `get_dashboard_summary` | dashboard service | authenticated + source-safe |
| `get_monitoring_summary` | monitoring/dashboard query | read: `monitoring_disbursement` |
| `list_loan_agreements` | loan agreement list query | read: `loan_agreement` |
| `lookup_master_data` | master list ringkas | read module terkait |

### Aturan Implementasi Backend

- Tulis query baru hanya di `sql/queries/ai_chat.sql` atau gunakan query/service existing.
- Jalankan `make generate` setelah query berubah.
- Jangan edit `internal/database/queries/*.go` manual.
- Handler hanya bind request dan response.
- Service bertanggung jawab untuk:
  - validasi prompt length,
  - permission sumber,
  - DeepSeek request,
  - tool-call loop,
  - sanitasi hasil tool,
  - response final.
- Tidak perlu transaksi DB karena endpoint read-only.
- Tambahkan rate limit sederhana jika sudah ada middleware lokal; jika belum ada, dokumentasikan sebagai hardening phase.

### Checklist

- [ ] Config DeepSeek ditambahkan dan tervalidasi.
- [ ] Feature flag `AI_CHAT_ENABLED` mencegah endpoint aktif tanpa sengaja.
- [ ] Model request/response AI Chat dibuat.
- [ ] DeepSeek client backend dibuat.
- [ ] Tool read-only MVP dibuat.
- [ ] Permission `ai_chat:read` diterapkan di route.
- [ ] Permission module sumber dicek sebelum tool dijalankan.
- [ ] `make generate` dijalankan jika ada query baru.
- [ ] `go test ./...` lulus.

---

## Phase 2 — Backend Guardrail, Logging, dan Tests

**Tujuan:** Memastikan AI chat tidak bocor data dan tidak bisa dipakai untuk mutasi.

### Guardrail

- Batasi panjang `message`.
- Batasi jumlah tool call per request, misalnya maksimal 4.
- Batasi jumlah row per tool, misalnya maksimal 20 record.
- Redaksi field sensitif sebelum dikirim ke DeepSeek.
- Tolak request jika feature flag mati.
- Tolak pertanyaan yang meminta perubahan data dengan pesan aman:
  - "Fitur chat hanya dapat membaca dan merangkum data."

### Logging

Minimal log internal:

- user id,
- request id,
- conversation id,
- tool yang dipanggil,
- jumlah record,
- latency DeepSeek,
- error code.

Jangan log:

- API key,
- full prompt berisi data sensitif,
- raw response besar dari DeepSeek.

### Tests

- Unit test permission tool:
  - user tanpa `monitoring_disbursement:read` tidak bisa memakai tool monitoring.
  - ADMIN bisa memakai semua tool.
- Unit test no-write:
  - prompt meminta "hapus/update/tambah" tetap ditolak.
- Integration smoke:
  - pertanyaan dashboard mengembalikan jawaban.
  - pertanyaan project journey memanggil tool journey.
  - DeepSeek error -> response aman.

### Checklist

- [ ] Guardrail prompt dan tool-call diterapkan.
- [ ] Logging internal aman.
- [ ] Test permission source module.
- [ ] Test no-write request.
- [ ] Test error handling DeepSeek.
- [ ] Test endpoint authenticated.

---

## Phase 3 — Frontend Chat Page

**Tujuan:** Menambah halaman chat tanpa bypass service/store dan permission guard.

### File yang Dibuat/Diubah

- `prism-frontend/src/types/ai-chat.types.ts`
- `prism-frontend/src/services/ai-chat.service.ts`
- `prism-frontend/src/stores/ai-chat.store.ts`
- `prism-frontend/src/pages/ai-chat/AIChatPage.vue`
- `prism-frontend/src/router/routes/ai-chat.routes.ts`
- `prism-frontend/src/router/index.ts`
- `prism-frontend/src/layouts/components/AppSidebar.vue`

### Frontend Contract

- Semua HTTP lewat `AIChatService`.
- State chat di Pinia store.
- Route meta:

```typescript
meta: {
  requiresAuth: true,
  title: 'Chat Data',
  permission: { module: 'ai_chat', action: 'read' },
}
```

### UI MVP

- Halaman bernama `Chat Data`.
- Message list user/assistant.
- Input multiline.
- Tombol kirim dan retry.
- Loading state.
- Source chips berdasarkan `sources[]`.
- Empty state dengan contoh pertanyaan yang aman, misalnya:
  - "Ringkas proyek yang sudah masuk Monitoring."
  - "Apa saja proyek dengan lender JICA?"
  - "Tampilkan penyerapan monitoring tahun 2025."

### Checklist

- [ ] Types AI Chat dibuat.
- [ ] Service AI Chat dibuat.
- [ ] Store AI Chat dibuat.
- [ ] Page AI Chat dibuat.
- [ ] Route AI Chat didaftarkan.
- [ ] Sidebar item `Chat Data` ditambahkan dengan permission `ai_chat`.
- [ ] Tidak ada axios langsung di komponen.
- [ ] Tidak ada interface lokal di file `.vue`.
- [ ] `npm.cmd run type-check` lulus.
- [ ] `npm.cmd run build` lulus.

---

## Phase 4 — UX Polish dan Operasional

**Tujuan:** Membuat fitur nyaman digunakan dan aman dipantau.

### Task

- Tambah streaming response jika dibutuhkan.
- Tambah copy button untuk jawaban.
- Tambah clear conversation lokal.
- Tambah indikator "read-only".
- Tambah batasan user-facing saat data tidak tersedia karena permission.
- Tambah observability sederhana untuk biaya/token usage jika DeepSeek mengembalikan usage.
- Pertimbangkan penyimpanan chat history hanya jika ada kebutuhan audit/riwayat. Jika disimpan, buat tabel/migration terpisah dan pastikan tidak menyimpan data sensitif berlebihan.

### Checklist

- [ ] UX chat responsif desktop/mobile.
- [ ] Pesan error user-friendly.
- [ ] Source hasil tool mudah dipahami.
- [ ] Token/cost usage minimal tercatat jika tersedia.
- [ ] Keputusan chat history disepakati.

---

## Phase 5 — Optional Semantic Search / RAG

**Tujuan:** Hanya dikerjakan jika user butuh pencarian semantik, bukan untuk MVP.

### Kapan Perlu

- User ingin mencari proyek berdasarkan kemiripan makna.
- User ingin Q&A atas teks panjang seperti objective, scope, outcomes, catatan, dokumen PDF, atau memo.
- Search SQL/filter biasa tidak cukup.

### Kandidat Implementasi

- Evaluasi `pgvector` melalui plan lanjutan `plans/PLAN_AI_01_Pgvector_Semantic_RAG.md`.
- Tambah tabel embedding per entity/chunk dengan metadata permission.
- Buat background sync embedding.
- Buat tool `semantic_search_projects`.
- Pastikan hasil semantic search tetap difilter permission.

### Checklist

- [ ] Kebutuhan semantic search terbukti.
- [ ] Desain embedding dan permission metadata disetujui.
- [ ] Migration pgvector disetujui.
- [ ] Sync/reindex strategy disetujui.

---

## Acceptance Criteria

- User dengan permission `ai_chat:read` dapat membuka halaman `Chat Data`.
- User tanpa permission `ai_chat:read` tidak dapat membuka halaman dan mendapat forbidden.
- DeepSeek API key hanya berada di backend env.
- Chat dapat menjawab pertanyaan read-only tentang project, journey, dashboard, monitoring, loan agreement, dan master data sesuai permission.
- Prompt yang meminta tambah/ubah/hapus data ditolak atau dijawab sebagai tidak didukung.
- Tidak ada raw SQL di file `.go`.
- Tidak ada edit manual pada `internal/database/queries/*.go`.
- Backend test lulus.
- Frontend type-check/build lulus.
- Docs API dan struktur backend/frontend sinkron dengan implementasi.

---

## Estimasi Effort

| Scope | Estimasi |
|-------|----------|
| Phase 0 | 0.5-1 hari |
| Phase 1 | 2-3 hari |
| Phase 2 | 1.5-2 hari |
| Phase 3 | 1.5-2.5 hari |
| Phase 4 | 1-2 hari |
| Total MVP | 6-10 hari kerja |
| Production hardening tambahan | +4-8 hari kerja |
| Optional pgvector/RAG | +1-3 minggu, tergantung cakupan dokumen dan embedding |
