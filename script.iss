[Setup]
AppName=Touch Sistemas - Desbravador
AppVersion=0.1
DefaultDirName={pf}\Touch Sistemas
DefaultGroupName=Touch Sistemas
OutputDir=output
OutputBaseFilename=TouchSistemasInstaller
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=admin

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}";  GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "touchsistemas-desbravador.exe"; DestDir: "{app}"
Source: "touchsistemas.ico"; DestDir: "{app}";

[Icons]
Name: "{group}\Touch Sistemas\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{commonprograms}\Touch Sistemas\Uninstall Touch Sistemas"; Filename: "{uninstallexe}"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{userdesktop}\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; IconFilename: "{app}\touchsistemas.ico"; Tasks: desktopicon