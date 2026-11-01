package unattend

import (
	"bytes"
	"fmt"
	"text/template"
	"winqemu/pkg/catalog"
	"winqemu/pkg/tuning"
)

type XMLData struct {
	ProductKey       string
	EditionName      string
	Username         string
	Password         string
	BypassTPMAndOOBE bool
	IsWin11          bool
	RunMAS           bool
}

const autounattendTemplate = `<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
  <settings pass="windowsPE">
    <component name="Microsoft-Windows-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      {{if .BypassTPMAndOOBE}}
      <RunSynchronous>
        <RunSynchronousCommand wcm:action="add">
          <Order>1</Order>
          <Description>Bypass Windows 11 TPM Check</Description>
          <Path>reg add HKLM\SYSTEM\Setup\LabConfig /v BypassTPMCheck /t REG_DWORD /d 1 /f</Path>
        </RunSynchronousCommand>
        <RunSynchronousCommand wcm:action="add">
          <Order>2</Order>
          <Description>Bypass Windows 11 Secure Boot Check</Description>
          <Path>reg add HKLM\SYSTEM\Setup\LabConfig /v BypassSecureBootCheck /t REG_DWORD /d 1 /f</Path>
        </RunSynchronousCommand>
        <RunSynchronousCommand wcm:action="add">
          <Order>3</Order>
          <Description>Bypass Windows 11 RAM Check</Description>
          <Path>reg add HKLM\SYSTEM\Setup\LabConfig /v BypassRAMCheck /t REG_DWORD /d 1 /f</Path>
        </RunSynchronousCommand>
        <RunSynchronousCommand wcm:action="add">
          <Order>4</Order>
          <Description>Bypass Windows 11 CPU Check</Description>
          <Path>reg add HKLM\SYSTEM\Setup\LabConfig /v BypassCPUCheck /t REG_DWORD /d 1 /f</Path>
        </RunSynchronousCommand>
        <RunSynchronousCommand wcm:action="add">
          <Order>5</Order>
          <Description>Bypass Windows 11 Storage Check</Description>
          <Path>reg add HKLM\SYSTEM\Setup\LabConfig /v BypassStorageCheck /t REG_DWORD /d 1 /f</Path>
        </RunSynchronousCommand>
      </RunSynchronous>
      {{end}}
      <UserData>
        <AcceptEula>true</AcceptEula>
        <ProductKey>
          <Key>{{.ProductKey}}</Key>
          <WillShowUI>OnError</WillShowUI>
        </ProductKey>
      </UserData>
    </component>
    <component name="Microsoft-Windows-PnpCustomizationsWinPE" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <DriverPaths>
        <PathAndCredentials wcm:action="add" wcm:keyValue="1">
          <Path>E:\</Path>
        </PathAndCredentials>
        <PathAndCredentials wcm:action="add" wcm:keyValue="2">
          <Path>D:\</Path>
        </PathAndCredentials>
      </DriverPaths>
    </component>
  </settings>
  <settings pass="specialize">
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <ComputerName>WinQEMU-VM</ComputerName>
      <TimeZone>UTC</TimeZone>
    </component>
  </settings>
  <settings pass="oobeSystem">
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <AutoLogon>
        <Password>
          <Value>{{.Password}}</Value>
          <PlainText>true</PlainText>
        </Password>
        <Enabled>true</Enabled>
        <LogonCount>1</LogonCount>
        <Username>{{.Username}}</Username>
      </AutoLogon>
      <OOBE>
        <HideEULAPage>true</HideEULAPage>
        <HideLocalAccountScreen>true</HideLocalAccountScreen>
        <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
        <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
        <NetworkLocation>Work</NetworkLocation>
        <ProtectYourPC>3</ProtectYourPC>
      </OOBE>
      <UserAccounts>
        <LocalAccounts>
          <LocalAccount wcm:action="add">
            <Password>
              <Value>{{.Password}}</Value>
              <PlainText>true</PlainText>
            </Password>
            <Description>Administrator Account</Description>
            <DisplayName>{{.Username}}</DisplayName>
            <Group>Administrators</Group>
            <Name>{{.Username}}</Name>
          </LocalAccount>
        </LocalAccounts>
      </UserAccounts>
      {{if .RunMAS}}
      <FirstLogonCommands>
        <SynchronousCommand wcm:action="add">
          <Order>1</Order>
          <Description>Run Microsoft Activation Scripts (MAS)</Description>
          <CommandLine>powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://get.massgrave.dev | iex"</CommandLine>
        </SynchronousCommand>
      </FirstLogonCommands>
      {{end}}
    </component>
  </settings>
</unattend>
`

// GenerateAutounattendXML produces the unattended installation XML.
func GenerateAutounattendXML(edition catalog.Edition, opts tuning.OptimizationConfig, username, password string) (string, error) {
	if username == "" {
		username = "User"
	}

	tmpl, err := template.New("autounattend").Parse(autounattendTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := XMLData{
		ProductKey:       edition.ProductKey,
		EditionName:      edition.Name,
		Username:         username,
		Password:         password,
		BypassTPMAndOOBE: opts.BypassTPMAndOOBE,
		IsWin11:          edition.IsWin11,
		RunMAS:           opts.MASActivationPrompt,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render autounattend XML: %w", err)
	}

	return buf.String(), nil
}
