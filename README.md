rsrc -arch amd64 -ico icon.ico -manifest postos.manifest -o rsrc.syso

go build -ldflags="-H windowsgui"  -o touchsistemas.exe

env GO111MODULE=on go build -ldflags "-H=windowsgui"
-----




go build -ldflags "-s -w" -o postos.exe
 
go build -ldflags -H=windowsgui "-s -w" -o postos.exe