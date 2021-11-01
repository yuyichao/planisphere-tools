$script_date = "20210823"

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
        Continue
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
# MIIhBgYJKoZIhvcNAQcCoIIg9zCCIPMCAQExCzAJBgUrDgMCGgUAMGkGCisGAQQB
# gjcCAQSgWzBZMDQGCisGAQQBgjcCAR4wJgIDAQAABBAfzDtgWUsITrck0sYpfvNR
# AgEAAgEAAgEAAgEAAgEAMCEwCQYFKw4DAhoFAAQU1yp87xxUmw8Ii8s3MSetLE+z
# 3mSgghuzMIIE/jCCA+agAwIBAgIQDUJK4L46iP9gQCHOFADw3TANBgkqhkiG9w0B
# AQsFADByMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYD
# VQQLExB3d3cuZGlnaWNlcnQuY29tMTEwLwYDVQQDEyhEaWdpQ2VydCBTSEEyIEFz
# c3VyZWQgSUQgVGltZXN0YW1waW5nIENBMB4XDTIxMDEwMTAwMDAwMFoXDTMxMDEw
# NjAwMDAwMFowSDELMAkGA1UEBhMCVVMxFzAVBgNVBAoTDkRpZ2lDZXJ0LCBJbmMu
# MSAwHgYDVQQDExdEaWdpQ2VydCBUaW1lc3RhbXAgMjAyMTCCASIwDQYJKoZIhvcN
# AQEBBQADggEPADCCAQoCggEBAMLmYYRnxYr1DQikRcpja1HXOhFCvQp1dU2UtAxQ
# tSYQ/h3Ib5FrDJbnGlxI70Tlv5thzRWRYlq4/2cLnGP9NmqB+in43Stwhd4CGPN4
# bbx9+cdtCT2+anaH6Yq9+IRdHnbJ5MZ2djpT0dHTWjaPxqPhLxs6t2HWc+xObTOK
# fF1FLUuxUOZBOjdWhtyTI433UCXoZObd048vV7WHIOsOjizVI9r0TXhG4wODMSlK
# XAwxikqMiMX3MFr5FK8VX2xDSQn9JiNT9o1j6BqrW7EdMMKbaYK02/xWVLwfoYer
# vnpbCiAvSwnJlaeNsvrWY4tOpXIc7p96AXP4Gdb+DUmEvQECAwEAAaOCAbgwggG0
# MA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMBYGA1UdJQEB/wQMMAoGCCsG
# AQUFBwMIMEEGA1UdIAQ6MDgwNgYJYIZIAYb9bAcBMCkwJwYIKwYBBQUHAgEWG2h0
# dHA6Ly93d3cuZGlnaWNlcnQuY29tL0NQUzAfBgNVHSMEGDAWgBT0tuEgHf4prtLk
# YaWyoiWyyBc1bjAdBgNVHQ4EFgQUNkSGjqS6sGa+vCgtHUQ23eNqerwwcQYDVR0f
# BGowaDAyoDCgLoYsaHR0cDovL2NybDMuZGlnaWNlcnQuY29tL3NoYTItYXNzdXJl
# ZC10cy5jcmwwMqAwoC6GLGh0dHA6Ly9jcmw0LmRpZ2ljZXJ0LmNvbS9zaGEyLWFz
# c3VyZWQtdHMuY3JsMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGGGGh0dHA6
# Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2NhY2VydHMu
# ZGlnaWNlcnQuY29tL0RpZ2lDZXJ0U0hBMkFzc3VyZWRJRFRpbWVzdGFtcGluZ0NB
# LmNydDANBgkqhkiG9w0BAQsFAAOCAQEASBzctemaI7znGucgDo5nRv1CclF0CiNH
# o6uS0iXEcFm+FKDlJ4GlTRQVGQd58NEEw4bZO73+RAJmTe1ppA/2uHDPYuj1UUp4
# eTZ6J7fz51Kfk6ftQ55757TdQSKJ+4eiRgNO/PT+t2R3Y18jUmmDgvoaU+2QzI2h
# F3MN9PNlOXBL85zWenvaDLw9MtAby/Vh/HUIAHa8gQ74wOFcz8QRcucbZEnYIpp1
# FUL1LTI4gdr0YKK6tFL7XOBhJCVPst/JKahzQ1HavWPWH1ub9y4bTxMd90oNcX6X
# t/Q/hOvB46NJofrOp79Wz7pZdmGJX36ntI5nePk2mOHLKNpbh6aKLzCCBTEwggQZ
# oAMCAQICEAqhJdbWMht+QeQF2jaXwhUwDQYJKoZIhvcNAQELBQAwZTELMAkGA1UE
# BhMCVVMxFTATBgNVBAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3LmRpZ2lj
# ZXJ0LmNvbTEkMCIGA1UEAxMbRGlnaUNlcnQgQXNzdXJlZCBJRCBSb290IENBMB4X
# DTE2MDEwNzEyMDAwMFoXDTMxMDEwNzEyMDAwMFowcjELMAkGA1UEBhMCVVMxFTAT
# BgNVBAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3LmRpZ2ljZXJ0LmNvbTEx
# MC8GA1UEAxMoRGlnaUNlcnQgU0hBMiBBc3N1cmVkIElEIFRpbWVzdGFtcGluZyBD
# QTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL3QMu5LzY9/3am6gpnF
# OVQoV7YjSsQOB0UzURB90Pl9TWh+57ag9I2ziOSXv2MhkJi/E7xX08PhfgjWahQA
# OPcuHjvuzKb2Mln+X2U/4Jvr40ZHBhpVfgsnfsCi9aDg3iI/Dv9+lfvzo7oiPhis
# EeTwmQNtO4V8CdPuXciaC1TjqAlxa+DPIhAPdc9xck4Krd9AOly3UeGheRTGTSQj
# MF287DxgaqwvB8z98OpH2YhQXv1mblZhJymJhFHmgudGUP2UKiyn5HU+upgPhH+f
# MRTWrdXyZMt7HgXQhBlyF/EXBu89zdZN7wZC/aJTKk+FHcQdPK/P2qwQ9d2srOlW
# /5MCAwEAAaOCAc4wggHKMB0GA1UdDgQWBBT0tuEgHf4prtLkYaWyoiWyyBc1bjAf
# BgNVHSMEGDAWgBRF66Kv9JLLgjEtUYunpyGd823IDzASBgNVHRMBAf8ECDAGAQH/
# AgEAMA4GA1UdDwEB/wQEAwIBhjATBgNVHSUEDDAKBggrBgEFBQcDCDB5BggrBgEF
# BQcBAQRtMGswJAYIKwYBBQUHMAGGGGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBD
# BggrBgEFBQcwAoY3aHR0cDovL2NhY2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0
# QXNzdXJlZElEUm9vdENBLmNydDCBgQYDVR0fBHoweDA6oDigNoY0aHR0cDovL2Ny
# bDQuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0QXNzdXJlZElEUm9vdENBLmNybDA6oDig
# NoY0aHR0cDovL2NybDMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0QXNzdXJlZElEUm9v
# dENBLmNybDBQBgNVHSAESTBHMDgGCmCGSAGG/WwAAgQwKjAoBggrBgEFBQcCARYc
# aHR0cHM6Ly93d3cuZGlnaWNlcnQuY29tL0NQUzALBglghkgBhv1sBwEwDQYJKoZI
# hvcNAQELBQADggEBAHGVEulRh1Zpze/d2nyqY3qzeM8GN0CE70uEv8rPAwL9xafD
# DiBCLK938ysfDCFaKrcFNB1qrpn4J6JmvwmqYN92pDqTD/iy0dh8GWLoXoIlHsS6
# HHssIeLWWywUNUMEaLLbdQLgcseY1jxk5R9IEBhfiThhTWJGJIdjjJFSLK8pieV4
# H9YLFKWA1xJHcLN11ZOFk362kmf7U2GJqPVrlsD0WGkNfMgBsbkodbeZY4UijGHK
# eZR+WfyMD+NvtQEmtmyl7odRIeRYYJu6DC0rbaLEfrvEJStHAgh8Sa4TtuF8QkIo
# xhhWz0E0tmZdtnR79VYzIi8iNrJLokqV2PWmjlIwggVvMIIEV6ADAgECAhBI/JO0
# YFWUjTanyYqJ1pQWMA0GCSqGSIb3DQEBDAUAMHsxCzAJBgNVBAYTAkdCMRswGQYD
# VQQIDBJHcmVhdGVyIE1hbmNoZXN0ZXIxEDAOBgNVBAcMB1NhbGZvcmQxGjAYBgNV
# BAoMEUNvbW9kbyBDQSBMaW1pdGVkMSEwHwYDVQQDDBhBQUEgQ2VydGlmaWNhdGUg
# U2VydmljZXMwHhcNMjEwNTI1MDAwMDAwWhcNMjgxMjMxMjM1OTU5WjBWMQswCQYD
# VQQGEwJHQjEYMBYGA1UEChMPU2VjdGlnbyBMaW1pdGVkMS0wKwYDVQQDEyRTZWN0
# aWdvIFB1YmxpYyBDb2RlIFNpZ25pbmcgUm9vdCBSNDYwggIiMA0GCSqGSIb3DQEB
# AQUAA4ICDwAwggIKAoICAQCN55QSIgQkdC7/FiMCkoq2rjaFrEfUI5ErPtx94jGg
# UW+shJHjUoq14pbe0IdjJImK/+8Skzt9u7aKvb0Ffyeba2XTpQxpsbxJOZrxbW6q
# 5KCDJ9qaDStQ6Utbs7hkNqR+Sj2pcaths3OzPAsM79szV+W+NDfjlxtd/R8SPYID
# dub7P2bSlDFp+m2zNKzBenjcklDyZMeqLQSrw2rq4C+np9xu1+j/2iGrQL+57g2e
# xtmeme/G3h+pDHazJyCh1rr9gOcB0u/rgimVcI3/uxXP/tEPNqIuTzKQdEZrRzUT
# dwUzT2MuuC3hv2WnBGsY2HH6zAjybYmZELGt2z4s5KoYsMYHAXVn3m3pY2MeNn9p
# ib6qRT5uWl+PoVvLnTCGMOgDs0DGDQ84zWeoU4j6uDBl+m/H5x2xg3RpPqzEaDux
# 5mczmrYI4IAFSEDu9oJkRqj1c7AGlfJsZZ+/VVscnFcax3hGfHCqlBuCF6yH6bbJ
# DoEcQNYWFyn8XJwYK+pF9e+91WdPKF4F7pBMeufG9ND8+s0+MkYTIDaKBOq3qgdG
# nA2TOglmmVhcKaO5DKYwODzQRjY1fJy67sPV+Qp2+n4FG0DKkjXp1XrRtX8ArqmQ
# qsV/AZwQsRb8zG4Y3G9i/qZQp7h7uJ0VP/4gDHXIIloTlRmQAOka1cKG8eOO7F/0
# 5QIDAQABo4IBEjCCAQ4wHwYDVR0jBBgwFoAUoBEKIz6W8Qfs4q8p74Klf9AwpLQw
# HQYDVR0OBBYEFDLrkpr/NZZILyhAQnAgNpFcF4XmMA4GA1UdDwEB/wQEAwIBhjAP
# BgNVHRMBAf8EBTADAQH/MBMGA1UdJQQMMAoGCCsGAQUFBwMDMBsGA1UdIAQUMBIw
# BgYEVR0gADAIBgZngQwBBAEwQwYDVR0fBDwwOjA4oDagNIYyaHR0cDovL2NybC5j
# b21vZG9jYS5jb20vQUFBQ2VydGlmaWNhdGVTZXJ2aWNlcy5jcmwwNAYIKwYBBQUH
# AQEEKDAmMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5jb21vZG9jYS5jb20wDQYJ
# KoZIhvcNAQEMBQADggEBABK/oe+LdJqYRLhpRrWrJAoMpIpnuDqBv0WKfVIHqI0f
# TiGFOaNrXi0ghr8QuK55O1PNtPvYRL4G2VxjZ9RAFodEhnIq1jIV9RKDwvnhXRFA
# Z/ZCJ3LFI+ICOBpMIOLbAffNRk8monxmwFE2tokCVMf8WPtsAO7+mKYulaEMUykf
# b9gZpk+e96wJ6l2CxouvgKe9gUhShDHaMuwV5KZMPWw5c9QLhTkg4IUaaOGnSDip
# 0TYld8GNGRbFiExmfS9jzpjoad+sPKhdnckcW67Y8y90z7h+9teDnRGWYpquRRPa
# f9xH+9/DUp/mBlXpnYzyOmJRvOwkDynUWICE5EV7WtgwggXnMIIET6ADAgECAhBX
# ePHYfcrZQGgqAsCzir4PMA0GCSqGSIb3DQEBDAUAMFQxCzAJBgNVBAYTAkdCMRgw
# FgYDVQQKEw9TZWN0aWdvIExpbWl0ZWQxKzApBgNVBAMTIlNlY3RpZ28gUHVibGlj
# IENvZGUgU2lnbmluZyBDQSBSMzYwHhcNMjEwNjE2MDAwMDAwWhcNMjQwNjE1MjM1
# OTU5WjBrMQswCQYDVQQGEwJVUzEXMBUGA1UECAwOTm9ydGggQ2Fyb2xpbmExDzAN
# BgNVBAcMBkR1cmhhbTEYMBYGA1UECgwPRHVrZSBVbml2ZXJzaXR5MRgwFgYDVQQD
# DA9EdWtlIFVuaXZlcnNpdHkwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIB
# gQDlO1mENyclVIbNM456V7ytmjBA6zpIW5ea2E5JZuONfefE9p4GyXtzX0tRpbBW
# OCTnncU5C8/pj0Ul2p+Zfn1pOtVWwwaS56E4KnfPR4U8KYccoW0Vmn831XswH+Ci
# 3g0sBofeL7+oGIL7EXFucU7sLS8AMrEH9fPw1jC4zvvzGOwCwYOOJZehZ/P6MAvW
# FKl73QczIHj2D3Mz91S4rzxPC/M5Dn6Cs1qLwg+OQHJlZFnA3D0WRBoTe+k4oA5x
# lwoNioHTDaJltv3jMtRarRMG7RSqiXkaz24I4mbg8x/T3ZiOWbn2BoU+zbXAg6tX
# WC7KUymmMEkqmzeZZEeTzEyH2DQXF/+rKbgGt/yGOxrJI1aY2E9qXhLE8tsebzpw
# +0R/OoBA67QFYKgxHFOqx9oCWSKoqd9iYmZPSFYoj5YBjjWgOhpofPDKtLd0BJ8j
# Ax77Tr03wxZ0btpmgiu4kW5L0qZHOJbT933FZRZV6HgpNPIT36Y1idKe/xyEk4W7
# rIMCAwEAAaOCAZwwggGYMB8GA1UdIwQYMBaAFA8qyyCHKLjsb0iuK1SmKaoXpM0M
# MB0GA1UdDgQWBBRyus/+HepFA7rZ8mAPtVKrw4dOvzAOBgNVHQ8BAf8EBAMCB4Aw
# DAYDVR0TAQH/BAIwADATBgNVHSUEDDAKBggrBgEFBQcDAzARBglghkgBhvhCAQEE
# BAMCBBAwSgYDVR0gBEMwQTA1BgwrBgEEAbIxAQIBAwIwJTAjBggrBgEFBQcCARYX
# aHR0cHM6Ly9zZWN0aWdvLmNvbS9DUFMwCAYGZ4EMAQQBMEkGA1UdHwRCMEAwPqA8
# oDqGOGh0dHA6Ly9jcmwuc2VjdGlnby5jb20vU2VjdGlnb1B1YmxpY0NvZGVTaWdu
# aW5nQ0FSMzYuY3JsMHkGCCsGAQUFBwEBBG0wazBEBggrBgEFBQcwAoY4aHR0cDov
# L2NydC5zZWN0aWdvLmNvbS9TZWN0aWdvUHVibGljQ29kZVNpZ25pbmdDQVIzNi5j
# cnQwIwYIKwYBBQUHMAGGF2h0dHA6Ly9vY3NwLnNlY3RpZ28uY29tMA0GCSqGSIb3
# DQEBDAUAA4IBgQBMV2kXRb0P/7AbHMwJaJlTgFSU0ZqGa7dMbfqhS33FnkTKhz+2
# KU3BzxGjmy3WFZaJois/DSNbxenffh5e5ukwM1Skf5DGRszzigcTa7FZutwfxgwY
# sLpBGAz5gcUOd8FA9fsAuoBQ4050rehdvZUvy1raY5RyWctfmBvi3FYfuoAeKI1V
# p1SCZmh8DcB7KxDZ1bIAwvrAY4sj5JqKfzoq8b+RZowpMDKSO6SCeS/dcEiIBW5d
# +/DrVa4vysgTeNtx7IpT94CmxPAIIjZbcg8uWJQEkG07NSc7qAyQY+u4dPUT+ax4
# WmOxfwQWXk6wmtKXyqJEujxN95oVWrUYBTYFo2N/LlGTXMoi1FQ3NTxK2+8yFdwJ
# BMwFkjPSIxwqJCiYqh8/6bNJJd2o5lsgUyWAGhgVGMA90g016ZbxuzoBi61QiDt0
# a5Q6yX5KoDvk5yTiBVpQSiJG5pubIcYA+mQlNREpW6R2RlkqBbrr77/UnHqOmDp8
# cPRlcNHzsUIrRKswggYaMIIEAqADAgECAhBiHW0MUgGeO5B5FSCJIRwKMA0GCSqG
# SIb3DQEBDAUAMFYxCzAJBgNVBAYTAkdCMRgwFgYDVQQKEw9TZWN0aWdvIExpbWl0
# ZWQxLTArBgNVBAMTJFNlY3RpZ28gUHVibGljIENvZGUgU2lnbmluZyBSb290IFI0
# NjAeFw0yMTAzMjIwMDAwMDBaFw0zNjAzMjEyMzU5NTlaMFQxCzAJBgNVBAYTAkdC
# MRgwFgYDVQQKEw9TZWN0aWdvIExpbWl0ZWQxKzApBgNVBAMTIlNlY3RpZ28gUHVi
# bGljIENvZGUgU2lnbmluZyBDQSBSMzYwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAw
# ggGKAoIBgQCbK51T+jU/jmAGQ2rAz/V/9shTUxjIztNsfvxYB5UXeWUzCxEeAEZG
# bEN4QMgCsJLZUKhWThj/yPqy0iSZhXkZ6Pg2A2NVDgFigOMYzB2OKhdqfWGVoYW3
# haT29PSTahYkwmMv0b/83nbeECbiMXhSOtbam+/36F09fy1tsB8je/RV0mIk8XL/
# tfCK6cPuYHE215wzrK0h1SWHTxPbPuYkRdkP05ZwmRmTnAO5/arnY83jeNzhP06S
# hdnRqtZlV59+8yv+KIhE5ILMqgOZYAENHNX9SJDm+qxp4VqpB3MV/h53yl41aHU5
# pledi9lCBbH9JeIkNFICiVHNkRmq4TpxtwfvjsUedyz8rNyfQJy/aOs5b4s+ac7I
# H60B+Ja7TVM+EKv1WuTGwcLmoU3FpOFMbmPj8pz44MPZ1f9+YEQIQty/NQd/2yGg
# W+ufflcZ/ZE9o1M7a5Jnqf2i2/uMSWymR8r2oQBMdlyh2n5HirY4jKnFH/9gRvd+
# QOfdRrJZb1sCAwEAAaOCAWQwggFgMB8GA1UdIwQYMBaAFDLrkpr/NZZILyhAQnAg
# NpFcF4XmMB0GA1UdDgQWBBQPKssghyi47G9IritUpimqF6TNDDAOBgNVHQ8BAf8E
# BAMCAYYwEgYDVR0TAQH/BAgwBgEB/wIBADATBgNVHSUEDDAKBggrBgEFBQcDAzAb
# BgNVHSAEFDASMAYGBFUdIAAwCAYGZ4EMAQQBMEsGA1UdHwREMEIwQKA+oDyGOmh0
# dHA6Ly9jcmwuc2VjdGlnby5jb20vU2VjdGlnb1B1YmxpY0NvZGVTaWduaW5nUm9v
# dFI0Ni5jcmwwewYIKwYBBQUHAQEEbzBtMEYGCCsGAQUFBzAChjpodHRwOi8vY3J0
# LnNlY3RpZ28uY29tL1NlY3RpZ29QdWJsaWNDb2RlU2lnbmluZ1Jvb3RSNDYucDdj
# MCMGCCsGAQUFBzABhhdodHRwOi8vb2NzcC5zZWN0aWdvLmNvbTANBgkqhkiG9w0B
# AQwFAAOCAgEABv+C4XdjNm57oRUgmxP/BP6YdURhw1aVcdGRP4Wh60BAscjW4HL9
# hcpkOTz5jUug2oeunbYAowbFC2AKK+cMcXIBD0ZdOaWTsyNyBBsMLHqafvIhrCym
# laS98+QpoBCyKppP0OcxYEdU0hpsaqBBIZOtBajjcw5+w/KeFvPYfLF/ldYpmlG+
# vd0xqlqd099iChnyIMvY5HexjO2AmtsbpVn0OhNcWbWDRF/3sBp6fWXhz7DcML4i
# TAWS+MVXeNLj1lJziVKEoroGs9Mlizg0bUMbOalOhOfCipnx8CaLZeVme5yELg09
# Jlo8BMe80jO37PU8ejfkP9/uPak7VLwELKxAMcJszkyeiaerlphwoKx1uHRzNyE6
# bxuSKcutisqmKL5OTunAvtONEoteSiabkPVSZ2z76mKnzAfZxCl/3dq3dUNw4rg3
# sTCggkHSRqTqlLMS7gjrhTqBmzu1L90Y1KWN/Y5JKdGvspbOrTfOXyXvmPL6E52z
# 1NZJ6ctuMFBQZH3pwWvqURR8AgQdULUvrxjUYbHHj95Ejza63zdrEcxWLDX6xWls
# /GDnVNueKjWUH3fTv1Y8Wdho698YADR7TNx8X8z2Bev6SivBBOHY+uqiirZtg0y9
# ShQoPzmCcn63Syatatvx157YK9hlcPmVoa1oDE5/L9Uo2bC5a4CH2RwxggS9MIIE
# uQIBATBoMFQxCzAJBgNVBAYTAkdCMRgwFgYDVQQKEw9TZWN0aWdvIExpbWl0ZWQx
# KzApBgNVBAMTIlNlY3RpZ28gUHVibGljIENvZGUgU2lnbmluZyBDQSBSMzYCEFd4
# 8dh9ytlAaCoCwLOKvg8wCQYFKw4DAhoFAKB4MBgGCisGAQQBgjcCAQwxCjAIoAKA
# AKECgAAwGQYJKoZIhvcNAQkDMQwGCisGAQQBgjcCAQQwHAYKKwYBBAGCNwIBCzEO
# MAwGCisGAQQBgjcCARUwIwYJKoZIhvcNAQkEMRYEFOvetkwndjO1RcCTGexOW909
# FcvmMA0GCSqGSIb3DQEBAQUABIIBgDgg/tvAZNjKQhjqVCMyr7XF0v9SJT2Jx6mS
# QaYROvyMj2S5VLbfocVa1E4llfre4sWboL85Vo7dr2DIV7yMqfXRmiGS4oskoLYX
# GvXKALEqncQLa9dBwM4l8KtKIFXTOkz888bml+TSX0ZH4mFV4dWShT7KNBMc7cTB
# 1a8Kn1Dv8V+LN7/0v3T7VLPhWHAx9ag17KLgkII/rNyC9wEt7Z4be3dka20HZPQQ
# VDgRZ0wG5aU4X8Qp6SotUf5PBqZ8VDZC4scHnN2LokE2jP5gryCDnKj2XgEcljLS
# +vZXsFUWu7RZEwzpwH7PGELkbv7q3rlSfR8GHnUxQrCh4yWdywjDBJRmYXf4YrgK
# lH7IepbCwUb6rse8eslTTiLbebxza1YlvHefYE1eaPPx6AN/Bb3hKoDT2ZF5QJWd
# O1AR/jIgDQ00jR/F3Xjcv6IioXgiOEa1w/JlTufiA4rPoNp8o6KyCcwO/JLhzh4O
# PM7M+IqWO/3FcEmKOoSPrkIsax200KGCAjAwggIsBgkqhkiG9w0BCQYxggIdMIIC
# GQIBATCBhjByMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkw
# FwYDVQQLExB3d3cuZGlnaWNlcnQuY29tMTEwLwYDVQQDEyhEaWdpQ2VydCBTSEEy
# IEFzc3VyZWQgSUQgVGltZXN0YW1waW5nIENBAhANQkrgvjqI/2BAIc4UAPDdMA0G
# CWCGSAFlAwQCAQUAoGkwGAYJKoZIhvcNAQkDMQsGCSqGSIb3DQEHATAcBgkqhkiG
# 9w0BCQUxDxcNMjExMTAxMjE1NzQ0WjAvBgkqhkiG9w0BCQQxIgQgMUrPh3va+MQQ
# 2jkZp3+C38xYV63Qsc1w2Xad78ZJ40owDQYJKoZIhvcNAQEBBQAEggEAXGt6AFiD
# A5KbKX0FNdSoCkwWz6fRzHza/YGmFaIr93w310XcvtSDJBQx0rC2UQP8syjf0MfG
# zs9LYYPMEhsO2o4CTn2haN2b9VP/GBKlmLOITX1sAkUF/xfCh1Z+6tjon7cbFMQI
# cZ6P77aquftGnqF072rFe5sZKZfnSq5GkwtJIS3f4ihv5oGT0NP9C+UCGYw22+Mx
# zqQ5yiSXKJK6mp1hcAR+OOLqVZoGZlUiKkjQk1fNLDZLIp4VRM8XWF5DilHG6lUg
# pufMFnr5mWyUFLQ11/16bWvP254jUMlhkhbzq7sXxFBAKcXfniq4EgDwcXaTqvmP
# AfaNxcM4OGbFRw==
# SIG # End signature block
