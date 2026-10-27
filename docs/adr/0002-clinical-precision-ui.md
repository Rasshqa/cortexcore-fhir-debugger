# ADR 0002: Clinical Precision Design System & 3-Pane Ergonomic Layout

## Status
Accepted

## Context
Aplikasi kesehatan beroperasi di lingkungan bertekanan tinggi: tim IT rumah sakit dan staf rekam medis sering kali harus menelusuri ribuan log error SATUSEHAT selama shift kerja panjang. Antarmuka bergaya dark mode ekstrem seragam atau UI konsumen umum menyebabkan kelelahan mata dan memperlambat pemecahan masalah teknis.

## Decision
Menerapkan bahasa desain **Clinical Precision** khusus untuk produk Klaster 1:
1. **Palet Warna Ergonomis:**
   - Background: Slate White (`#F8FAFC`) / Hospital Soft Ivory.
   - Panel & Navigasi: Deep Clinical Slate (`#0F172A`).
   - Aksen Aksi & Status Sehat: Clinical Emerald (`#059669`).
   - Peringatan Kritis / Reject: Alert Crimson (`#DC2626`).
   - Highlight Field: Amber (`#D97706`).
2. **Tata Letak 3-Pane Ergonomis:**
   - **Pane Kiri (260px):** File Explorer & Riwayat Batch Rejection log dari SIMRS.
   - **Pane Tengah (Fleksibel):** JSON Code Tree Viewer dengan line number, gutter error marker, dan syntax highlighting.
   - **Pane Kanan (340px):** SATUSEHAT Issue Inspector — kartu diagnosa error yang memetakan pesan teknis HTTP 400 ke aturan Kemenkes beserta tombol *One-Click Auto Fix*.
3. **Branding Permanen:**
   - Badge AetherCortex di pojok atas.
   - Footer wajib: `AetherCortex Solution · Architected by Rashqa Andrean Fitrah Sulaeman`.

## Consequences
- **Positif:** Kenyamanan visual maksimal untuk kerja analitis berjam-jam, navigasi error instan tanpa scrolling bolak-balik.
- **Negatif:** Membutuhkan resolusi layar minimum 1280x720 untuk kenyamanan 3-pane.
