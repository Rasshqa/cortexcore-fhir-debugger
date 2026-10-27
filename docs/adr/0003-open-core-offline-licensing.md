# ADR 0003: Open Core Tiering Model & Offline Ed25519 License Verification

## Status
Accepted

## Context
AetherCortex mengadopsi model bisnis komersial B2B untuk faskes dan SIMRS di Indonesia. Untuk menghasilkan customer sungguhan tanpa infrastruktur DRM cloud yang mahal dan rentan downtime jaringan intranet faskes, verifikasi lisensi fitur berbayar (Enterprise) harus dapat diverifikasi secara instan tanpa koneksi internet.

## Decision
1. **Tiering Fitur:**
   - **Community Tier (Free / Rp0):**
     - Validasi interaktif single payload JSON/Bundle.
     - Visual error tree & rekomendasi perbaikan berbasis standar SATUSEHAT.
     - Auto-fix patch generator untuk single file.
   - **Enterprise Tier (Paid):**
     - Batch folder scanner (memeriksa ribuan berkas log JSON transaksi SIMRS dalam satu operasi).
     - Ekspor laporan audit kepatuhan interoperabilitas terstruktur (JSON/Markdown/HTML report) untuk direksi RS.
     - Custom rule override untuk profil internal faskes.
2. **Kriptografi Lisensi Offline:**
   - Menggunakan tanda tangan digital **Ed25519**.
   - Public key AetherCortex di-embed di dalam binary Go.
   - Token lisensi memuat nama faskes, masa berlaku, dan digital signature.
   - Sistem memverifikasi token secara murni matematis di sisi klien tanpa panggilan API keluar.

## Consequences
- **Positif:** Keamanan tinggi, verifikasi instan (<1ms), zero cloud cost, faskes merasa aman karena tidak ada data telemetry yang dikirim ke server lisensi eksternal.
- **Negatif:** Revokasi lisensi real-time memerlukan update berkala file revocation list jika diperlukan.
