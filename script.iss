[Setup]
AppName=Touch Sistemas - Desbravador
AppVersion=0.1.0
WizardStyle=modern
DefaultDirName={pf}\Touch Sistemas
DefaultGroupName=Touch Sistemas
UninstallDisplayIcon={app}\touchsistemas-desbravador.exe
OutputBaseFilename=TouchSistemasDesbravadorInstaller
Compression=lzma2
SolidCompression=yes
OutputDir=output

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}";  GroupDescription: "{cm:AdditionalIcons}";
Name: "autostarticon"; Description: "{cm:AutoStartProgram, Touch Sistemas - Desbravador}"; GroupDescription: "{cm:AdditionalIcons}";
Name: "StartAfterInstall"; Description: "Run application after install"

[Files]
Source: "touchsistemas-desbravador.exe"; DestDir: "{app}"
Source: "touchsistemas.ico"; DestDir: "{app}";

[Icons]
Name: "{group}\Touch Sistemas\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{commonprograms}\Touch Sistemas\Uninstall Touch Sistemas"; Filename: "{uninstallexe}"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{userdesktop}\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; IconFilename: "{app}\touchsistemas.ico"; Tasks: desktopicon
Name: "{userstartup}\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; Parameters: "/auto"; Tasks: autostarticon

[Run]
Filename: "{app}\touchsistemas-desbravador.exe"; Flags: shellexec skipifsilent nowait; Tasks: StartAfterInstall