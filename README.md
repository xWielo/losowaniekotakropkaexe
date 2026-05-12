# Losowanie Kota

Aplikacja na Windows która losuje i wyświetla zdjęcie kota.  
Zdjęcia są **wbudowane w plik .exe** podczas kompilacji — gotowy program nie potrzebuje żadnych zewnętrznych plików.

## Wymagania

- [Go](https://go.dev/dl/) (jedyna zależność, tylko do budowania)

## Jak zbudować

1. Zainstaluj Go ze strony https://go.dev/dl/
2. Wrzuć zdjęcia kotów do folderu `koty\`
3. Kliknij dwukrotnie `build.bat`
4. Gotowe — pojawi się plik `LosowanieKota.exe`

Zdjęcia są **na stałe wbudowane** w exe. Możesz go przenieść gdziekolwiek — nie potrzebuje folderu `koty\` do działania.  
Chcesz dodać nowe zdjęcia? Wrzuć je do `koty\` i zbuduj ponownie.

## Jak używać

1. Uruchom `LosowanieKota.exe`
2. Klikaj przycisk **Losuj kota!**

Obsługiwane formaty: `.jpg` `.jpeg` `.png` `.gif` `.webp`

## Struktura projektu

```
├── koty\               ← tu wrzucasz zdjęcia PRZED zbudowaniem
├── build.bat           ← buduje LosowanieKota.exe
├── main.go             ← kod źródłowy
├── go.mod / go.sum     ← zależności Go
└── LosowanieKota.exe   ← gotowy program (po zbudowaniu)
```
