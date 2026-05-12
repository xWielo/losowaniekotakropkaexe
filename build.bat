@echo off
chcp 65001 >nul
echo.
echo  ==========================
echo   Budowanie LosowanieKota
echo  ==========================
echo.

where go >nul 2>&1
if errorlevel 1 (
    echo BLAD: Go nie jest zainstalowane!
    echo Pobierz Go z: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo Pobieranie zaleznosci...
go mod tidy
if errorlevel 1 (
    echo.
    echo BLAD: Nie mozna pobrac zaleznosci. Sprawdz polaczenie z internetem.
    pause
    exit /b 1
)

echo Kompilowanie...
set CGO_ENABLED=0
go build -ldflags "-H windowsgui" -o LosowanieKota.exe .
if errorlevel 1 (
    echo.
    echo BLAD: Kompilacja nie powiodla sie.
    pause
    exit /b 1
)

echo.
echo  Gotowe! Plik: LosowanieKota.exe
echo  Wrzuc zdjecia kotow do folderu koty\
echo  i uruchom LosowanieKota.exe
echo.
pause
