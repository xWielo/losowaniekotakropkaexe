package main

import (
	"embed"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	_ "golang.org/x/image/webp"
	"golang.org/x/sys/windows"
)

//go:embed all:koty
var kotyFS embed.FS

// ── Win32 constants ──────────────────────────────────────────────────────────

const (
	wsOverlapped  = uintptr(0x00000000)
	wsCaption     = uintptr(0x00C00000)
	wsSysMenu     = uintptr(0x00080000)
	wsMinimize    = uintptr(0x00020000)
	wsChild       = uintptr(0x40000000)
	wsVisible     = uintptr(0x10000000)
	wsTabStop     = uintptr(0x00010000)
	bsPushbutton  = uintptr(0x00000000)
	bsCenter      = uintptr(0x00000300)
	wsExAppWindow = uintptr(0x00040000)

	wsFixedWindow = wsOverlapped | wsCaption | wsSysMenu | wsMinimize

	csHRedraw = uint32(0x0002)
	csVRedraw = uint32(0x0001)

	swShow = uintptr(5)

	wmDestroy   = uintptr(0x0002)
	wmPaint     = uintptr(0x000F)
	wmCommand   = uintptr(0x0111)
	wmKatLoaded = uintptr(0x0401) // WM_USER+1

	dtCenter     = uintptr(0x00000001)
	dtVCenter    = uintptr(0x00000004)
	dtSingleLine = uintptr(0x00000020)
	dtWord       = uintptr(0x00000010)

	srccopy     = uintptr(0x00CC0020)
	halftone    = uintptr(4)
	transparent = uintptr(1)

	idcArrow = uintptr(32512)
	idcBtn   = uintptr(100)

	// Client-area layout
	cW       = 660
	imgH     = 460
	statusH  = 32
	btnAreaH = 52
	cH       = imgH + statusH + btnAreaH
)

// ── DLL procs ────────────────────────────────────────────────────────────────

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterClassExW = user32.NewProc("RegisterClassExW")
	pCreateWindowExW  = user32.NewProc("CreateWindowExW")
	pDefWindowProcW   = user32.NewProc("DefWindowProcW")
	pPostQuitMessage  = user32.NewProc("PostQuitMessage")
	pGetMessageW      = user32.NewProc("GetMessageW")
	pTranslateMessage = user32.NewProc("TranslateMessage")
	pDispatchMessageW = user32.NewProc("DispatchMessageW")
	pLoadCursorW      = user32.NewProc("LoadCursorW")
	pShowWindow       = user32.NewProc("ShowWindow")
	pUpdateWindow     = user32.NewProc("UpdateWindow")
	pBeginPaint       = user32.NewProc("BeginPaint")
	pEndPaint         = user32.NewProc("EndPaint")
	pFillRect         = user32.NewProc("FillRect")
	pInvalidateRect   = user32.NewProc("InvalidateRect")
	pGetClientRect    = user32.NewProc("GetClientRect")
	pEnableWindow     = user32.NewProc("EnableWindow")
	pPostMessageW     = user32.NewProc("PostMessageW")
	pDrawTextW        = user32.NewProc("DrawTextW")

	pCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	pCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	pStretchBlt         = gdi32.NewProc("StretchBlt")
	pDeleteDC           = gdi32.NewProc("DeleteDC")
	pDeleteObject       = gdi32.NewProc("DeleteObject")
	pSelectObject       = gdi32.NewProc("SelectObject")
	pSetStretchBltMode  = gdi32.NewProc("SetStretchBltMode")
	pCreateSolidBrush   = gdi32.NewProc("CreateSolidBrush")
	pSetBkMode          = gdi32.NewProc("SetBkMode")
	pSetTextColor       = gdi32.NewProc("SetTextColor")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

// ── Win32 structs ────────────────────────────────────────────────────────────

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type paintStruct struct {
	hdc         uintptr
	fErase      int32
	rcPaint     rect
	fRestore    int32
	fIncUpdate  int32
	rgbReserved [32]byte
}

type rect struct{ Left, Top, Right, Bottom int32 }

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      [2]int32
}

// BITMAPINFOHEADER — 40 bytes
type bitmapInfoHeader struct {
	biSize          uint32 // = 40
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

// ── App state ────────────────────────────────────────────────────────────────

var (
	hwndMain  uintptr
	hwndBtn   uintptr
	cats      []string
	deck      []int
	hBitmap   uintptr
	bmpW, bmpH int
	statusTxt  = "Kliknij przycisk, zeby wylosowac kota!"
	bgBrush    uintptr
)

func nextCat() int {
	if len(deck) == 0 {
		deck = rand.Perm(len(cats))
	}
	idx := deck[len(deck)-1]
	deck = deck[:len(deck)-1]
	return idx
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func listCats() []string {
	entries, _ := fs.ReadDir(kotyFS, "koty")
	exts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && exts[strings.ToLower(filepath.Ext(e.Name()))] {
			out = append(out, "koty/"+e.Name())
		}
	}
	return out
}

func imageToBitmap(img image.Image) (uintptr, int, int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	hdr := bitmapInfoHeader{
		biSize:     40,
		biWidth:    int32(w),
		biHeight:   -int32(h), // negative = top-down
		biPlanes:   1,
		biBitCount: 32,
	}

	var pBits uintptr
	hdc, _, _ := pCreateCompatibleDC.Call(0)
	hbm, _, _ := pCreateDIBSection.Call(
		hdc,
		uintptr(unsafe.Pointer(&hdr)),
		0, // DIB_RGB_COLORS
		uintptr(unsafe.Pointer(&pBits)),
		0, 0,
	)
	pDeleteDC.Call(hdc)
	if hbm == 0 || pBits == 0 {
		return 0, 0, 0
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(pBits)), w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			i := (y*w + x) * 4
			data[i+0] = byte(bl >> 8)
			data[i+1] = byte(g >> 8)
			data[i+2] = byte(r >> 8)
			data[i+3] = 0xff
		}
	}
	return hbm, w, h
}

