<script>
  import { onMount } from 'svelte';
  import {
    ValidateJSON,
    AutoFixJSON,
    OpenJSONFile,
    SaveFixedJSON,
    BatchScanDirectory,
    VerifyLicense
  } from '../wailsjs/go/main/App.js';

  // State
  let payloadText = $state(`{
  "resourceType": "Bundle",
  "type": "transaction",
  "entry": [
    {
      "fullUrl": "urn:uuid:6112ff39-9d0a-42c2-8b4b-9e1bfd8e4f1a",
      "resource": {
        "resourceType": "Encounter",
        "status": "finished",
        "class": {
          "system": "http://terminology.hl7.org/CodeSystem/v3-ActCode",
          "code": "AMB"
        },
        "period": {
          "start": "2026-03-31 08:30:00"
        },
        "subject": {
          "reference": "urn:uuid:d12c8b0e-6c6a-4632-9cb7-2c1cf6da93c0"
        }
      }
    },
    {
      "fullUrl": "urn:uuid:d12c8b0e-6c6a-4632-9cb7-2c1cf6da93c0",
      "resource": {
        "resourceType": "Patient",
        "identifier": [
          {
            "system": "https://fhir.kemkes.go.id/id/nik",
            "value": "3201-2345-6789-0001"
          }
        ],
        "name": [
          {
            "text": "Ahmad Fauzi"
          }
        ],
        "gender": "male",
        "birthDate": "1988-11-20"
      }
    }
  ]
}`);

  let validationReport = $state(null);
  let patches = $state([]);
  let licenseStatus = $state({
    is_licensed: false,
    organization: "Community Edition (Free)",
    tier: "COMMUNITY",
    expires_at: "Unlimited"
  });
  let activeTab = $state("editor"); // "editor" | "batch"
  let batchSummary = $state(null);
  let batchDir = $state("");
  let selectedIssue = $state(null);
  let isScanning = $state(false);

  onMount(async () => {
    // Initial validation
    runValidation();
    // Check community license default
    try {
      const lic = await VerifyLicense("");
      if (lic) licenseStatus = lic;
    } catch (e) {
      console.error(e);
    }
  });

  async function runValidation() {
    try {
      validationReport = await ValidateJSON(payloadText);
    } catch (e) {
      console.error(e);
    }
  }

  async function handleAutoFix() {
    try {
      const res = await AutoFixJSON(payloadText);
      if (res.error) {
        alert("Gagal melakukan auto-fix: " + res.error);
        return;
      }
      payloadText = res.fixed;
      patches = res.patches || [];
      await runValidation();
    } catch (e) {
      alert("Error: " + e);
    }
  }

  async function handleOpenFile() {
    try {
      const res = await OpenJSONFile();
      if (res && res.content) {
        payloadText = res.content;
        patches = [];
        await runValidation();
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function handleSaveFixed() {
    try {
      const res = await SaveFixedJSON(payloadText);
      if (res.saved) {
        alert("Berhasil disimpan: " + res.saved);
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function handleBatchScan() {
    if (!licenseStatus.is_licensed) {
      alert("Fitur Batch Scan Log SIMRS membutuhkan lisensi Enterprise.\nHubungi lisensi@aethercortex.id atau masukkan license key faskes Anda.");
      return;
    }
    isScanning = true;
    try {
      const res = await BatchScanDirectory(500);
      if (res.summary) {
        batchSummary = res.summary;
        batchDir = res.dir;
        activeTab = "batch";
      }
    } catch (e) {
      alert("Batch scan gagal: " + e);
    } finally {
      isScanning = false;
    }
  }

  function handleSelectBatchFile(f) {
    // Load batch file path into editor
  }
</script>

<div class="app-shell">
  <!-- Topbar Header -->
  <header class="topbar">
    <div class="topbar-title">
      <span class="dot"></span>
      <span class="brand">CortexCore :: FHIR-Reject Debugger</span>
      <span>·</span>
      <span>SATUSEHAT HL7 FHIR R4</span>
    </div>
    <div class="topbar-actions">
      <button class="btn btn-ghost" onclick={handleOpenFile}>
        Buka File...
      </button>
      <button class="btn btn-amber" onclick={handleAutoFix}>
        Auto-Fix Heuristik
      </button>
      <button class="btn btn-primary" onclick={handleSaveFixed}>
        Simpan .fixed.json
      </button>
      <button class="btn btn-ghost" onclick={handleBatchScan} disabled={isScanning}>
        {#if isScanning}Memindai...{:else}Batch Scan Folder{/if}
      </button>
    </div>
  </header>

  <!-- 3-Pane Main Work Area -->
  <main class="main-panes">
    <!-- Left Pane: Resource & Inspection Hierarchy / Batch File List -->
    <section class="pane pane-left">
      <div class="pane-header">
        <span>Struktur Bundel</span>
        {#if validationReport}
          <span class="badge {validationReport.passed ? 'badge-ok' : 'badge-error'}">
            {validationReport.passed ? 'VALID' : `${validationReport.error_count} REJECT`}
          </span>
        {/if}
      </div>
      <div class="pane-body">
        {#if activeTab === "editor"}
          <div class="file-item active">
            <span style="font-family: monospace; font-size: 10px;">[RAM]</span>
            <span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">Active Payload</span>
            {#if validationReport}
              <span class="badge {validationReport.passed ? 'badge-ok' : 'badge-error'}">
                {validationReport.passed ? '0 Err' : validationReport.error_count}
              </span>
            {/if}
          </div>

          <!-- Quick Spec Links -->
          <div style="padding: 12px; font-size: 11px; color: var(--cp-muted);">
            <div style="font-weight: 600; color: var(--cp-text); margin-bottom: 6px;">Profil Kemenkes RI:</div>
            <ul style="list-style: none; display: flex; flex-direction: column; gap: 4px;">
              <li>✓ PMK No. 24/2022 (RME)</li>
              <li>✓ UU No. 27/2022 (PDP)</li>
              <li>✓ ICD-10 Diagnosis</li>
              <li>✓ ICD-9-CM Prosedur</li>
              <li>✓ KFA Farmasi & Obat</li>
              <li>✓ LOINC Tanda Vital</li>
            </ul>
          </div>
        {:else if batchSummary}
          {#each batchSummary.files as f}
            <div class="file-item" onclick={() => handleSelectBatchFile(f)}>
              <span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{f.file_name}</span>
              <span class="badge {f.passed ? 'badge-ok' : 'badge-error'}">
                {f.passed ? 'PASS' : `${f.error_count} ERR`}
              </span>
            </div>
          {/each}
        {/if}
      </div>

      <!-- License Status Footprint -->
      <div style="padding: 8px 12px; background: var(--cp-panel-mid); border-top: 1px solid var(--cp-border); font-size: 10px;">
        <div style="color: var(--cp-muted);">Lisensi:</div>
        <div style="color: {licenseStatus.is_licensed ? 'var(--cp-emerald-lt)' : 'var(--cp-amber-lt)'}; font-weight: 600;">
          {licenseStatus.organization}
        </div>
      </div>
    </section>

    <!-- Center Pane: Interactive JSON Editor & Gutter -->
    <section class="pane pane-center">
      <div class="pane-header">
        <span>FHIR JSON Payload</span>
        <button class="btn btn-ghost" style="padding: 2px 8px; font-size: 10px;" onclick={runValidation}>
          Validasi Ulang
        </button>
      </div>

      <div class="editor-wrapper">
        <textarea
          class="json-editor-area"
          bind:value={payloadText}
          oninput={runValidation}
          placeholder="Tempel FHIR Bundle JSON di sini..."
          spellcheck="false"
        ></textarea>
      </div>

      {#if patches.length > 0}
        <div class="patch-log">
          <div style="font-weight: 700; margin-bottom: 2px;">LOG HEURISTIK AUTO-FIX ({patches.length} patch diterapkan):</div>
          {#each patches as p}
            <div class="patch-log-item">{p}</div>
          {/each}
        </div>
      {/if}

      <!-- Status Bar -->
      <footer class="status-bar">
        {#if validationReport}
          {#if validationReport.passed}
            <span class="status-pass">✓ 100% SATUSEHAT CONFORMANT — Tidak ada error penolakan</span>
          {:else}
            <span class="status-fail">✕ {validationReport.error_count} Kesalahan Penolakan Fatal (HTTP 400/422 SATUSEHAT)</span>
            {#if validationReport.warning_count > 0}
              <span class="status-muted">· {validationReport.warning_count} Peringatan Kualitas</span>
            {/if}
          {/if}
        {:else}
          <span class="status-muted">Menunggu validasi...</span>
        {/if}
      </footer>
    </section>

    <!-- Right Pane: OperationOutcome Inspector & Fix Recommendation Card -->
    <section class="pane pane-right">
      <div class="pane-header">
        <span>Rekomendasi Perbaikan</span>
        {#if validationReport}
          <span style="font-size: 10px; color: var(--cp-muted);">
            {validationReport.issues.length} Temuan
          </span>
        {/if}
      </div>

      <div class="pane-body">
        {#if validationReport && validationReport.issues.length > 0}
          {#each validationReport.issues as iss}
            <div
              class="issue-card {selectedIssue === iss ? 'selected' : ''}"
              onclick={() => selectedIssue = iss}
            >
              <div class="issue-card-header {iss.severity === 'FATAL_REJECT' ? 'error' : 'warning'}">
                <span class="issue-severity-dot {iss.severity === 'FATAL_REJECT' ? 'sev-error' : 'sev-warning'}"></span>
                <span>[{iss.rule_id}]</span>
                <span style="margin-left: auto; font-size: 9px; opacity: 0.8;">{iss.resource}</span>
              </div>
              <div class="issue-card-body">
                <div class="issue-path">{iss.path}</div>
                <div class="issue-msg">{iss.message}</div>
                <div class="issue-fix">{iss.suggestion}</div>
              </div>
            </div>
          {/each}
        {:else if validationReport && validationReport.passed}
          <div class="empty-state">
            <div style="font-size: 28px; color: var(--cp-emerald);">✓</div>
            <div class="empty-title">Bundle Bersih & Valid</div>
            <div class="empty-sub">
              Tidak ditemukan pelanggaran spesifikasi profil Kemenkes. Payload siap dikirim ke SATUSEHAT API endpoint.
            </div>
          </div>
        {:else}
          <div class="empty-state">
            <div class="empty-title">Siap Memeriksa</div>
            <div class="empty-sub">
              Tempel berkas FHIR JSON di editor tengah atau klik Buka File.
            </div>
          </div>
        {/if}
      </div>
    </section>
  </main>

  <!-- Legal & Attribution Footer -->
  <footer class="attribution">
    <div>
      <span class="attribution-brand">AetherCortex</span> Solution · Cluster 01: Health & SATUSEHAT
    </div>
    <div>
      Architected by <strong>Rashqa Andrean Fitrah Sulaeman</strong> · Local-First Zero Cloud Cost
    </div>
  </footer>
</div>
