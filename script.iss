[Setup]
AppName=Touch Sistemas - Desbravador
AppVersion=1.0.0
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
Source: ".env"; DestDir: "{app}";
Source: "touchsistemas.log"; DestDir: "{app}";

[Icons]
Name: "{group}\Touch Sistemas\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{commonprograms}\Touch Sistemas\Uninstall Touch Sistemas"; Filename: "{uninstallexe}"; WorkingDir: "{app}"; IconFilename: "{app}\touchsistemas.ico"
Name: "{userdesktop}\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; IconFilename: "{app}\touchsistemas.ico"; Tasks: desktopicon
Name: "{userstartup}\Touch Sistemas - Desbravador"; Filename: "{app}\touchsistemas-desbravador.exe"; Parameters: "/auto"; Tasks: autostarticon

[Run]
Filename: "{app}\touchsistemas-desbravador.exe"; Flags: shellexec skipifsilent nowait; Tasks: StartAfterInstall

[Code]
procedure SetFilePermissions(const FileName: String);
var
  ResultCode: Integer;
  Params: String;
begin
  if not FileExists(FileName) then
    Exit;

  Params := Format('"%s" /grant Everyone:(F)', [FileName]);
  if Exec('icacls', Params, '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
  begin
    if ResultCode <> 0 then
      MsgBox('Failed to set permissions for ' + FileName, mbError, MB_OK);
  end
  else
  begin
    MsgBox('Failed to run icacls utility', mbError, MB_OK);
  end;
end;

function ShouldRun: Boolean;
begin
  Result := True; // By default, the installer runs
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if (CurStep = ssPostInstall) and ShouldRun then
  begin
    // Set permissions after installation is complete
    SetFilePermissions(ExpandConstant('{app}\touchsistemas.log'));
  end;
end;