$ErrorActionPreference = 'Stop';

$packageName= 'functions'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url        = 'https://github.com/iron-io/functions/releases/download/0.2.57/fn.exe'
#$fileLocation = Join-Path $toolsDir 'NAME_OF_EMBEDDED_INSTALLER_FILE'
#$fileLocation = '\\SHARE_LOCATION\to\INSTALLER_FILE'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  fileType      = 'EXE'
  url           = $url
  #file         = $fileLocation

  softwareName  = 'Iron.io Functions*'

  checksum      = 'D4E0628696732C8A237C6D88FC7D409551D4D223C4ED0752880673FAD6D33319'
  checksumType  = 'sha256'

  validExitCodes= @(0, 3010, 1641)
}

Install-ChocolateyPackage @packageArgs