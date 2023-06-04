[Setup]
AppName=Touch Sistemas
AppVersion=1.0
DefaultDirName={pf}\Touch Sistemas
DefaultGroupName=Touch Sistemas
OutputDir=output
OutputBaseFilename=TouchSistemasInstaller
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=admin

[Files]
Source: "touchsistemas.exe"; DestDir: "{app}"
Source: "touchsistemas.ico"; DestDir: "{app}"; Flags: onlyifdoesntexist

[Icons]
Name: "{group}\Touch Sistemas\Touch Sistemas"; Filename: "{app}\touchsistemas.exe"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{commonprograms}\Touch Sistemas\Uninstall Touch Sistemas"; Filename: "{uninstallexe}"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
