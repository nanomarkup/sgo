function GenDoc {
    param (
        [string]$PackageName
    )
    $CurrLocation = Get-Location
    Set-Location -Path $PackageName
    $Status = Start-Process -FilePath 'go' -ArgumentList 'doc -all' -RedirectStandardOutput 'readme.txt' -NoNewWindow -PassThru -Wait 
    Set-Location -Path $CurrLocation
    Assert($Status.ExitCode -eq 0) 'The "go doc" command failed'
}

# Synopsis: Generate sources
task code {
    $Status = Start-Process -FilePath 'sb' -ArgumentList 'code' -NoNewWindow -PassThru -Wait
    Assert($Status.ExitCode -eq 0) 'The "code" command failed'
}

# Synopsis: Build sources
task build {
    $Status = Start-Process -FilePath 'sb' -ArgumentList 'build' -NoNewWindow -PassThru -Wait 
    Assert($Status.ExitCode -eq 0) 'The "build" command failed'
}

# Synopsis: Generate & build sources
task cbuild code, build

# Synopsis: Remove generated files
task clean {
    $AppPath = './sgo/sgo'
    if ($PSVersionTable.Platform -ne 'Unix') {
        $AppPath += '.exe'
    }
    if (Test-Path -Path $AppPath) {
        Remove-Item -Path $AppPath
    }
    if (Test-Path -Path './.test') {
        Remove-Item -Path './.test' -Recurse
    }
}

# Synopsis: Install plugin
task install {        
    $AppName = 'sgo'
    if ($PSVersionTable.Platform -eq 'Unix') {
        $GoPath = "${Env:HOME}/go"
    } else {        
        $GoPath = "${Env:GOPATH}".TrimEnd(';')
        $AppName += '.exe'
    }
    Set-Location -Path 'sgo'
    Copy-Item -Path $AppName -Destination '../bin/'
    Copy-Item -Path $AppName -Destination "$GoPath/bin/"
}

# Synopsis: Generate, build & install plugin
task cinstall cbuild, install

# Synopsis: Run tests
task test {
    $Status = Start-Process -FilePath 'go' -ArgumentList 'test' -NoNewWindow -PassThru -Wait
    Assert($Status.ExitCode -eq 0) 'The test command failed'
}

# Synopsis: Generate documentation
task doc {
    GenDoc -PackageName '.'
    GenDoc -PackageName 'helper/hashicorp/hclog'
    GenDoc -PackageName 'plugins'
    GenDoc -PackageName 'plugins/sgo'
}

task . cbuild, test, clean, doc