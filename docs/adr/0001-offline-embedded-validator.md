# ADR 0001: Offline-First Embedded FHIR R4 Validation Architecture

## Status
Accepted

## Context
Rumah sakit dan klinik di Indonesia terikat PMK No. 24 Tahun 2022 dan UU Perlindungan Data Pribadi (UU PDP). Mengirim payload Rekam Medis Elektronik (RME) yang berisi NIK, nama pasien, dan rekam diagnosis ke cloud pihak ketiga untuk sekadar "validasi JSON" melanggar hukum dan berisiko kebocoran data medis. Selain itu, jaringan SIMRS di faskes sering kali berada di intranet terisolasi.

## Decision
`CortexCore :: FHIR-Reject Debugger` mengimplementasikan mesin validasi 100% lokal di dalam binary Go:
1. Skema JSON FHIR R4 resmi di-embed secara langsung (`//go:embed`).
2. Mesin validasi semantik memvalidasi kelengkapan mandatory field profil SATUSEHAT (misal: `Encounter.identifier`, `Condition.clinicalStatus`, timezone WIB/WITA/WIT).
3. Snapshot kamus terminologi esensial (KFA, ICD-10 umum) disimpan lokal via SQLite embedded.
4. Koneksi ke API SATUSEHAT Sandbox bersifat opsional murni (BYOK OAuth2), hanya dipicu saat pengguna secara eksplisit menekan tombol "Live Ping Verification".

## Consequences
- **Positif:** Kecepatan validasi <50ms per bundle, zero cloud cost, 100% kepatuhan UU PDP, dapat berjalan di laptop staf IT tanpa internet.
- **Negatif:** Pembaruan kamus KFA Kemenkes memerlukan patch update database lokal berkala.
