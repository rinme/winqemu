package vm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const libvirtDomainTemplate = `<domain type='kvm'>
  <name>{{.Name}}</name>
  <memory unit='MiB'>{{.RAMMB}}</memory>
  <currentMemory unit='MiB'>{{.RAMMB}}</currentMemory>
  <vcpu placement='static'>{{.TotalVCPU}}</vcpu>
  <os>
    <type arch='x86_64' machine='q35'>hvm</type>
    {{if .Tuning.SecureBoot}}
    <loader readonly='yes' type='pflash'>/usr/share/edk2-ovmf/x64/OVMF_CODE.secboot.fd</loader>
    {{end}}
    <boot dev='cdrom'/>
    <boot dev='hd'/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <hyperv mode='custom'>
      {{if .Tuning.HyperV.Relaxed}}<relaxed state='on'/>{{end}}
      {{if .Tuning.HyperV.VAPIC}}<vapic state='on'/>{{end}}
      {{if .Tuning.HyperV.Spinlocks}}<spinlocks state='on' retries='{{.Tuning.HyperV.SpinlockRetries}}'/>{{end}}
      {{if .Tuning.HyperV.Time}}<time state='on'/>{{end}}
      {{if .Tuning.HyperV.Synic}}<synic state='on'/>{{end}}
      {{if .Tuning.HyperV.STimer}}<stimer state='on'/>{{end}}
      {{if .Tuning.HyperV.Reset}}<reset state='on'/>{{end}}
      {{if .Tuning.HyperV.Frequencies}}<frequencies state='on'/>{{end}}
      {{if .Tuning.HyperV.Reenlightenment}}<reenlightenment state='on'/>{{end}}
    </hyperv>
    <vmport state='off'/>
  </features>
  <cpu mode='host-passthrough' check='none' migratable='on'>
    <topology sockets='1' dies='1' cores='{{.CPUCores}}' threads='{{.CPUThreads}}'/>
  </cpu>
  <clock offset='localtime'>
    <timer name='rtc' tickpolicy='catchup'/>
    <timer name='pit' tickpolicy='delay'/>
    <timer name='hpet' present='no'/>
    <timer name='hypervclock' present='yes'/>
  </clock>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2' cache='none' io='native'/>
      <source file='{{.DiskPath}}'/>
      <target dev='sda' bus='scsi'/>
    </disk>
    {{if .ISOPath}}
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='{{.ISOPath}}'/>
      <target dev='sdb' bus='sata'/>
      <readonly/>
    </disk>
    {{end}}
    {{if .VirtIOISOPath}}
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='{{.VirtIOISOPath}}'/>
      <target dev='sdc' bus='sata'/>
      <readonly/>
    </disk>
    {{end}}
    {{if .UnattendDisk}}
    <disk type='file' device='floppy'>
      <driver name='qemu' type='raw'/>
      <source file='{{.UnattendDisk}}'/>
      <target dev='fda' bus='fdc'/>
    </disk>
    {{end}}
    <controller type='scsi' model='virtio-scsi' index='0'/>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <channel type='spicevmc'>
      <target type='virtio' name='com.redhat.spice.0'/>
    </channel>
    <input type='tablet' bus='usb'/>
    <input type='keyboard' bus='ps2'/>
    {{if .Tuning.TPM2Emulation}}
    <tpm model='tpm-crb'>
      <backend type='emulator' version='2.0'/>
    </tpm>
    {{end}}
    <graphics type='spice' autoport='yes'>
      <image compression='off'/>
    </graphics>
    <video>
      <model type='qxl' ram='65536' vram='65536' vgamem='16384' heads='1' primary='yes'/>
    </video>
    <sound model='ich9'/>
    {{if .Tuning.LookingGlassEnabled}}
    <shmem name='looking-glass'>
      <model type='ivshmem-plain'/>
      <size unit='M'>{{if .Tuning.LookingGlassMemMB}}{{.Tuning.LookingGlassMemMB}}{{else}}64{{end}}</size>
    </shmem>
    {{end}}
  </devices>
</domain>
`

type LibvirtData struct {
	*VMConfig
	TotalVCPU int
}

// GenerateLibvirtXML generates a Libvirt domain XML definition string.
func GenerateLibvirtXML(cfg *VMConfig) (string, error) {
	tmpl, err := template.New("libvirt").Parse(libvirtDomainTemplate)
	if err != nil {
		return "", err
	}

	data := LibvirtData{
		VMConfig:  cfg,
		TotalVCPU: cfg.CPUCores * cfg.CPUThreads,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// WriteLibvirtXML writes the Libvirt domain XML into the VM output directory.
func WriteLibvirtXML(cfg *VMConfig) (string, error) {
	xmlContent, err := GenerateLibvirtXML(cfg)
	if err != nil {
		return "", err
	}

	xmlPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s.xml", cfg.Name))
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		return "", err
	}

	return xmlPath, nil
}

// RegisterLibvirtDomain defines the VM domain with virsh.
func RegisterLibvirtDomain(xmlPath string) error {
	cmd := exec.Command("virsh", "define", xmlPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh define failed: %w (output: %s)", err, string(out))
	}
	return nil
}

// StartLibvirtDomain starts the VM domain with virsh.
func StartLibvirtDomain(domainName string) error {
	cmd := exec.Command("virsh", "start", domainName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh start failed: %w (output: %s)", err, string(out))
	}
	return nil
}

// LaunchVirtManager opens virt-manager for the user.
func LaunchVirtManager(domainName string) error {
	var cmd *exec.Cmd
	if domainName != "" {
		cmd = exec.Command("virt-manager", "--show-domain-console", domainName)
	} else {
		cmd = exec.Command("virt-manager")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open virt-manager: %w", err)
	}
	return nil
}

// BuildVirtInstallCommand generates an alternative virt-install command line string.
func BuildVirtInstallCommand(cfg *VMConfig) string {
	osVariant := "win10"
	if cfg.SelectedEdition.IsWin11 {
		osVariant = "win11"
	}

	var parts []string
	parts = append(parts, "virt-install")
	parts = append(parts, fmt.Sprintf("--name %s", cfg.Name))
	parts = append(parts, fmt.Sprintf("--memory %d", cfg.RAMMB))
	parts = append(parts, fmt.Sprintf("--vcpus %d", cfg.CPUCores*cfg.CPUThreads))
	parts = append(parts, fmt.Sprintf("--cpu host-passthrough,topology.sockets=1,topology.cores=%d,topology.threads=%d", cfg.CPUCores, cfg.CPUThreads))
	parts = append(parts, fmt.Sprintf("--os-variant %s", osVariant))
	parts = append(parts, fmt.Sprintf("--disk path=%s,format=qcow2,bus=scsi", cfg.DiskPath))
	if cfg.ISOPath != "" {
		parts = append(parts, fmt.Sprintf("--cdrom %s", cfg.ISOPath))
	}
	if cfg.VirtIOISOPath != "" {
		parts = append(parts, fmt.Sprintf("--disk path=%s,device=cdrom", cfg.VirtIOISOPath))
	}
	if cfg.UnattendDisk != "" {
		parts = append(parts, fmt.Sprintf("--disk path=%s,device=floppy", cfg.UnattendDisk))
	}
	parts = append(parts, "--network default,model=virtio")
	parts = append(parts, "--graphics spice")
	parts = append(parts, "--noautoconsole")

	return strings.Join(parts, " \\\n  ")
}