func fitRect(iw, ih, bw, bh int) (x, y, w, h int) {
	if iw == 0 || ih == 0 {
		return 0, 0, bw, bh
	}
	scale := math.Min(float64(bw)/float64(iw), float64(bh)/float64(ih))
	w = int(float64(iw) * scale)
	h = int(float64(ih) * scale)
	x = (bw - w) / 2
	y = (bh - h) / 2
	return
}

func u16(s string) uintptr {
	p, _ := syscall.UTF16PtrFromString(s)
	return uintptr(unsafe.Pointer(p))
}

// ── Window procedure ─────────────────────────────────────────────────────────

func wndProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch message {

	case wmPaint:
		var ps paintStruct
		hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		var rc rect
		pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bgBrush)

		if hBitmap != 0 {
			x, y, dw, dh := fitRect(bmpW, bmpH, cW, imgH)
			memDC, _, _ := pCreateCompatibleDC.Call(hdc)
			pSelectObject.Call(memDC, hBitmap)
			pSetStretchBltMode.Call(hdc, halftone)
			pStretchBlt.Call(
				hdc, uintptr(x), uintptr(y), uintptr(dw), uintptr(dh),
				memDC, 0, 0, uintptr(bmpW), uintptr(bmpH),
				srccopy,
			)
			pDeleteDC.Call(memDC)
		}

		pSetBkMode.Call(hdc, transparent)
		pSetTextColor.Call(hdc, 0x00BBBBBB)
		r := rect{4, int32(imgH + 6), int32(cW - 4), int32(imgH + statusH)}
		pDrawTextW.Call(hdc, u16(statusTxt), ^uintptr(0),
			uintptr(unsafe.Pointer(&r)), dtCenter|dtVCenter|dtSingleLine)

		pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case wmCommand:
		if wParam&0xFFFF == idcBtn {
			if len(cats) == 0 {
				statusTxt = "Wrzuc zdjecia do folderu koty\\"
				pInvalidateRect.Call(hwnd, 0, 1)
				return 0
			}
			pEnableWindow.Call(hwndBtn, 0)
			statusTxt = "Losuje..."
			pInvalidateRect.Call(hwnd, 0, 1)
			go func() {
				time.Sleep(time.Second)
				idx := nextCat()
				pPostMessageW.Call(hwnd, wmKatLoaded, uintptr(idx), 0)
			}()
		}
		return 0

	case wmKatLoaded:
		idx := int(wParam)
		f, err := kotyFS.Open(cats[idx])
		if err == nil {
			img, _, decErr := image.Decode(f)
			f.Close()
			if decErr == nil {
				if hBitmap != 0 {
					pDeleteObject.Call(hBitmap)
				}
				hBitmap, bmpW, bmpH = imageToBitmap(img)
			}
		}
		statusTxt = filepath.Base(cats[idx])
		pEnableWindow.Call(hwndBtn, 1)
		pInvalidateRect.Call(hwnd, 0, 1)
		return 0

	case wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := pDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return r
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	cats = listCats()
	bgBrush, _, _ = pCreateSolidBrush.Call(0x00221a1a) // dark

	hInst, _, _ := pGetModuleHandleW.Call(0)
	cursor, _, _ := pLoadCursorW.Call(0, idcArrow)
	clsName, _ := syscall.UTF16PtrFromString("LKota")

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		style:         csHRedraw | csVRedraw,
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInst,
		hCursor:       cursor,
		hbrBackground: bgBrush,
		lpszClassName: clsName,
	}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwndMain, _, _ = pCreateWindowExW.Call(
		wsExAppWindow,
		uintptr(unsafe.Pointer(clsName)),
		u16("Losowanie Kota"),
		wsFixedWindow,
		100, 100,
		uintptr(cW+16), uintptr(cH+39), // +border/titlebar
		0, 0, hInst, 0,
	)

	const btnW = 200
	hwndBtn, _, _ = pCreateWindowExW.Call(
		0,
		u16("BUTTON"),
		u16("Losuj kota!"),
		wsChild|wsVisible|wsTabStop|bsPushbutton|bsCenter,
		uintptr((cW-btnW)/2), uintptr(imgH+statusH+8),
		btnW, 36,
		hwndMain, idcBtn, hInst, 0,
	)

	if len(cats) == 0 {
		statusTxt = "Wrzuc zdjecia do folderu koty\\ i zrestartuj"
		pEnableWindow.Call(hwndBtn, 0)
	}

	pShowWindow.Call(hwndMain, swShow)
	pUpdateWindow.Call(hwndMain)

	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
