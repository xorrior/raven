function New-RavenHeader
{
    <#
    .SYNOPSIS Generates a new header file in the post build event for the slack-client
    
    Author: @tifkin_ Lee Christensen
    #>
    [CmdletBinding()]
    [OutputType([string])]
    Param
    (
        [Parameter(Mandatory=$true,
                   ValueFromPipelineByPropertyName=$true,
                   Position=0)]
        $AssemblyPath
    )

	$Bytes = Get-Content -Raw -Encoding Byte $AssemblyPath
	$OutputStr = New-Object System.Text.StringBuilder

	$Counter = 1
	foreach($Byte in $Bytes) {
		$null = $OutputStr.Append("0x$('{0:X2}' -f $Byte),") 

		if($Counter % 12 -eq 0) {
			$null = $OutputStr.AppendLine()
			$null = $OutputStr.Append("`t")
		}
		$Counter++
	}

	$null = $OutputStr.Remove($OutputStr.Length-1,1)

	$Source = @'
#ifndef NETRAVENDLL_H_
#define NETRAVENDLL_H_

static const unsigned char netraven_dll[] = {
    REPLACE
};

static const unsigned int NETRAVEN_dll_len = LENGTH;

#endif
'@

	$Source = $Source -replace 'REPLACE',$OutputStr.ToString()
	$Source = $Source -replace 'LENGTH',$Bytes.Length
	$Source
}
