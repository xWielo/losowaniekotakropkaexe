@echo off
chcp 65001 >nul
echo.
echo  =========================
echo   Budowanie LosowanieKota
echo  =========================
echo.

where go >nul 2>&1
if errorlevel 1 (
    echo BLAD: Go nie jest zainstalowane!
    echo Pobierz Go z: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

where gcc >nul 2>&1
if errorlevel 1 (
    echo BLAD: Brak GCC ^(wymagane przez Fyne^)
    echo.
    echo Zainstaluj MSYS2: https://www.msys2.org/
    echo Potem w terminalu MSYS2 uruchom:
    echo   pacman -S mingw-w64-x86_64-gcc
    echo.
    echo Nastepnie dodaj do PATH:
    echo   C:\msys64\mingw64\bin
    echo.
    pause
    exit /b 1
)

echo Pobieranie zaleznosci...
go mod tidy
if errorlevel 1 (
    echo.
    echo BLAD: Nie mozna pobrac zaleznosci. Sprawdz internet.
    pause
    exit /b 1
)

echo.
echo Kompilowanie...
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
