# PLAN AI-01 — pgvector Semantic Search & RAG

> **Scope:** Menambah semantic search/RAG berbasis `pgvector` untuk memperkaya fitur DeepSeek Read-Only Chat.
> **Deliverable:** Extension `pgvector`, tabel embedding/chunk, pipeline indexing read-only dari data PRISM, tool semantic search yang tetap permission-aware, dan UI/admin status minimal.
> **Dependensi:** `PLAN_AI_00_DeepSeek_Readonly_Chat.md` sudah selesai dan stabil.
> **Posisi fase:** AI-00 -> AI-01. Jangan dikerjakan sebelum AI chat deterministic tool-calling berjalan aman.
> **Referensi wajib:** `AGENTS.md`, `docs/PRISM_Business_Rules.md`, `docs/prism_ddl.sql`, `docs/PRISM_API_Contract.md`, `docs/PRISM_Backend_Structure.md`, `docs/PRISM_Frontend_Structure.md`, `docs/PRISM_Error_Handling.md`, `docs/PRISM_Dev_Workflow.md`.

---

## Prinsip Wajib

- `pgvector` hanya untuk pencarian semantik/context retrieval, bukan sumber kebenaran angka agregat.
- Pertanyaan numerik tetap harus dijawab dari query/service deterministic PRISM.
- DeepSeek tidak boleh mengakses database langsung.
- Model tidak boleh membuat atau mengeksekusi raw SQL.
- Semua query vector tetap ditulis di `sql/queries/*.sql` dan digenerate dengan `sqlc`.
- Semua hasil semantic search wajib difilter berdasarkan permission user.
- Chunk tidak boleh mencampur data dari module permission berbeda dalam satu potongan teks.
- Embedding pipeline boleh menulis ke tabel indeks AI, tetapi tidak boleh mengubah tabel bisnis PRISM.
- Jangan menyimpan secret, token, password hash, atau data audit sensitif dalam chunk.

---

## Catatan Provider Embedding

DeepSeek V4 Pro dipakai untuk chat/reasoning. Untuk `pgvector`, sistem tetap butuh **embedding model** yang menghasilkan vector numerik.

Sebelum implementasi, pilih salah satu:

| Opsi | Keterangan |
|------|------------|
| Provider embedding eksternal | Misalnya provider OpenAI-compatible yang punya endpoint embeddings. Perlu API key terpisah. |
| Embedding lokal | Misalnya service lokal untuk multilingual embedding. Cocok jika data tidak boleh dikirim ke pihak ketiga. |
| DeepSeek embedding | Hanya dipakai jika dokumentasi resmi saat implementasi sudah menyediakan endpoint embedding. Jangan diasumsikan ada. |

Dimensi vector harus mengikuti model embedding yang dipilih. Jangan hardcode `1536` tanpa keputusan provider.

---

## Phase 0 — Approval Gate & Design Decision

**Tujuan:** Mengunci alasan penggunaan pgvector, provider embedding, dan boundary data sebelum migration dibuat.

### Task

1. Konfirmasi use case yang memang butuh semantic search:
   - "Cari proyek yang mirip dengan ..."
   - "Cari objective/scope/outcome yang membahas ..."
   - "Tanya dokumen/narasi panjang yang tidak cukup dengan filter SQL."
2. Tentukan embedding provider:
   - base URL,
   - model,
   - dimensi vector,
   - batas data yang boleh dikirim,
   - biaya dan rate limit.
3. Tentukan entity yang diindeks untuk MVP:
   - `bb_project`: project name, objective, scope_of_work, outputs, outcomes.
   - `gb_project`: project name, objective, scope_of_project, activities.
   - `dk_project`: project name, objectives, activity details.
   - `loan_agreement`: loan code, lender, dates, amount summary.
   - `monitoring_disbursement`: budget year, quarter, planned/realized summary, komponen.
4. Tentukan permission module per chunk:
   - `bb_project`
   - `gb_project`
   - `daftar_kegiatan`
   - `loan_agreement`
   - `monitoring_disbursement`
5. Update dokumen desain:
   - `docs/PRISM_Business_Rules.md`
   - `docs/PRISM_API_Contract.md`
   - `docs/PRISM_Backend_Structure.md`
   - `docs/PRISM_Dev_Workflow.md`

### Checklist

- [ ] Use case semantic search disetujui.
- [ ] Provider embedding dan dimensi vector disetujui.
- [ ] Daftar entity MVP disetujui.
- [ ] Mapping permission per entity disetujui.
- [ ] Keputusan data eksternal terdokumentasi.

---

## Phase 1 — PostgreSQL & Migration pgvector

**Tujuan:** Menyiapkan extension dan tabel indeks semantic tanpa mengganggu tabel bisnis.

### Verifikasi Environment

1. Cek apakah image PostgreSQL dev punya pgvector:

