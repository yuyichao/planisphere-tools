$script_date = "20211103"

Function Send-PSReport {
<#
    .SYNOPSIS
        Publishes computer information to Planisphere. A personal or Support Group Self Report Key is required for submission and can be provided
        via command-line parameter, key file, or configuration file entry. By default, the function looks for a configuration INI file 
        in C:\etc\ named "planisphere-report" with or without an ".ini" extension.
    .PARAMETER SelfReportKey
        Use this parameter to pass a Self Report Key via the command line.
    .PARAMETER SelfReportKeyFile
        Use this parameter to pass a Self Report Key via a file.
    .PARAMETER ConfigFile
        Use this parameter to define a configuration file located in a non-default location with a non-default filename.
    .PARAMETER UserName
        Use this parameter to pass a user's Duke NetID when using a Support Group-based Self Report Key. A user' Duke NetID can also be passed in a config file.
    .EXAMPLE
        . .\planisphere-report.ps1; Send-PSReport -SelfReportKey "a1b2c3d4-e5f6-g7h8-i9j0-a1b2c3d4e5f6"
            
        Description
        -----------
        Defining a Self Report Key via the command line.
    .EXAMPLE
        . .\planisphere-report.ps1; Send-PSReport -SelfReportKey "a1b2c3d4-e5f6-g7h8-i9j0-a1b2c3d4e5f6" -User netid123
            
        Description
        -----------
        Defining a Support Group-based Self Report Key and an assigned user via the command line.

    .EXAMPLE
        . .\planisphere-report.ps1; Send-PSReport -SelfReportKeyFile "C:\etc\planisphere-report-key.txt"
            
        Description
        -----------
        Defining a Self Report Key via the command line using a text file to store the key.

    .EXAMPLE
        . .\planisphere-report.ps1; Send-PSReport -ConfigFile "C:\non-default-location\non-default-config-filename.ini"
            
        Description
        -----------
        Defining a configuration INI file located in a non-default location with a non-default filename.

    .EXAMPLE
        . .\planisphere-report.ps1; Send-PSReport
            
        Description
        -----------
        If a configuration INI file is used with the default name and location and containing the Self Report Key, no arguments are required.

    .EXAMPLE
        . .\planisphere-report.ps1; Send-PSReport -Test
            
        Description
        -----------
        Use the -Test parameter to send the JSON results to screen rather than to Planisphere.
#>

    [CmdletBinding()]
    Param(
        [Parameter()]
        [Alias('Key')]
        [String]$SelfReportKey=$null,

        [Parameter()]
        [Alias('KeyFile')]
        [String]$SelfReportKeyFile=$null,

        [Parameter()]
        [String]$ConfigFile=$null,

        [Parameter()]
        [String]$UserName=$null,

    # for future use
    #    [Parameter()]
    #    [String]$Overrides=$null,

    # for future use
    #    [Parameter()]
    #    [String]$ExtraData=$null,

        [Parameter(DontShow)]
        [Switch]$Test
    )

#    Process {
    try {Remove-TypeData System.Array} catch {}
    # Necessary to get around the ConvertTo-Json "Value","Count" bug
    # https://stackoverflow.com/a/38212718/45375 
    # (OR IS IT?! Throws an error on MS PS7, but bug was showing on Win PS5. Argh.)
    # Also throws an error when run more than once in the same session, which really 
    # only happens in testing, so...
    
    # Test for the various config options:
    # ConfigFile param?
    if (Test-StringVar $ConfigFile) {
        if (Test-Path $ConfigFile) {
            $SelfReportIni = Get-IniContent $ConfigFile
        } else {
            Write-Error "ConfigFile '$ConfigFile' not found."
        }
    } else {
        # Default location/name(s)
        if (Test-Path "/etc/planisphere-report") {
            $SelfReportIni = Get-IniContent "/etc/planisphere-report"
        } elseif (Test-Path "/etc/planisphere-report.ini") {
            $SelfReportIni = Get-IniContent "/etc/planisphere-report.ini"
        } else {
            #Nothing?
            $SelfReportIni = $null
        }
    }
    # Who's got the SelfReportKey? 
    if (Test-StringVar $SelfReportKey) {
        # if passed as a param, use it
        # (which requires no additional action)
    } elseif (Test-StringVar $SelfReportKeyFile) {
        # or, if passed a keyfile param...
        if (Test-Path $SelfReportKeyFile) {
            # ...and it exists, use it
            $SelfReportKey = Get-Content $SelfReportKeyFile -First 1
        } else {
            # passed a non-existent keyfile, throw an error
            "Passed keyfile does not exist. Exiting..."
        }
    } elseif (Test-StringVar $SelfReportIni.config.key) {
        # or, if it exists, use the key in the INI file
        $SelfReportKey = $SelfReportIni.config.key.Trim()
    } elseif (Test-StringVar $SelfReportIni.config.keyfile) {
        # or, if there's a keyfile in the INI...
        if (Test-Path $SelfReportIni.config.keyfile) {
            # ...and it exists, use it
            $SelfReportKey = Get-Content $SelfReportIni.config.keyfile -First 1
        } else {
            # passed a non-existent keyfile, throw an error
            "INI keyfile does not exist. Exiting..."
        }
    } elseif (Test-Path "/etc/planisphere-report-key") {
        # or, if it exists, use the default keyfile
        $SelfReportKey = Get-Content "/etc/planisphere-report-key" -First 1
    } elseif (Test-Path "/etc/planisphere-report-key.txt") {
        # or, if it exists, use the default keyfile, TXT-style
        $SelfReportKey = Get-Content "/etc/planisphere-report-key.txt" -First 1
    } else {
        # or throw an error... we need a key from somewhere...
        Throw "Missing Self Report Key. Exiting..."
    }
    if (!(Test-IsGuid $SelfReportKey)) {
        # Key not in the proper format, throw an error
        Throw "Invalid Self Report Key parameter. Exiting..."
    }
    
    # Most of the info we want is available from Get-CompuetrInfo
    $info = Get-ComputerInfo
    
    # Collect installed software from all hives
    $installed_software_keys = @()
    $installed_software_keys += "Registry::HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\"
    $installed_software_keys += "Registry::HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\"
    $users_hives = Get-ChildItem -Path "Registry::HKU\"
    foreach ($hive in $users_hives) {
        if ($hive.Name -notmatch '_Classes') {
            $installed_software_keys += "Registry::$($hive.Name)\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\"
        }
    }
    $installed_software = @()
    foreach ($key in $installed_software_keys) {
        $subkeys = Get-ChildItem -Path $key
        foreach ($subkey in $subkeys) {
            if ($subkey.GetValue('DisplayName') -gt "") {
                # Cast as [string] to avoid "method on $null" errors
                $installed_name = ([string]$subkey.GetValue('DisplayName')).Trim()
                $installed_publisher = ([string]$subkey.GetValue('Publisher')).Trim()
                $installed_version = ([string]$subkey.GetValue('DisplayVersion')).Trim()
                # PS expects an array here
                $installed_software_info = @()
                $installed_software_info += , ("$installed_name ($installed_publisher)" , $installed_version)
                if ($installed_software -notcontains $installed_software_info[0]) {
                    $installed_software += , $installed_software_info[0]
                }
                
            }
        }
    }
    
    # Collect MAC addresses
    $mac_addresses = (Get-NetAdapter).MacAddress 
    
    # Windows PowerShell v5 misspelled "BiosSer[i]alNumber" in Get-ComputerInfo. It hasn't been fixed yet. :-/
    # https://github.com/PowerShell/PowerShell/issues/10349
    # (It's fixed in Core and Microsoft PowerShell, though, so inconsistent. Yay.)
    $serial_number = &{If(Test-StringVar $info.BiosSerialNumber) {$info.BiosSerialNumber} Else {$info.BiosSeralNumber}}
    
    # They also misspelled "CsPhy[s]icallyInstalledMemory". Yeesh.
    # https://windowsserver.uservoice.com/forums/301869-powershell/suggestions/37195837-get-computerinfo-typo-in-csphyicallyinstalledmemor
    $installed_memory = [int]((&{If(Test-StringVar $info.CsPhysicallyInstalledMemory) {$info.CsPhysicallyInstalledMemory} Else {$info.CsPhyicallyInstalledMemory}}) / 1024)

    # If a UserName was not defined in a param...
    if (!(Test-StringVar $UserName)) {
        # ...analyze Security Event Log data for most frequent non-machine logins in the last week
        Try {
            # the FilterXPath is SOOOO much faster than going through them in PowerShell. srsly.
            $Events = Get-WinEvent -LogName "Security" -FilterXPath '
            <QueryList>
                <Query Id="0" Path="Security">
                    <Select Path="Security">
                        (*[EventData[Data[@Name="LogonType"]="2"]] or *[EventData[Data[@Name="LogonType"]="7"]]) 
                        and *[System[(EventID="4624") and TimeCreated[timediff(@SystemTime) &lt;= 604800000]]]
                    </Select>
                    <Suppress Path="Security">
                        (*[EventData[Data[@Name="TargetDomainName"]="Font Driver Host"]] 
                        or *[EventData[Data[@Name="TargetDomainName"]="Window Manager"]])
                    </Suppress>
                </Query>
            </QueryList>'
            $UserNames = @('')
            foreach ($Event in $Events) {
                $EventDataXML = ([xml]$Event.ToXml()).Event.EventData.Data
                foreach ($i in $EventDataXML) {
                    $EventDataHash = [ordered]@{}
                    foreach ($j in $i) {
                        $EventDataHash.Add($j.Name, $j.'#text')
                    }
                    $UserNames += ($EventDataHash['TargetUserName'])
                }
            }
            $UserName = ($UserNames | Group-Object | Sort-Object Count -descending | Select-Object -First 1).Name
        } Catch [System.Exception] {
            # Most likely nobody has logged in or unlocked the device in a week, so no System Event Log records were found
            $UserName = $null
        }
    }
    
    # Build hash table of standard values        
    $hash = [ordered]@{}
    $hash.Add("key", (Get-CimInstance win32_useraccount)[0].SID.SubString(0,41))
    $hash.Add("last_active", (Get-Date -UFormat "%Y-%m-%d %H:%M:%S %Z"))
    
    $hash_data = [ordered]@{}
    $hash_data.Add("hostname", $(hostname))
    $hash_data.Add("device_type", $info.CsChassisSKUNumber)
    $hash_data.Add("manufacturer", $info.CsManufacturer)
    $hash_data.Add("model", $info.CsModel)
    $hash_data.Add("serial", $serial_number)
    $hash_data.Add("memory_mb", $installed_memory)
    $hash_data.Add("mac_addresses", $mac_addresses)
    $hash_data.Add("os_family", "Windows")
    $hash_data.Add("os_fullname", "$($info.WindowsProductName) ($($info.WindowsVersion))")
    $hash_data.Add("disk_encrypted", ((Get-BitLockerVolume -MountPoint C:).VolumeStatus -eq "FullyEncrypted"))
    $hash_data.Add("installed_software", $installed_software)
    $hash_data.Add("status", "deployed")
    $hash_data.Add("usage_type", "assigned_user")
    if (Test-StringVar $UserName) { $hash_data.Add("username", $UserName) }
    $hash.Add("data", $hash_data)

    $hash_extra_data = [ordered]@{}
    $hash_extra_data.Add("script_version", $script_date)
    $hash.Add("extra_data", $hash_extra_data)

    #Add overrides and extra_data from INI
    if ($null -ne $SelfReportIni.overrides.keys) {
        foreach ($key in $SelfReportIni.overrides.keys) {
            switch ($key) {
                # Can't Add() if the key exists, so use hash[key]=value here instead
                { $_ -in @("url","key") } { $hash[$key] = $SelfReportIni.overrides[$key] }
                Default { $hash.data[$key] = $SelfReportIni.overrides[$key].Trim() }
            } 
        }
    }
    if ($null -ne $SelfReportIni.extra_data.keys) {
        foreach ($key in $SelfReportIni.extra_data.keys) {
            $hash_extra_data.Add($key, $SelfReportIni.extra_data[$key].Trim())
        }
    }
    
    # Convert the hash to json
    $json = $hash | ConvertTo-Json -Depth 99

    # Set up the request headers
    $headers = New-Object "System.Collections.Generic.Dictionary[[String],[String]]"
    $headers.Add("PLANISPHERE-REPORT-KEY", $SelfReportKey)
    $headers.Add("User-Agent", 'planisphere-report')
    $headers.Add("Content-Type", 'application/json')

    # Allow for debug testing to planisphere-test by using a INI config URL.
    $url = 'https://planisphere.oit.duke.edu/self_report'
    if ($null -ne $SelfReportIni.config.url) { $url = $SelfReportIni.config.url }
    
    if ($Test) {
        $response ="report-key: $SelfReportKey`n"
        $response += $json
        $response
    } else {
#            try {
            $response = Invoke-RestMethod $url -Method Post -Headers $headers -Body $json -ContentType 'application/json'
#            } catch {}
    }
}

