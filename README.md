# CortexCore :: FHIR-Reject Debugger

[![Build & Test](https://img.shields.io/badge/build-passing-brightgreen.svg)]()
[![FHIR](https://img.shields.io/badge/HL7_FHIR-R4-blue.svg)]()
[![SATUSEHAT](https://img.shields.io/badge/Kemenkes-SATUSEHAT-059669.svg)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Zero Cloud Cost](https://img.shields.io/badge/Cloud_Exfiltration-0%25_(Offline)-green.svg)]()

> **Interactive Offline HL7 FHIR R4 Validator & Heuristic Auto-Sanitizer for Indonesian Hospital Information Systems (SIMRS) connecting to SATUSEHAT Kemenkes RI.**

Part of **AetherCortex Solutions** · Cluster 01: Health & SATUSEHAT  
Architected by **Rashqa Andrean Fitrah Sulaeman**

---

## Masalah yang Dipecahkan

Setiap faskes (Rumah Sakit, Puskesmas, Klinik) di Indonesia diwajibkan oleh **PMK No. 24 Tahun 2022** untuk mengintegrasikan Rekam Medis Elektronik (RME) mereka ke platform nasional **SATUSEHAT Kemenkes RI**. 

Namun, tim vendor & IT SIMRS di lapangan terus-menerus menghadapi hambatan penolakan transaksi (*HTTP 400 Bad Request* / *HTTP 422 Unprocessable Entity*) dengan pesan `OperationOutcome` yang sulit dipahami oleh programmer SIMRS:
- Format timestamp tanpa offset zona waktu (`+07:00` WIB / `+08:00` WITA).
- Format NIK kotor (terbawa tanda minus, spasi, atau bukan 16 digit angka).
- URL spesifikasi `system` terminologi Kemenkes yang hilang atau keliru (`KFA`, `ICD-10`, `ICD-9-CM`, `LOINC`).
- Inkonsistensi referensi internal resource `urn:uuid:` pada transaksi `Bundle`.

Menguji payload pasien ke API publik SATUSEHAT juga rentan melanggar **UU No. 27 Tahun 2022 tentang Perlindungan Data Pribadi (UU PDP)** jika data belum di-anonymize.

## Solusi: CortexCore FHIR-Reject Debugger

Aplikasi desktop native (Wails v2 + Go + Svelte 5) yang **100% offline**:
1. **Validasi Profil Resmi Kemenkes RI v1.0:**
   - `Patient`: NIK 16 digit, gender, tanggal lahir.
   - `Encounter`: Lifecycle status, class rawat jalan/inap, periode ISO-8601 ber-offset.
   - `Condition`: Kode diagnosis ICD-10 dengan URI resmi Kemenkes.
   - `Observation`: Tanda vital LOINC.
   - `Procedure`: Kode tindakan ICD-9-CM.
   - `MedicationRequest`: Resep obat dengan kode KFA 9-digit Kemenkes.
   - `Bundle`: Tipe transaksi & integritas referensi UUID.
2. **Heuristic Auto-Sanitizer (1-Click Auto-Fix):**
   - Otomatis mengonversi timestamp lokal ke ISO-8601 dengan offset zona waktu.
   - Membersihkan NIK dari karakter non-numerik.
   - Menyuntikkan standard system URIs yang terlupa oleh SIMRS.
   - Menghasilkan payload `.fixed.json` yang siap di-POST ulang ke SATUSEHAT.
3. **Audit Grade 3-Pane Interface (Clinical Precision Theme):**
   - Panel Kiri: Struktur bundel & status lisensi.
   - Panel Tengah: Editor monospaced dengan status kepatuhan interaktif.
   - Panel Kanan: Dek kartu rekomendasi solusi per baris kesalahan.
4. **Offline Ed25519 Cryptographic Licensing (Enterprise Tier):**
   - Verifikasi tanda tangan digital offline tanpa ketergantungan koneksi internet.
   - Fitur Batch Folder Scanner: audit ratusan file log transaksi SIMRS sekaligus.
   - Ekspor ringkasan kepatuhan interoperabilitas untuk direksi faskes.

---

## Instalasi & Menjalankan

### Prasyarat
- Go 1.22+
- Node.js 18+ & npm
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation)

### Build dari Source
```bash
# Clone repository
git clone https://github.com/Rasshqa/cortexcore-fhir-debugger.git
cd cortexcore-fhir-debugger

# Install frontend dependencies
cd frontend && npm install && cd ..

# Jalankan dalam mode development (live reload)
wails dev

# Build binary produksi
wails build -tags webkit2_41
```

Binary hasil build akan berada di `build/bin/fhir-debugger`.

---

## Unit Testing

Sistem dilengkapi suite unit test lengkap untuk mesin validasi, sanitasi, pemindai batch, dan lisensi kriptografi:

```bash
go test -v ./...
```

Output:
```text
=== RUN   TestSanitizeTimestampAndTimezone
--- PASS: TestSanitizeTimestampAndTimezone (0.00s)
=== RUN   TestSanitizeNIK
--- PASS: TestSanitizeNIK (0.00s)
=== RUN   TestScanDirectory
--- PASS: TestScanDirectory (0.00s)
=== RUN   TestVerifyLicenseKey
--- PASS: TestVerifyLicenseKey (0.00s)
=== RUN   TestValidateEncounterMissingClass
--- PASS: TestValidateEncounterMissingClass (0.00s)
=== RUN   TestValidatePatientInvalidNIK
--- PASS: TestValidatePatientInvalidNIK (0.00s)
=== RUN   TestValidateValidBundleTransaction
--- PASS: TestValidateValidBundleTransaction (0.00s)
PASS
```

---

## Hak Cipta & Lisensi

- Arsitektur & Implementasi oleh **Rashqa Andrean Fitrah Sulaeman**.
- Lisensi Kode Terbuka: **MIT License**.
- Solusi Komersial Faskes & Dukungan Integrasi: [AetherCortex](https://github.com/Rasshqa).
