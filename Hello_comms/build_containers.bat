Rem ECHO OFF
setlocal enabledelayedexpansion

set c=node
for /l %%x in (5001, 1, 5005) do (
    set /A t=%%x-5000
    set n=%c%!t!
    docker run -i -t -d --rm --name !n! kadlab %%x
)
