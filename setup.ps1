# Setup PowerShell
# Please run this script with admin rights
$ErrorActionPreference = 'Stop'
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope LocalMachine
# Install PowerShell Modules
Install-Module InvokeBuild -Force