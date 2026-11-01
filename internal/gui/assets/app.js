// WinQEMU Frontend Application State & IPC Controller

let state = {
  host: null,
  catalog: null,
  selectedFamily: null,
  selectedEdition: null,
  existingVMs: [],
  useLocalISO: false,
  tuningProfile: "gaming",
  createdVMName: "",
};

// Titles & Subtitles for Navigation
const tabMeta = {
  overview: { title: "System Overview", subtitle: "Host telemetry and virtualization dependencies" },
  os: { title: "OS & Editions", subtitle: "Choose genuine Windows installation media curated by WispyDocs" },
  media: { title: "Installation Media", subtitle: "High-speed resumable downloader and ISO verification" },
  hardware: { title: "Hardware Config", subtitle: "Configure CPU, RAM, and virtual storage resources" },
  tuning: { title: "Performance Tuning", subtitle: "Hyper-V enlightenments, Looking Glass, and audio profiles" },
  deploy: { title: "Deploy & Launch", subtitle: "Build VM artifacts and launch in virt-manager" },
  vms: { title: "VM Manager", subtitle: "Manage and launch previously created virtual machines" },
};

document.addEventListener("DOMContentLoaded", async () => {
  setupNavigation();
  await loadHostInfo();
  await loadCatalog();
  await loadExistingVMs();
});

function setupNavigation() {
  document.querySelectorAll(".nav-item").forEach(item => {
    item.addEventListener("click", () => {
      const tab = item.getAttribute("data-tab");
      document.querySelectorAll(".nav-item").forEach(i => i.classList.remove("active"));
      document.querySelectorAll(".tab-pane").forEach(p => p.classList.remove("active"));

      item.classList.add("active");
      const targetPane = document.getElementById(`pane-${tab}`);
      if (targetPane) targetPane.classList.add("active");

      if (tabMeta[tab]) {
        document.getElementById("page-title").innerText = tabMeta[tab].title;
        document.getElementById("page-subtitle").innerText = tabMeta[tab].subtitle;
      }
    });
  });
}

// 1. Host Telemetry
async function loadHostInfo() {
  try {
    if (window.getHostInfo) {
      const info = JSON.parse(await window.getHostInfo());
      state.host = info;
      renderHostTelemetry(info);
    }
  } catch (err) {
    console.error("Failed to load host info:", err);
  }
}

function renderHostTelemetry(info) {
  document.getElementById("host-badge").innerText = `${info.distro_pretty} (${info.arch})`;
  document.getElementById("stat-distro").innerText = info.distro_pretty;
  document.getElementById("stat-ram").innerText = `${Math.round(info.total_ram_mb / 1024)} GB`;
  document.getElementById("stat-cpus").innerText = `${info.logical_cpus} CPUs`;

  let accelStr = "Unavailable";
  if (info.kvm_supported) accelStr = "KVM Enabled (/dev/kvm)";
  else if (info.hvf_supported) accelStr = "HVF (Apple Silicon/Intel)";
  else if (info.whpx_supported) accelStr = "WHPX (Windows)";
  document.getElementById("stat-accel").innerText = accelStr;

  // Defaults for Hardware inputs
  const defRAM = Math.max(4096, Math.floor(info.total_ram_mb / 2));
  const defCPU = Math.min(8, Math.max(2, Math.floor(info.logical_cpus / 2)));
  document.getElementById("input-ram").value = defRAM;
  document.getElementById("input-cpu").value = defCPU;
  document.getElementById("label-ram").innerText = `RAM Memory (MB) [Auto: ${defRAM} MB]:`;
  document.getElementById("label-cpu").innerText = `vCPU Cores [Host SMT Matched: ${defCPU} Cores]:`;

  // Render Dependencies table
  const tbody = document.getElementById("deps-table-body");
  tbody.innerHTML = "";
  let hasMissing = false;

  (info.dependencies || []).forEach(dep => {
    const tr = document.createElement("tr");
    let badgeClass = "badge-success";
    let badgeText = "✔ Installed";

    if (!dep.installed) {
      if (dep.required) {
        badgeClass = "badge-danger";
        badgeText = "✖ Missing (Required)";
        hasMissing = true;
      } else {
        badgeClass = "badge-warning";
        badgeText = "⚠ Optional";
      }
    }

    tr.innerHTML = `
      <td style="font-weight: 600;">${dep.name}</td>
      <td style="color: var(--text-muted);">${dep.description}</td>
      <td><span class="badge ${badgeClass}">${badgeText}</span></td>
    `;
    tbody.appendChild(tr);
  });

  const installBtn = document.getElementById("btn-install-deps");
  if (!hasMissing) {
    installBtn.style.display = "none";
  }
}

// 2. Catalog & Editions
async function loadCatalog() {
  try {
    if (window.getCatalog) {
      const cat = JSON.parse(await window.getCatalog());
      state.catalog = cat;
      renderFamilies(cat.families || []);
    }
  } catch (err) {
    console.error("Failed to load catalog:", err);
  }
}

