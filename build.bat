@ECHO OFF

REM use this file to cross build static-linked edgexproxy for linux OS

SET CGO_ENABED=0
SET GOARCH=amd64
SET GOOS=linux
CD core 
go build -o edgexproxy