function Get-IniContent ($FilePath) {
# Lifted from https://devblogs.microsoft.com/scripting/use-powershell-to-work-with-any-ini-file/
# A more complete version lives at https://github.com/lipkau/PsIni, but this simpler version is fine here.
    $ini = @{}
    switch -regex -file $FilePath
    {
        "^\[(.+)\]" # Section
        {
            $Section = $Matches[1]
            $ini[$Section] = @{}
            $CommentCount = 0
        }
        "^(;.*)$" # Comment
        {
            $Value = $Matches[1]
            $CommentCount = $CommentCount + 1
            $Name = "Comment" + $CommentCount
            $ini[$Section][$Name] = $Value
        }
        "(.+?)\s*=(.*)" # Key
        {
            $Name,$Value = $Matches[1..2]
            $ini[$Section][$Name] = $Value
        }
    }
    return $ini
}

function Test-StringVar ([string]$StringVar) {
# A simple not-null and not-blank test function
# created in order to streamline the code a bit.
    if ($null -eq $StringVar) {
        $false
    } elseif ($StringVar.Trim() -eq '') {
        $false
    } else {
        $true
    }
}

function Test-IsGuid {
# Lifted from https://pscustomobject.github.io/powershell/functions/PowerShell-Validate-Guid/
    [OutputType([bool])]
    param
    (
        [Parameter(Mandatory = $true)]
        [string]$ObjectGuid
    )

    # Define verification regex
    [regex]$guidRegex = '(?im)^[{(]?[0-9A-F]{8}[-]?(?:[0-9A-F]{4}[-]?){3}[0-9A-F]{12}[)}]?$'

    # Check guid against regex
    return $ObjectGuid -match $guidRegex
}

