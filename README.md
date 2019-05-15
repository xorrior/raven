# Raven
CobaltStrike External C2 for Websockets
* Additional info: https://www.cobaltstrike.com/help-externalc2
# Server Build (Ubuntu)

* Copy the server folder to your VPS.
* Run ./setup.sh as root or with sudo

# Client build (Windows)

* Copy the client folder to a Windows development host.
* Install the Windows 10, 8.1, or 7 [SDK](https://developer.microsoft.com/en-us/windows/downloads/windows-10-sdk)
* Verify that MSBuild is installed and note the MSBuild tools version noted in the registry key here: HKLM:\SOFTWARE\Microsoft\MSBuild\ToolsVersions\[Version]
* Run New-RavenPayload.ps1. Use Get-Help to view the available options.

* That's all folks. YMMV