$script_date = "20211102"

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