```sql
SELECT name, default_version
FROM pg_available_extensions
WHERE name = 'vector';
```

2. Jika tidak tersedia, update environment dev:
   - gunakan image PostgreSQL 16 yang sudah menyertakan pgvector, atau
   - install package pgvector pada Docker image database dev.
3. Jangan ubah tabel bisnis dengan drop-recreate.

### Migration

Buat migration incremental baru:

- `prism-backend/migrations/0000XX_ai_pgvector.up.sql`
- `prism-backend/migrations/0000XX_ai_pgvector.down.sql`

Contoh struktur konseptual:

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE ai_embedding_document (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type       VARCHAR(50) NOT NULL,
    entity_id         UUID NOT NULL,
    permission_module VARCHAR(50) NOT NULL,
    title             TEXT NOT NULL,
    source_hash       TEXT NOT NULL,
    metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
    indexed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (entity_type, entity_id, permission_module)
);

CREATE TABLE ai_embedding_chunk (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id       UUID NOT NULL REFERENCES ai_embedding_document(id) ON DELETE CASCADE,
    entity_type       VARCHAR(50) NOT NULL,
    entity_id         UUID NOT NULL,
    permission_module VARCHAR(50) NOT NULL,
    chunk_index       INT NOT NULL,
    content           TEXT NOT NULL,
    content_hash      TEXT NOT NULL,
    token_estimate    INT NOT NULL DEFAULT 0,
    embedding         vector(<DIMENSION>) NOT NULL,
    metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
    indexed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, chunk_index, content_hash)
);
```

Ganti `<DIMENSION>` sesuai provider embedding yang dipilih.

### Index

Gunakan cosine distance untuk teks:

```sql
CREATE INDEX idx_ai_embedding_chunk_vector_hnsw
ON ai_embedding_chunk
USING hnsw (embedding vector_cosine_ops);

CREATE INDEX idx_ai_embedding_chunk_permission
ON ai_embedding_chunk(permission_module);

CREATE INDEX idx_ai_embedding_chunk_entity
ON ai_embedding_chunk(entity_type, entity_id);
```

Jika dataset masih kecil dan HNSW belum tersedia di environment, evaluasi `ivfflat` atau sequential scan sementara, tetapi dokumentasikan tradeoff.

### Checklist

- [ ] `pg_available_extensions` diverifikasi.
- [ ] Docker/dev database mendukung `vector`.
- [ ] Migration up/down dibuat.
- [ ] `docs/prism_ddl.sql` disinkronkan sebagai referensi schema.
- [ ] Index vector dibuat.
- [ ] Migration diuji di database fresh.

---

## Phase 2 — sqlc Queries & Embedding Repository

**Tujuan:** Menambah query typed untuk indeks embedding tanpa raw SQL di Go.

### File

- `prism-backend/sql/queries/ai_embedding.sql`
- hasil generate di `internal/database/queries/` melalui `make generate`

### Query MVP

```sql
-- name: UpsertAIEmbeddingDocument :one
INSERT INTO ai_embedding_document (
    entity_type, entity_id, permission_module, title, source_hash, metadata, indexed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW()
)
ON CONFLICT (entity_type, entity_id, permission_module)
DO UPDATE SET
    title = EXCLUDED.title,
    source_hash = EXCLUDED.source_hash,
    metadata = EXCLUDED.metadata,
    indexed_at = NOW(),
    updated_at = NOW()
RETURNING *;

-- name: DeleteAIEmbeddingChunksByDocument :exec
DELETE FROM ai_embedding_chunk
WHERE document_id = $1;

-- name: InsertAIEmbeddingChunk :one
INSERT INTO ai_embedding_chunk (
    document_id, entity_type, entity_id, permission_module,
    chunk_index, content, content_hash, token_estimate, embedding, metadata, indexed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9::vector, $10, NOW()
)
RETURNING id;

-- name: SearchAIEmbeddingChunks :many
SELECT
    id,
    document_id,
    entity_type,
    entity_id,
    permission_module,
    content,
    metadata,
    1 - (embedding <=> $1::vector) AS score
FROM ai_embedding_chunk
WHERE permission_module = ANY($2::text[])
ORDER BY embedding <=> $1::vector
LIMIT $3;

-- name: DeleteStaleAIEmbeddingDocuments :exec
DELETE FROM ai_embedding_document
WHERE entity_type = $1
  AND indexed_at < $2;
