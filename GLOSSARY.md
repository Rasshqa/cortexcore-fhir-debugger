# GLOSSARY: Cluster 01 — Digital Health & SATUSEHAT Interoperability

## Domain Terms

### SATUSEHAT
Platform ekosistem data kesehatan nasional yang dikelola oleh Kementerian Kesehatan Republik Indonesia (Kemenkes RI), mewajibkan seluruh fasilitas pelayanan kesehatan (Faskes) terintegrasi secara digital sesuai PMK No. 24 Tahun 2022.

### SIMRS (Sistem Informasi Manajemen Rumah Sakit)
Sistem komputerisasi yang memproses dan mengintegrasikan seluruh alur proses pelayanan Rumah Sakit (administrasi, billing, farmasi, laboratorium, rekam medis) ke dalam satu jaringan.

### HL7 FHIR R4 (Fast Healthcare Interoperability Resources, Release 4)
Standar protokol interoperabilitas data medis global berbasis REST dan JSON. Setiap entitas medis direpresentasikan sebagai *Resource* (misal: `Patient`, `Encounter`, `Condition`).

### FHIR Bundle
Dokumen JSON tunggal yang memuat sekumpulan (koleksi) Resource FHIR yang saling berelasi, dikirim dalam satu payload transaksi HTTP POST ke SATUSEHAT.

### OperationOutcome
Resource standar FHIR yang dikembalikan oleh server SATUSEHAT ketika sebuah request gagal diproses (HTTP 400/422), memuat kode error, lokasi baris/ekspresi JSONPath yang salah, dan pesan diagnostik.

### KFA (Kamus Farmasi dan Alat Kesehatan)
Standar terminologi kodefikasi obat dan alkes resmi Kemenkes RI. Setiap peresepan (`MedicationRequest`) wajib menggunakan kode KFA resmi, bukan nama dagang lokal.

### ICD-10 & ICD-9-CM
Standar klasifikasi penyakit internasional (diagnosa) dan prosedur tindakan medis (operasi/pemeriksaan) yang wajib dipetakan dalam resource `Condition` dan `Procedure`.

### Offline-First Validator
Mesin validasi skema dan semantik yang berjalan 100% pada CPU lokal pengguna tanpa mengirimkan satu byte pun data pasien ke internet, memenuhi prinsip kerahasiaan medis UU Pelindungan Data Pribadi (UU PDP).

### Auto-Fix Patch
Diff terstruktur (RFC 6902 JSON Patch) yang dihasilkan secara otomatis oleh sistem untuk memperbaiki galat umum (salah format timestamp, URI sistem terminologi yang hilang, sanitasi NIK).