function renderFamilies(families) {
  const container = document.getElementById("family-tabs");
  container.innerHTML = "";

  families.forEach((fam, idx) => {
    const btn = document.createElement("button");
    btn.className = `btn ${idx === 0 ? "btn-primary" : "btn-secondary"}`;
    btn.innerText = fam.display_name;
    btn.onclick = () => {
      container.querySelectorAll(".btn").forEach(b => {
        b.className = "btn btn-secondary";
      });
      btn.className = "btn btn-primary";
      state.selectedFamily = fam;
      renderEditions(fam.editions || []);
    };
    container.appendChild(btn);
  });

  if (families.length > 0) {
    state.selectedFamily = families[0];
    renderEditions(families[0].editions || []);
  }
}

function renderEditions(editions) {
  const container = document.getElementById("editions-list");
  container.innerHTML = "";

  editions.forEach((ed, idx) => {
    const card = document.createElement("div");
    card.className = `edition-card ${idx === 0 ? "selected" : ""}`;
    if (idx === 0) selectEdition(ed);

    let wispyBadge = "";
    if (ed.is_recommended) {
      wispyBadge = `<span class="badge badge-wispy">⭐ Recommended by WispyDocs</span>`;
    }

    card.innerHTML = `
      <div class="edition-header">
        <div class="edition-title">
          <span>${ed.name}</span>
          <span class="badge badge-secondary" style="background: rgba(255,255,255,0.06); font-family: var(--font-mono);">${ed.arch}</span>
          ${wispyBadge}
        </div>
        <div style="font-size: 11.5px; color: var(--text-dim); font-family: var(--font-mono);">${ed.version} • ${ed.release}</div>
      </div>
      <div class="edition-desc">${ed.recommendation}</div>
    `;

    card.onclick = () => {
      container.querySelectorAll(".edition-card").forEach(c => c.classList.remove("selected"));
      card.classList.add("selected");
      selectEdition(ed);
    };

    container.appendChild(card);
  });
}

function selectEdition(ed) {
  state.selectedEdition = ed;
  document.getElementById("input-vm-name").value = ed.name.replace(/\s+/g, "-");
  document.getElementById("media-edition-name").innerText = `Target: ${ed.name} (${ed.arch})`;

  // Check cache status
  if (window.checkISOCache) {
    window.checkISOCache(ed.id, ed.arch).then(res => {
      const data = JSON.parse(res);
      const notice = document.getElementById("cached-notice");
      const btn = document.getElementById("btn-start-download");
      if (data.cached) {
        notice.style.display = "block";
        btn.innerText = "Re-download Genuine ISO";
      } else {
        notice.style.display = "none";
        btn.innerText = "📥 Download Genuine ISO from Microsoft CDN";
      }
    });
  }
}

// 3. Media Actions
window.triggerDownload = function() {
  if (!state.selectedEdition) return;
  document.getElementById("progress-wrapper").style.display = "block";
  document.getElementById("btn-start-download").disabled = true;
  const status = document.getElementById("download-status");
  status.style.color = "";
  status.innerText = "Connecting to high-speed mirror...";

  if (window.startDownload) {
    window.startDownload(state.selectedEdition.id, state.selectedEdition.arch);
  }
};

window.toggleLocalISO = function() {
  const isChecked = document.getElementById("check-use-local-iso").checked;
  state.useLocalISO = isChecked;
  document.getElementById("local-iso-container").style.display = isChecked ? "block" : "none";
};

// Callback from Go backend
window.onDownloadProgress = function(percent, speedStr, done, err, downloadedGB, totalGB) {
  const bar = document.getElementById("download-progress-bar");
  const status = document.getElementById("download-status");
  const speed = document.getElementById("download-speed");

  bar.style.width = `${percent}%`;
  if (downloadedGB !== undefined && totalGB !== undefined && totalGB > 0) {
    status.innerText = `Downloading installation media... ${percent.toFixed(1)}% (${downloadedGB.toFixed(2)} / ${totalGB.toFixed(2)} GB)`;
  } else {
    status.innerText = `Downloading installation media... ${percent.toFixed(1)}%`;
  }
  speed.innerText = speedStr;

  if (err) {
    status.innerText = `✖ Error: ${err}`;
    status.style.color = "var(--danger)";
    document.getElementById("btn-start-download").disabled = false;
    return;
  }

  if (done) {
    status.innerText = "✔ Download completed and genuine ISO verified!";
    status.style.color = "var(--success)";
    document.getElementById("btn-start-download").disabled = false;
    const notice = document.getElementById("cached-notice");
    if (notice) notice.style.display = "block";
  }
};