```

### sqlc Type Handling

- Validasi hasil `make generate` untuk parameter `embedding`.
- Jika sqlc tidak menghasilkan tipe yang bersih untuk `vector`, pilih salah satu:
  - pakai string vector literal dengan cast `$1::vector`, atau
  - setelah approval dependency, pakai `pgvector-go` dan sqlc override.
- Jangan pakai `any`/`interface{}` untuk jalur data DB final.

### Checklist

- [ ] Query embedding ditulis di `sql/queries/ai_embedding.sql`.
- [ ] `make generate` lulus.
- [ ] Tidak ada edit manual di `internal/database/queries/*.go`.
- [ ] Tipe parameter embedding jelas dan tidak memakai `any` pada kode final.
- [ ] Query search memfilter `permission_module`.

---

## Phase 3 — Embedding Client & Indexer Service

**Tujuan:** Membuat pipeline untuk membangun chunk teks dan menyimpan vector.

### File

- `prism-backend/internal/model/ai_embedding.go`
- `prism-backend/internal/service/ai_embedding_client.go`
- `prism-backend/internal/service/ai_embedding_indexer.go`
- `prism-backend/internal/service/ai_embedding_chunker.go`

### Config

Tambahkan env:

```env
AI_EMBEDDING_ENABLED=false
AI_EMBEDDING_PROVIDER=
AI_EMBEDDING_BASE_URL=
AI_EMBEDDING_API_KEY=
AI_EMBEDDING_MODEL=
AI_EMBEDDING_DIMENSIONS=
AI_EMBEDDING_BATCH_SIZE=32
AI_EMBEDDING_TIMEOUT_SECONDS=60
AI_EMBEDDING_REINDEX_ON_START=false
```

### Chunk Builder

Tiap chunk harus berisi teks yang jelas dan self-contained:

```text
Jenis Data: Blue Book Project
Kode: BB-2025-001
Nama Proyek: ...
Program: ...
Mitra Kerja Bappenas: ...
Executing Agency: ...
Lokasi: ...
Objective: ...
Scope of Work: ...
Outputs: ...
Outcomes: ...
```

Aturan:

- Satu chunk hanya untuk satu `permission_module`.
- Simpan `entity_type`, `entity_id`, label, dan route target di `metadata`.
- Gunakan `content_hash` untuk skip embedding jika konten tidak berubah.
- Batasi ukuran chunk agar tidak terlalu besar.
- Untuk angka agregat, simpan sebagai konteks ringan saja; jawaban final tetap ambil angka dari tool deterministic.

### Indexer

Indexer MVP:

- Reindex manual per entity type.
- Reindex semua entity saat admin memanggil endpoint.
- Skip unchanged content berdasarkan hash.
- Delete stale document setelah reindex entity selesai.
- Batch embedding call sesuai rate limit.

### Checklist

- [ ] Embedding client dibuat dengan timeout dan error aman.
- [ ] API key embedding tidak pernah dikirim ke frontend.
- [ ] Chunker membuat teks permission-safe.
- [ ] Indexer skip unchanged content.
- [ ] Indexer dapat reindex per entity type.
- [ ] Stale chunks/documents dibersihkan.
- [ ] Unit test chunker dan hash behavior.

---

## Phase 4 — Admin Reindex API & Status

**Tujuan:** Memberi kontrol operasional tanpa menjalankan proses manual di database.

### Endpoint

| Method | Endpoint | Permission | Keterangan |
|--------|----------|------------|------------|
| `GET` | `/api/v1/ai-chat/embeddings/status` | ADMIN only | Status index, jumlah document/chunk, last indexed |
| `POST` | `/api/v1/ai-chat/embeddings/reindex` | ADMIN only | Reindex manual |

**Request Reindex:**

```json
{
  "entity_type": "bb_project",
  "force": false
}
```

`entity_type` boleh kosong untuk reindex semua MVP entity.

### Response Status

```json
{
  "data": {
    "enabled": true,
    "provider": "local",
    "model": "bge-m3",
    "dimensions": 1024,
    "documents": 120,
    "chunks": 360,
    "last_indexed_at": "2026-05-02T10:00:00Z"
  }
}
```

### Checklist

- [ ] Status endpoint dibuat.
- [ ] Reindex endpoint ADMIN only dibuat.
- [ ] Reindex tidak memblokir request terlalu lama atau punya timeout jelas.
- [ ] Error provider embedding dikembalikan aman.
- [ ] Log reindex mencatat entity type, jumlah chunk, dan durasi.

---

## Phase 5 — Semantic Search Tool untuk AI Chat

**Tujuan:** Menghubungkan pgvector ke DeepSeek tool-calling tanpa menghilangkan deterministic tools.

### Tool Baru

`semantic_search_context`

Input:

```json
{
  "query": "proyek irigasi yang mirip dengan ketahanan pangan di wilayah timur",
  "entity_types": ["bb_project", "gb_project"],
  "limit": 8
}
```

Output backend ke DeepSeek:

```json
{
  "matches": [
    {
      "entity_type": "bb_project",
      "entity_id": "uuid",
      "label": "BB-2025-001 - ...",
      "score": 0.82,
      "content": "potongan konteks singkat",
      "route": "/blue-books/.../projects/..."
    }
  ]
}
```

### Aturan Tool

- Backend membuat query embedding dari `query`.
- Search hanya terhadap `permission_module` yang dimiliki user.
- Limit maksimum kecil, misalnya 10.
- Jika pertanyaan meminta angka pasti, DeepSeek diarahkan untuk memakai tool deterministic setelah semantic search menemukan kandidat entity.
- Hasil semantic search harus dikembalikan sebagai sources di response frontend.

### Checklist

- [ ] Tool `semantic_search_context` ditambahkan di AI chat service.
- [ ] Tool memfilter permission user.
- [ ] Tool membatasi limit dan entity type.
- [ ] Tool tidak mengembalikan data sensitif.
- [ ] AI system prompt menjelaskan bahwa semantic search bukan sumber angka final.
- [ ] Integration test semantic search + deterministic follow-up.

---

## Phase 6 — Frontend Enhancements

**Tujuan:** Menampilkan sumber semantic search dan status indexing secara jelas.

### Chat Page

Update:

- Source chips menampilkan score/jenis entity.
- Source dapat diklik ke halaman detail jika route tersedia.
- Pesan saat semantic index belum aktif:
  - "Pencarian semantik belum aktif. Chat tetap memakai data terstruktur."

### Admin UI Opsional

Jika diperlukan, tambah panel ADMIN:

- Status embedding.
- Jumlah document/chunk.
- Last indexed.
- Tombol reindex.

File kandidat:

- `prism-frontend/src/types/ai-embedding.types.ts`
- `prism-frontend/src/services/ai-embedding.service.ts`
- `prism-frontend/src/pages/ai-chat/AIEmbeddingAdminPage.vue`

### Checklist

- [ ] Source semantic search tampil jelas.
- [ ] Link detail tidak broken.
- [ ] ADMIN dapat melihat status index jika UI admin disetujui.
- [ ] `npm.cmd run type-check` lulus.
- [ ] `npm.cmd run build` lulus.

---

## Phase 7 — Verification & Hardening

**Tujuan:** Memastikan akurasi, keamanan, performa, dan operasional cukup layak.

### Test

- Permission:
  - STAFF tanpa permission monitoring tidak mendapat chunk monitoring.
  - STAFF tanpa `bb_project:read` tidak mendapat chunk proyek.
  - ADMIN mendapat semua chunk yang diizinkan kebijakan.
- Security:
  - Prompt injection dalam content chunk tidak boleh membuat model mengeksekusi tool mutasi.
  - Query "hapus/update/tambah" tetap ditolak oleh AI-00 guardrail.
- Data quality:
  - Perubahan data proyek menghasilkan hash baru dan reindex.
  - Data terhapus menghasilkan stale cleanup.
- Performance:
  - Search p95 masih wajar pada dataset lokal.
  - Index vector dipakai oleh query planner saat dataset cukup besar.
- Recovery:
  - Embedding provider down tidak membuat chat deterministic ikut mati.
  - Reindex dapat diulang idempotent.

### Checklist

- [ ] `go test ./...` lulus.
- [ ] `npm.cmd run type-check` lulus.
- [ ] `npm.cmd run build` lulus.
- [ ] Fresh database migration lulus.
- [ ] Reindex manual lulus.
- [ ] Semantic search permission test lulus.
- [ ] DeepSeek chat tetap berfungsi saat embedding disabled.

---

## Acceptance Criteria

- Database memiliki extension `vector` dan tabel embedding terpisah dari tabel bisnis.
- Embedding index dapat dibuat ulang dari data PRISM tanpa mutasi tabel bisnis.
- Semantic search hanya mengembalikan chunk sesuai permission user.
- AI Chat dapat memakai `semantic_search_context` untuk mencari konteks berbasis makna.
- Jawaban numerik tetap memakai deterministic tools/query PRISM.
- Tidak ada raw SQL di file `.go`.
- Tidak ada edit manual di `internal/database/queries/*.go`.
- API key embedding hanya berada di backend env.
- Fitur tetap aman saat `AI_EMBEDDING_ENABLED=false`.
- Docs API, backend structure, frontend structure, dan dev workflow sinkron.

---

## Estimasi Effort

| Scope | Estimasi |
|-------|----------|
| Phase 0 | 0.5-1 hari |
| Phase 1 | 1-2 hari |
| Phase 2 | 1-2 hari |
| Phase 3 | 3-5 hari |
| Phase 4 | 1-2 hari |
| Phase 5 | 2-3 hari |
| Phase 6 | 1-2 hari |
| Phase 7 | 2-4 hari |
| Total MVP pgvector | 11-21 hari kerja |

Effort paling tidak pasti ada di provider embedding, sqlc type mapping untuk `vector`, dan setup extension pada environment PostgreSQL dev/prod.

