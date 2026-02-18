@echo off
set PUB_CACHE=%~dp0..\dart\.pub-cache
call %PUB_CACHE%\bin\protoc-gen-dart.bat %*
