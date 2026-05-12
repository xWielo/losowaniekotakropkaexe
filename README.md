# Losowanie Kota

Aplikacja na Windows która losuje i wyświetla zdjęcie kota z lokalnego folderu.

## Wymagania

- [Go](https://go.dev/dl/) (jedyna zależność)

## Jak zbudować

1. Zainstaluj Go ze strony https://go.dev/dl/
2. Sklonuj lub pobierz to repozytorium
3. Wrzuć swoje zdjęcia kotów do folderu `koty\`
4. Kliknij dwukrotnie `build.bat`
5. Gotowe — pojawi się plik `LosowanieKota.exe`

## Jak używać

1. Upewnij się że w folderze `koty\` są zdjęcia
2. Uruchom `LosowanieKota.exe`
3. Klikaj przycisk **Losuj kota!**

Obsługiwane formaty: `.jpg` `.jpeg` `.png` `.gif` `.webp`

## Struktura projektu

```
├── koty\               ← tu wrzucasz zdjęcia
├── build.bat           ← buduje LosowanieKota.exe
├── main.go             ← kod źródłowy
├── go.mod / go.sum     ← zależności Go
└── LosowanieKota.exe   ← gotowy program (po zbudowaniu)
```