// 4. Tuning
window.setProfile = function(prof) {
  state.tuningProfile = prof;
  const btnGaming = document.getElementById("btn-profile-gaming");
  const btnBalanced = document.getElementById("btn-profile-balanced");

  if (prof === "gaming") {
    btnGaming.className = "btn btn-primary";
    btnBalanced.className = "btn btn-secondary";
    document.getElementById("tune-looking-glass").checked = true;
    document.getElementById("tune-audio").value = "scream";
  } else {
    btnGaming.className = "btn btn-secondary";
    btnBalanced.className = "btn btn-primary";
    document.getElementById("tune-looking-glass").checked = false;
    document.getElementById("tune-audio").value = "ich9";
  }
};

// 5. Deploy & Launch
window.triggerDeploy = function() {
  const cfg = {
    name: document.getElementById("input-vm-name").value.trim() || "Windows-VM",
    ram_mb: parseInt(document.getElementById("input-ram").value) || 8192,
    cpu_cores: parseInt(document.getElementById("input-cpu").value) || 4,
    disk_size_gb: parseInt(document.getElementById("input-disk").value) || 64,
    storage_dir: document.getElementById("input-storage-dir").value.trim(),
    use_local_iso: state.useLocalISO,
    local_iso_path: document.getElementById("input-local-iso-path").value.trim(),
    edition_id: state.selectedEdition ? state.selectedEdition.id : "",
    edition_arch: state.selectedEdition ? state.selectedEdition.arch : "x64",
    bypass_tpm: document.getElementById("tune-bypass").checked,
    looking_glass: document.getElementById("tune-looking-glass").checked,
    virtio_scsi: document.getElementById("tune-virtio").checked,
    mas_activate: document.getElementById("tune-mas").checked,
    audio_backend: document.getElementById("tune-audio").value,
  };

  document.getElementById("btn-deploy").disabled = true;
  document.getElementById("deploy-progress").style.display = "block";
  document.getElementById("deploy-success-box").style.display = "none";

  if (window.deployVM) {
    window.deployVM(JSON.stringify(cfg));
  }
};

// Callback from Go backend
window.onDeployProgress = function(stage, percent, done, resultJSON, err) {
  const text = document.getElementById("deploy-status-text");
  const bar = document.getElementById("deploy-progress-bar");

  text.innerText = stage;
  bar.style.width = `${percent}%`;

  if (done) {
    document.getElementById("btn-deploy").disabled = false;
    document.getElementById("deploy-progress").style.display = "none";
    document.getElementById("deploy-success-box").style.display = "block";

    if (resultJSON) {
      const res = JSON.parse(resultJSON);
      state.createdVMName = res.vm_name;
      document.getElementById("deploy-details").innerHTML = `
        <strong>VM Name:</strong> ${res.vm_name}<br>
        <strong>Disk Image:</strong> ${res.disk_path}<br>
        <strong>Orchestration:</strong> ${res.is_libvirt_ready ? "Registered in Libvirt (virt-manager)" : "Standalone QEMU Launch Script"}
      `;
    }
    loadExistingVMs();
  }

  if (err) {
    text.innerText = `Error: ${err}`;
    document.getElementById("btn-deploy").disabled = false;
  }
};

window.triggerLaunchCreated = function() {
  if (state.createdVMName && window.launchVM) {
    window.launchVM(state.createdVMName);
  }
};

// 6. Existing VMs
async function loadExistingVMs() {
  try {
    if (window.getExistingVMs) {
      const vms = JSON.parse(await window.getExistingVMs(""));
      state.existingVMs = vms;
      renderExistingVMs(vms);
    }
  } catch (err) {
    console.error("Failed to load existing VMs:", err);
  }
}

function renderExistingVMs(vms) {
  const tbody = document.getElementById("vms-table-body");
  tbody.innerHTML = "";

  if (!vms || vms.length === 0) {
    tbody.innerHTML = `<tr><td colspan="4" style="text-align: center; color: var(--text-muted); padding: 20px;">No WinQEMU virtual machines found</td></tr>`;
    return;
  }

  vms.forEach(v => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td style="font-weight: 600; color: #ffffff;">${v.name}</td>
      <td style="color: var(--text-muted);">${v.disk_size_mb} MB</td>
      <td><span class="badge ${v.is_libvirt_defined ? "badge-success" : "badge-secondary"}">${v.is_libvirt_defined ? "Libvirt (virt-manager)" : "Script Launcher"}</span></td>
      <td>
        <button class="btn btn-primary" style="padding: 5px 12px; font-size: 12px;" onclick="window.launchExistingVM('${v.name}')">
          ▶ Start
        </button>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

window.refreshExistingVMs = function() {
  loadExistingVMs();
};

window.launchExistingVM = function(name) {
  if (window.launchVM) {
    window.launchVM(name);
  }
};

window.installDependencies = function() {
  if (window.installPackages) {
    window.installPackages();
  }
};