# SIG # Begin signature block
# MIIlPAYJKoZIhvcNAQcCoIIlLTCCJSkCAQExCzAJBgUrDgMCGgUAMGkGCisGAQQB
# gjcCAQSgWzBZMDQGCisGAQQBgjcCAR4wJgIDAQAABBAfzDtgWUsITrck0sYpfvNR
# AgEAAgEAAgEAAgEAAgEAMCEwCQYFKw4DAhoFAAQUU5KoyJKx7E3wGTgqUR1Jz0kX
# 7L6ggh/pMIIEMjCCAxqgAwIBAgIBATANBgkqhkiG9w0BAQUFADB7MQswCQYDVQQG
# EwJHQjEbMBkGA1UECAwSR3JlYXRlciBNYW5jaGVzdGVyMRAwDgYDVQQHDAdTYWxm
# b3JkMRowGAYDVQQKDBFDb21vZG8gQ0EgTGltaXRlZDEhMB8GA1UEAwwYQUFBIENl
# cnRpZmljYXRlIFNlcnZpY2VzMB4XDTA0MDEwMTAwMDAwMFoXDTI4MTIzMTIzNTk1
# OVowezELMAkGA1UEBhMCR0IxGzAZBgNVBAgMEkdyZWF0ZXIgTWFuY2hlc3RlcjEQ
# MA4GA1UEBwwHU2FsZm9yZDEaMBgGA1UECgwRQ29tb2RvIENBIExpbWl0ZWQxITAf
# BgNVBAMMGEFBQSBDZXJ0aWZpY2F0ZSBTZXJ2aWNlczCCASIwDQYJKoZIhvcNAQEB
# BQADggEPADCCAQoCggEBAL5AnfRu4ep2hxxNRUSOvkbIgwadwSr+GB+O5AL686td
# UIoWMQuaBtDFcCLNSS1UY8y2bmhGC1Pqy0wkwLxyTurxFa70VJoSCsN6sjNg4tqJ
# VfMiWPPe3M/vg4aijJRPn2jymJBGhCfHdr/jzDUsi14HZGWCwEiwqJH5YZ92IFCo
# kcdmtet4YgNW8IoaE+oxox6gmf049vYnMlhvB/VruPsUK6+3qszWY19zjNoFmag4
# qMsXeDZRrOme9Hg6jc8P2ULimAyrL58OAd7vn5lJ8S3frHRNG5i1R8XlKdH5kBjH
# Ypy+g8cmez6KJcfA3Z3mNWgQIJ2P2N7Sw4ScDV7oL8kCAwEAAaOBwDCBvTAdBgNV
# HQ4EFgQUoBEKIz6W8Qfs4q8p74Klf9AwpLQwDgYDVR0PAQH/BAQDAgEGMA8GA1Ud
# EwEB/wQFMAMBAf8wewYDVR0fBHQwcjA4oDagNIYyaHR0cDovL2NybC5jb21vZG9j
# YS5jb20vQUFBQ2VydGlmaWNhdGVTZXJ2aWNlcy5jcmwwNqA0oDKGMGh0dHA6Ly9j
# cmwuY29tb2RvLm5ldC9BQUFDZXJ0aWZpY2F0ZVNlcnZpY2VzLmNybDANBgkqhkiG
# 9w0BAQUFAAOCAQEACFb8AvCb6P+k+tZ7xkSAzk/ExfYAWMymtrwUSWgEdujm7l3s
# Ag9g1o1QGE8mTgHj5rCl7r+8dFRBv/38ErjHT1r0iWAFf2C3BUrz9vHCv8S5dIa2
# LX1rzNLzRt0vxuBqw8M0Ayx9lt1awg6nCpnBBYurDC/zXDrPbDdVCYfeU0BsWO/8
# tqtlbgT2G9w84FoVxp7Z8VlIMCFlA2zs6SFz7JsDoeA3raAVGI/6ugLOpyypEBMs
# 1OUIJqsil2D4kF501KKaU73yqWjgom7C12yxow+ev+to51byrvLjKzg6CYG1a4XX
# vi3tPxq3smPi9WIsgtRqAEFQ8TmDn5XpNpaYbjCCBP4wggPmoAMCAQICEA1CSuC+
# Ooj/YEAhzhQA8N0wDQYJKoZIhvcNAQELBQAwcjELMAkGA1UEBhMCVVMxFTATBgNV
# BAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3LmRpZ2ljZXJ0LmNvbTExMC8G
# A1UEAxMoRGlnaUNlcnQgU0hBMiBBc3N1cmVkIElEIFRpbWVzdGFtcGluZyBDQTAe
# Fw0yMTAxMDEwMDAwMDBaFw0zMTAxMDYwMDAwMDBaMEgxCzAJBgNVBAYTAlVTMRcw
# FQYDVQQKEw5EaWdpQ2VydCwgSW5jLjEgMB4GA1UEAxMXRGlnaUNlcnQgVGltZXN0
# YW1wIDIwMjEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDC5mGEZ8WK
# 9Q0IpEXKY2tR1zoRQr0KdXVNlLQMULUmEP4dyG+RawyW5xpcSO9E5b+bYc0VkWJa
# uP9nC5xj/TZqgfop+N0rcIXeAhjzeG28ffnHbQk9vmp2h+mKvfiEXR52yeTGdnY6
# U9HR01o2j8aj4S8bOrdh1nPsTm0zinxdRS1LsVDmQTo3VobckyON91Al6GTm3dOP
# L1e1hyDrDo4s1SPa9E14RuMDgzEpSlwMMYpKjIjF9zBa+RSvFV9sQ0kJ/SYjU/aN
# Y+gaq1uxHTDCm2mCtNv8VlS8H6GHq756WwogL0sJyZWnjbL61mOLTqVyHO6fegFz
# +BnW/g1JhL0BAgMBAAGjggG4MIIBtDAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/
# BAIwADAWBgNVHSUBAf8EDDAKBggrBgEFBQcDCDBBBgNVHSAEOjA4MDYGCWCGSAGG
# /WwHATApMCcGCCsGAQUFBwIBFhtodHRwOi8vd3d3LmRpZ2ljZXJ0LmNvbS9DUFMw
# HwYDVR0jBBgwFoAU9LbhIB3+Ka7S5GGlsqIlssgXNW4wHQYDVR0OBBYEFDZEho6k
# urBmvrwoLR1ENt3janq8MHEGA1UdHwRqMGgwMqAwoC6GLGh0dHA6Ly9jcmwzLmRp
# Z2ljZXJ0LmNvbS9zaGEyLWFzc3VyZWQtdHMuY3JsMDKgMKAuhixodHRwOi8vY3Js
# NC5kaWdpY2VydC5jb20vc2hhMi1hc3N1cmVkLXRzLmNybDCBhQYIKwYBBQUHAQEE
# eTB3MCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdpY2VydC5jb20wTwYIKwYB
# BQUHMAKGQ2h0dHA6Ly9jYWNlcnRzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFNIQTJB
# c3N1cmVkSURUaW1lc3RhbXBpbmdDQS5jcnQwDQYJKoZIhvcNAQELBQADggEBAEgc
# 3LXpmiO85xrnIA6OZ0b9QnJRdAojR6OrktIlxHBZvhSg5SeBpU0UFRkHefDRBMOG
# 2Tu9/kQCZk3taaQP9rhwz2Lo9VFKeHk2eie38+dSn5On7UOee+e03UEiifuHokYD
# Tvz0/rdkd2NfI1Jpg4L6GlPtkMyNoRdzDfTzZTlwS/Oc1np72gy8PTLQG8v1Yfx1
# CAB2vIEO+MDhXM/EEXLnG2RJ2CKadRVC9S0yOIHa9GCiurRS+1zgYSQlT7LfySmo
# c0NR2r1j1h9bm/cuG08THfdKDXF+l7f0P4TrweOjSaH6zqe/Vs+6WXZhiV9+p7SO
# Z3j5NpjhyyjaW4emii8wggUxMIIEGaADAgECAhAKoSXW1jIbfkHkBdo2l8IVMA0G
# CSqGSIb3DQEBCwUAMGUxCzAJBgNVBAYTAlVTMRUwEwYDVQQKEwxEaWdpQ2VydCBJ
# bmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5jb20xJDAiBgNVBAMTG0RpZ2lDZXJ0
# IEFzc3VyZWQgSUQgUm9vdCBDQTAeFw0xNjAxMDcxMjAwMDBaFw0zMTAxMDcxMjAw
# MDBaMHIxCzAJBgNVBAYTAlVTMRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNV
# BAsTEHd3dy5kaWdpY2VydC5jb20xMTAvBgNVBAMTKERpZ2lDZXJ0IFNIQTIgQXNz
# dXJlZCBJRCBUaW1lc3RhbXBpbmcgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
# ggEKAoIBAQC90DLuS82Pf92puoKZxTlUKFe2I0rEDgdFM1EQfdD5fU1ofue2oPSN
# s4jkl79jIZCYvxO8V9PD4X4I1moUADj3Lh477sym9jJZ/l9lP+Cb6+NGRwYaVX4L
# J37AovWg4N4iPw7/fpX786O6Ij4YrBHk8JkDbTuFfAnT7l3ImgtU46gJcWvgzyIQ
# D3XPcXJOCq3fQDpct1HhoXkUxk0kIzBdvOw8YGqsLwfM/fDqR9mIUF79Zm5WYScp
# iYRR5oLnRlD9lCosp+R1PrqYD4R/nzEU1q3V8mTLex4F0IQZchfxFwbvPc3WTe8G
# Qv2iUypPhR3EHTyvz9qsEPXdrKzpVv+TAgMBAAGjggHOMIIByjAdBgNVHQ4EFgQU
# 9LbhIB3+Ka7S5GGlsqIlssgXNW4wHwYDVR0jBBgwFoAUReuir/SSy4IxLVGLp6ch
# nfNtyA8wEgYDVR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwEwYDVR0l
# BAwwCgYIKwYBBQUHAwgweQYIKwYBBQUHAQEEbTBrMCQGCCsGAQUFBzABhhhodHRw
# Oi8vb2NzcC5kaWdpY2VydC5jb20wQwYIKwYBBQUHMAKGN2h0dHA6Ly9jYWNlcnRz
# LmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEFzc3VyZWRJRFJvb3RDQS5jcnQwgYEGA1Ud
# HwR6MHgwOqA4oDaGNGh0dHA6Ly9jcmw0LmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEFz
# c3VyZWRJRFJvb3RDQS5jcmwwOqA4oDaGNGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNv
# bS9EaWdpQ2VydEFzc3VyZWRJRFJvb3RDQS5jcmwwUAYDVR0gBEkwRzA4BgpghkgB
# hv1sAAIEMCowKAYIKwYBBQUHAgEWHGh0dHBzOi8vd3d3LmRpZ2ljZXJ0LmNvbS9D
# UFMwCwYJYIZIAYb9bAcBMA0GCSqGSIb3DQEBCwUAA4IBAQBxlRLpUYdWac3v3dp8
# qmN6s3jPBjdAhO9LhL/KzwMC/cWnww4gQiyvd/MrHwwhWiq3BTQdaq6Z+CeiZr8J
# qmDfdqQ6kw/4stHYfBli6F6CJR7Euhx7LCHi1lssFDVDBGiy23UC4HLHmNY8ZOUf
# SBAYX4k4YU1iRiSHY4yRUiyvKYnleB/WCxSlgNcSR3CzddWThZN+tpJn+1Nhiaj1
# a5bA9FhpDXzIAbG5KHW3mWOFIoxhynmUfln8jA/jb7UBJrZspe6HUSHkWGCbugwt
# K22ixH67xCUrRwIIfEmuE7bhfEJCKMYYVs9BNLZmXbZ0e/VWMyIvIjayS6JKldj1
# po5SMIIFbzCCBFegAwIBAgIQSPyTtGBVlI02p8mKidaUFjANBgkqhkiG9w0BAQwF
# ADB7MQswCQYDVQQGEwJHQjEbMBkGA1UECAwSR3JlYXRlciBNYW5jaGVzdGVyMRAw
# DgYDVQQHDAdTYWxmb3JkMRowGAYDVQQKDBFDb21vZG8gQ0EgTGltaXRlZDEhMB8G
# A1UEAwwYQUFBIENlcnRpZmljYXRlIFNlcnZpY2VzMB4XDTIxMDUyNTAwMDAwMFoX
# DTI4MTIzMTIzNTk1OVowVjELMAkGA1UEBhMCR0IxGDAWBgNVBAoTD1NlY3RpZ28g
# TGltaXRlZDEtMCsGA1UEAxMkU2VjdGlnbyBQdWJsaWMgQ29kZSBTaWduaW5nIFJv
# b3QgUjQ2MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAjeeUEiIEJHQu
# /xYjApKKtq42haxH1CORKz7cfeIxoFFvrISR41KKteKW3tCHYySJiv/vEpM7fbu2
# ir29BX8nm2tl06UMabG8STma8W1uquSggyfamg0rUOlLW7O4ZDakfko9qXGrYbNz
# szwLDO/bM1flvjQ345cbXf0fEj2CA3bm+z9m0pQxafptszSswXp43JJQ8mTHqi0E
# q8Nq6uAvp6fcbtfo/9ohq0C/ue4NnsbZnpnvxt4fqQx2sycgoda6/YDnAdLv64Ip
# lXCN/7sVz/7RDzaiLk8ykHRGa0c1E3cFM09jLrgt4b9lpwRrGNhx+swI8m2JmRCx
# rds+LOSqGLDGBwF1Z95t6WNjHjZ/aYm+qkU+blpfj6Fby50whjDoA7NAxg0POM1n
# qFOI+rgwZfpvx+cdsYN0aT6sxGg7seZnM5q2COCABUhA7vaCZEao9XOwBpXybGWf
# v1VbHJxXGsd4RnxwqpQbghesh+m2yQ6BHEDWFhcp/FycGCvqRfXvvdVnTyheBe6Q
# THrnxvTQ/PrNPjJGEyA2igTqt6oHRpwNkzoJZplYXCmjuQymMDg80EY2NXycuu7D
# 1fkKdvp+BRtAypI16dV60bV/AK6pkKrFfwGcELEW/MxuGNxvYv6mUKe4e7idFT/+
# IAx1yCJaE5UZkADpGtXChvHjjuxf9OUCAwEAAaOCARIwggEOMB8GA1UdIwQYMBaA
# FKARCiM+lvEH7OKvKe+CpX/QMKS0MB0GA1UdDgQWBBQy65Ka/zWWSC8oQEJwIDaR
# XBeF5jAOBgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zATBgNVHSUEDDAK
# BggrBgEFBQcDAzAbBgNVHSAEFDASMAYGBFUdIAAwCAYGZ4EMAQQBMEMGA1UdHwQ8
# MDowOKA2oDSGMmh0dHA6Ly9jcmwuY29tb2RvY2EuY29tL0FBQUNlcnRpZmljYXRl
# U2VydmljZXMuY3JsMDQGCCsGAQUFBwEBBCgwJjAkBggrBgEFBQcwAYYYaHR0cDov
# L29jc3AuY29tb2RvY2EuY29tMA0GCSqGSIb3DQEBDAUAA4IBAQASv6Hvi3SamES4
# aUa1qyQKDKSKZ7g6gb9Fin1SB6iNH04hhTmja14tIIa/ELiueTtTzbT72ES+Btlc
# Y2fUQBaHRIZyKtYyFfUSg8L54V0RQGf2QidyxSPiAjgaTCDi2wH3zUZPJqJ8ZsBR
# NraJAlTH/Fj7bADu/pimLpWhDFMpH2/YGaZPnvesCepdgsaLr4CnvYFIUoQx2jLs
# FeSmTD1sOXPUC4U5IOCFGmjhp0g4qdE2JXfBjRkWxYhMZn0vY86Y6GnfrDyoXZ3J
# HFuu2PMvdM+4fvbXg50RlmKarkUT2n/cR/vfw1Kf5gZV6Z2M8jpiUbzsJA8p1FiA
# hORFe1rYMIIF5zCCBE+gAwIBAgIQV3jx2H3K2UBoKgLAs4q+DzANBgkqhkiG9w0B
# AQwFADBUMQswCQYDVQQGEwJHQjEYMBYGA1UEChMPU2VjdGlnbyBMaW1pdGVkMSsw
# KQYDVQQDEyJTZWN0aWdvIFB1YmxpYyBDb2RlIFNpZ25pbmcgQ0EgUjM2MB4XDTIx
# MDYxNjAwMDAwMFoXDTI0MDYxNTIzNTk1OVowazELMAkGA1UEBhMCVVMxFzAVBgNV
# BAgMDk5vcnRoIENhcm9saW5hMQ8wDQYDVQQHDAZEdXJoYW0xGDAWBgNVBAoMD0R1
# a2UgVW5pdmVyc2l0eTEYMBYGA1UEAwwPRHVrZSBVbml2ZXJzaXR5MIIBojANBgkq
# hkiG9w0BAQEFAAOCAY8AMIIBigKCAYEA5TtZhDcnJVSGzTOOele8rZowQOs6SFuX
# mthOSWbjjX3nxPaeBsl7c19LUaWwVjgk553FOQvP6Y9FJdqfmX59aTrVVsMGkueh
# OCp3z0eFPCmHHKFtFZp/N9V7MB/got4NLAaH3i+/qBiC+xFxbnFO7C0vADKxB/Xz
# 8NYwuM778xjsAsGDjiWXoWfz+jAL1hSpe90HMyB49g9zM/dUuK88TwvzOQ5+grNa
# i8IPjkByZWRZwNw9FkQaE3vpOKAOcZcKDYqB0w2iZbb94zLUWq0TBu0Uqol5Gs9u
# COJm4PMf092Yjlm59gaFPs21wIOrV1guylMppjBJKps3mWRHk8xMh9g0Fxf/qym4
# Brf8hjsaySNWmNhPal4SxPLbHm86cPtEfzqAQOu0BWCoMRxTqsfaAlkiqKnfYmJm
# T0hWKI+WAY41oDoaaHzwyrS3dASfIwMe+069N8MWdG7aZoIruJFuS9KmRziW0/d9
# xWUWVeh4KTTyE9+mNYnSnv8chJOFu6yDAgMBAAGjggGcMIIBmDAfBgNVHSMEGDAW
# gBQPKssghyi47G9IritUpimqF6TNDDAdBgNVHQ4EFgQUcrrP/h3qRQO62fJgD7VS
# q8OHTr8wDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwEwYDVR0lBAwwCgYI
# KwYBBQUHAwMwEQYJYIZIAYb4QgEBBAQDAgQQMEoGA1UdIARDMEEwNQYMKwYBBAGy
# MQECAQMCMCUwIwYIKwYBBQUHAgEWF2h0dHBzOi8vc2VjdGlnby5jb20vQ1BTMAgG
# BmeBDAEEATBJBgNVHR8EQjBAMD6gPKA6hjhodHRwOi8vY3JsLnNlY3RpZ28uY29t
# L1NlY3RpZ29QdWJsaWNDb2RlU2lnbmluZ0NBUjM2LmNybDB5BggrBgEFBQcBAQRt
# MGswRAYIKwYBBQUHMAKGOGh0dHA6Ly9jcnQuc2VjdGlnby5jb20vU2VjdGlnb1B1
# YmxpY0NvZGVTaWduaW5nQ0FSMzYuY3J0MCMGCCsGAQUFBzABhhdodHRwOi8vb2Nz
# cC5zZWN0aWdvLmNvbTANBgkqhkiG9w0BAQwFAAOCAYEATFdpF0W9D/+wGxzMCWiZ
# U4BUlNGahmu3TG36oUt9xZ5Eyoc/tilNwc8Ro5st1hWWiaIrPw0jW8Xp334eXubp
# MDNUpH+QxkbM84oHE2uxWbrcH8YMGLC6QRgM+YHFDnfBQPX7ALqAUONOdK3oXb2V
# L8ta2mOUclnLX5gb4txWH7qAHiiNVadUgmZofA3AeysQ2dWyAML6wGOLI+Sain86
# KvG/kWaMKTAykjukgnkv3XBIiAVuXfvw61WuL8rIE3jbceyKU/eApsTwCCI2W3IP
# LliUBJBtOzUnO6gMkGPruHT1E/mseFpjsX8EFl5OsJrSl8qiRLo8TfeaFVq1GAU2
# BaNjfy5Rk1zKItRUNzU8StvvMhXcCQTMBZIz0iMcKiQomKofP+mzSSXdqOZbIFMl
# gBoYFRjAPdINNemW8bs6AYutUIg7dGuUOsl+SqA75Ock4gVaUEoiRuabmyHGAPpk
# JTURKVukdkZZKgW66++/1Jx6jpg6fHD0ZXDR87FCK0SrMIIGGjCCBAKgAwIBAgIQ
# Yh1tDFIBnjuQeRUgiSEcCjANBgkqhkiG9w0BAQwFADBWMQswCQYDVQQGEwJHQjEY
# MBYGA1UEChMPU2VjdGlnbyBMaW1pdGVkMS0wKwYDVQQDEyRTZWN0aWdvIFB1Ymxp
# YyBDb2RlIFNpZ25pbmcgUm9vdCBSNDYwHhcNMjEwMzIyMDAwMDAwWhcNMzYwMzIx
# MjM1OTU5WjBUMQswCQYDVQQGEwJHQjEYMBYGA1UEChMPU2VjdGlnbyBMaW1pdGVk
# MSswKQYDVQQDEyJTZWN0aWdvIFB1YmxpYyBDb2RlIFNpZ25pbmcgQ0EgUjM2MIIB
# ojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAmyudU/o1P45gBkNqwM/1f/bI
# U1MYyM7TbH78WAeVF3llMwsRHgBGRmxDeEDIArCS2VCoVk4Y/8j6stIkmYV5Gej4
# NgNjVQ4BYoDjGMwdjioXan1hlaGFt4Wk9vT0k2oWJMJjL9G//N523hAm4jF4UjrW
# 2pvv9+hdPX8tbbAfI3v0VdJiJPFy/7XwiunD7mBxNtecM6ytIdUlh08T2z7mJEXZ
# D9OWcJkZk5wDuf2q52PN43jc4T9OkoXZ0arWZVeffvMr/iiIROSCzKoDmWABDRzV
# /UiQ5vqsaeFaqQdzFf4ed8peNWh1OaZXnYvZQgWx/SXiJDRSAolRzZEZquE6cbcH
# 747FHncs/Kzcn0Ccv2jrOW+LPmnOyB+tAfiWu01TPhCr9VrkxsHC5qFNxaThTG5j
# 4/Kc+ODD2dX/fmBECELcvzUHf9shoFvrn35XGf2RPaNTO2uSZ6n9otv7jElspkfK
# 9qEATHZcodp+R4q2OIypxR//YEb3fkDn3UayWW9bAgMBAAGjggFkMIIBYDAfBgNV
# HSMEGDAWgBQy65Ka/zWWSC8oQEJwIDaRXBeF5jAdBgNVHQ4EFgQUDyrLIIcouOxv
# SK4rVKYpqhekzQwwDgYDVR0PAQH/BAQDAgGGMBIGA1UdEwEB/wQIMAYBAf8CAQAw
# EwYDVR0lBAwwCgYIKwYBBQUHAwMwGwYDVR0gBBQwEjAGBgRVHSAAMAgGBmeBDAEE
# ATBLBgNVHR8ERDBCMECgPqA8hjpodHRwOi8vY3JsLnNlY3RpZ28uY29tL1NlY3Rp
# Z29QdWJsaWNDb2RlU2lnbmluZ1Jvb3RSNDYuY3JsMHsGCCsGAQUFBwEBBG8wbTBG
# BggrBgEFBQcwAoY6aHR0cDovL2NydC5zZWN0aWdvLmNvbS9TZWN0aWdvUHVibGlj
# Q29kZVNpZ25pbmdSb290UjQ2LnA3YzAjBggrBgEFBQcwAYYXaHR0cDovL29jc3Au
# c2VjdGlnby5jb20wDQYJKoZIhvcNAQEMBQADggIBAAb/guF3YzZue6EVIJsT/wT+
# mHVEYcNWlXHRkT+FoetAQLHI1uBy/YXKZDk8+Y1LoNqHrp22AKMGxQtgCivnDHFy
# AQ9GXTmlk7MjcgQbDCx6mn7yIawsppWkvfPkKaAQsiqaT9DnMWBHVNIabGqgQSGT
# rQWo43MOfsPynhbz2Hyxf5XWKZpRvr3dMapandPfYgoZ8iDL2OR3sYztgJrbG6VZ
# 9DoTXFm1g0Rf97Aaen1l4c+w3DC+IkwFkvjFV3jS49ZSc4lShKK6BrPTJYs4NG1D
# GzmpToTnwoqZ8fAmi2XlZnuchC4NPSZaPATHvNIzt+z1PHo35D/f7j2pO1S8BCys
# QDHCbM5Mnomnq5aYcKCsdbh0czchOm8bkinLrYrKpii+Tk7pwL7TjRKLXkomm5D1
# Umds++pip8wH2cQpf93at3VDcOK4N7EwoIJB0kak6pSzEu4I64U6gZs7tS/dGNSl
# jf2OSSnRr7KWzq03zl8l75jy+hOds9TWSenLbjBQUGR96cFr6lEUfAIEHVC1L68Y
# 1GGxx4/eRI82ut83axHMViw1+sVpbPxg51Tbnio1lB93079WPFnYaOvfGAA0e0zc
# fF/M9gXr+korwQTh2Prqooq2bYNMvUoUKD85gnJ+t0smrWrb8dee2CvYZXD5laGt
# aAxOfy/VKNmwuWuAh9kcMYIEvTCCBLkCAQEwaDBUMQswCQYDVQQGEwJHQjEYMBYG
# A1UEChMPU2VjdGlnbyBMaW1pdGVkMSswKQYDVQQDEyJTZWN0aWdvIFB1YmxpYyBD
# b2RlIFNpZ25pbmcgQ0EgUjM2AhBXePHYfcrZQGgqAsCzir4PMAkGBSsOAwIaBQCg
# eDAYBgorBgEEAYI3AgEMMQowCKACgAChAoAAMBkGCSqGSIb3DQEJAzEMBgorBgEE
# AYI3AgEEMBwGCisGAQQBgjcCAQsxDjAMBgorBgEEAYI3AgEVMCMGCSqGSIb3DQEJ
# BDEWBBTQ8X6CMsqjdVaaiBhV8ls/XoXwyDANBgkqhkiG9w0BAQEFAASCAYApVOAC
# hjVQbgCzp0sBf5DBaV7SSfdETm6KXecv7xTozDiu6vEBDh8c/poVAxvCp0ht+CBj
# 35S9jGdP6XYtA50Cb//0hYDCAF+mRYweBUlW6rcSmOSXbiG8MAKXHDZhAkx7c8fr
# v5+/6LpQdko9eUrBYJ8m4YxdT8fJbBjV1f9tl9F4nU9jUN53RRlxuHjAc+irRuDh
# 24vm0LkhqNm2MRhaWuVm9sKAbll2c9IIXkokOL02fMqdaPufUqVHGL30e1CepR/V
# jRy/bh0mEMXUxSBOmCo4GlOavF5gGRl0EZSDioW0SfoTTliRcMaBcjIB7MClgS1H
# xUQ/ck3f8AIlxDkSuZ3gjnC93uQptm+tUujEnydQlrVeQFqPGLmPQ76UZsOfCZ6W
# UF8DY+jkukHTwlUTV4APtkbn86xqfD94PtMQKhE4/dnjX5ZcMS4bv6RyGj8Rf72U
# PqxZAGD4toMKaJcFWsx+G0eOf3Bmyjv1fSxxsysYuPWGfoAZiAqdZ9xfy3qhggIw
# MIICLAYJKoZIhvcNAQkGMYICHTCCAhkCAQEwgYYwcjELMAkGA1UEBhMCVVMxFTAT
# BgNVBAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3LmRpZ2ljZXJ0LmNvbTEx
# MC8GA1UEAxMoRGlnaUNlcnQgU0hBMiBBc3N1cmVkIElEIFRpbWVzdGFtcGluZyBD
# QQIQDUJK4L46iP9gQCHOFADw3TANBglghkgBZQMEAgEFAKBpMBgGCSqGSIb3DQEJ
# AzELBgkqhkiG9w0BBwEwHAYJKoZIhvcNAQkFMQ8XDTIxMTEwMzIyMzcwNVowLwYJ
# KoZIhvcNAQkEMSIEIMVutElAwLT2WtBfM9Hpl9tmTq5G+Ay/3dBpPh1DaFoLMA0G
# CSqGSIb3DQEBAQUABIIBAH2NJNMlPhbPT6sSEkQtZ/TV8AEythxQWalNKBY8TEBH
# UeQ8Xgd7q83dRxQ4F2jbD/EDG1NLpZGL8HEkP9NQIxYsODqJC59v8l3p5Pa1MjfH
# X1fOGuY+NXAx8QdZO0RneV15WmzbyhwPkTVdu8Wl1aib0E5X9yyogB7KjXph+M5v
# W87V4PGILkxKtfXDQV3ayMQjZCZPl/Y5XL2AI2kerMyXTJDrnxMdr5aRSLvZ5qE+
# OMMSpU19DI9XUnN5w8AvYz9KkDvxDm1k6lsl7MKaenPOuBzV30jJGSCQEZg83sv/
# TR2H5D+Lgs4zAqYUTSbMW/TIpYbqWEJxLNBN7oZ3g/k=
# SIG # End signature block
