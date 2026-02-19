//go:build windows

package tray

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"

	agentsync "devproxy-agent/sync"
)

var (
	// Windows DLLs
	shell32  = syscall.NewLazyDLL("shell32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// Shell32 functions
	pShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")

	// User32 functions
	pRegisterClassEx     = user32.NewProc("RegisterClassExW")
	pCreateWindowEx      = user32.NewProc("CreateWindowExW")
	pDefWindowProc       = user32.NewProc("DefWindowProcW")
	pGetMessage          = user32.NewProc("GetMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessage     = user32.NewProc("DispatchMessageW")
	pPostQuitMessage     = user32.NewProc("PostQuitMessage")
	pCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	pAppendMenu          = user32.NewProc("AppendMenuW")
	pTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	pDestroyMenu         = user32.NewProc("DestroyMenu")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	pGetCursorPos        = user32.NewProc("GetCursorPos")
	pPostMessage         = user32.NewProc("PostMessageW")
	pLoadIcon            = user32.NewProc("LoadIconW")

	// Kernel32
	pGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmApp           = 0x8000
	wmTrayIcon      = wmApp + 1
	wmCommand       = 0x0111
	wmDestroy       = 0x0002
	wmRButtonUp     = 0x0205
	wmLButtonDblClk = 0x0203

	nimAdd     = 0x00000000
	nimDelete  = 0x00000002
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifMessage = 0x00000001

	idOpenConfig = 1001
	idSyncNow    = 1002
	idPause      = 1003
	idQuit       = 1004

	tpmLeftAlign   = 0x0000
	tpmBottomAlign = 0x0020

	mfString    = 0x00000000
	mfSeparator = 0x00000800

	idiApplication = 32512

	csVRedraw = 0x0001
	csHRedraw = 0x0002
)

type wndClassEx struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   syscall.Handle
	icon       syscall.Handle
	cursor     syscall.Handle
	background syscall.Handle
	menuName   *uint16
	className  *uint16
	iconSm     syscall.Handle
}

type point struct {
	x, y int32
}

type msg struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             syscall.Handle
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            syscall.Handle
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     syscall.Handle
}

var (
	hwnd    syscall.Handle
	nid     notifyIconData
	guiPort int
	quitCh  chan struct{}
)

// Run starts the system tray. This blocks on Windows.
func Run(port int, quit chan struct{}) {
	guiPort = port
	quitCh = quit

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInstance, _, _ := pGetModuleHandle.Call(0)

	className, _ := syscall.UTF16PtrFromString("DevProxyAgent")
	wc := wndClassEx{
		style:     csHRedraw | csVRedraw,
		wndProc:   syscall.NewCallback(wndProc),
		instance:  syscall.Handle(hInstance),
		className: className,
	}
	wc.size = uint32(unsafe.Sizeof(wc))

	icon, _, _ := pLoadIcon.Call(0, uintptr(idiApplication))
	wc.icon = syscall.Handle(icon)
	wc.iconSm = syscall.Handle(icon)

	pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	windowName, _ := syscall.UTF16PtrFromString("DevProxy Agent")
	h, _, _ := pCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0, 0, 0, 0, 0, 0, 0, hInstance, 0,
	)
	hwnd = syscall.Handle(h)

	// Add tray icon
	nid.cbSize = uint32(unsafe.Sizeof(nid))
	nid.hWnd = hwnd
	nid.uID = 1
	nid.uFlags = nifIcon | nifTip | nifMessage
	nid.uCallbackMessage = wmTrayIcon
	nid.hIcon = syscall.Handle(icon)

	tip, _ := syscall.UTF16FromString("DevProxy Agent")
	copy(nid.szTip[:], tip)

	pShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))

	// Message loop
	var m msg
	for {
		ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}

	// Remove tray icon
	pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
}

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmTrayIcon:
		switch lParam {
		case wmRButtonUp:
			showContextMenu(hwnd)
		case wmLButtonDblClk:
			openConfig()
		}
		return 0

	case wmCommand:
		switch wParam & 0xFFFF {
		case idOpenConfig:
			openConfig()
		case idSyncNow:
			agentsync.SyncNow()
		case idPause:
			agentsync.TogglePause()
		case idQuit:
			pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
			pPostQuitMessage.Call(0)
			if quitCh != nil {
				close(quitCh)
			}
		}
		return 0

	case wmDestroy:
		pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
		pPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := pDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func showContextMenu(hwnd syscall.Handle) {
	menu, _, _ := pCreatePopupMenu.Call()

	openStr, _ := syscall.UTF16PtrFromString("Open Config")
	syncStr, _ := syscall.UTF16PtrFromString("Sync Now")

	status := agentsync.GetStatus()
	var pauseStr *uint16
	if status.Paused {
		pauseStr, _ = syscall.UTF16PtrFromString("Resume Sync")
	} else {
		pauseStr, _ = syscall.UTF16PtrFromString("Pause Sync")
	}
	quitStr, _ := syscall.UTF16PtrFromString("Quit")

	pAppendMenu.Call(menu, mfString, idOpenConfig, uintptr(unsafe.Pointer(openStr)))
	pAppendMenu.Call(menu, mfString, idSyncNow, uintptr(unsafe.Pointer(syncStr)))
	pAppendMenu.Call(menu, mfString, idPause, uintptr(unsafe.Pointer(pauseStr)))
	pAppendMenu.Call(menu, mfSeparator, 0, 0)
	pAppendMenu.Call(menu, mfString, idQuit, uintptr(unsafe.Pointer(quitStr)))

	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	pSetForegroundWindow.Call(uintptr(hwnd))
	pTrackPopupMenu.Call(menu, tpmLeftAlign|tpmBottomAlign, uintptr(pt.x), uintptr(pt.y), 0, uintptr(hwnd), 0)
	pDestroyMenu.Call(menu)

	// Fix for menu not closing
	pPostMessage.Call(uintptr(hwnd), 0, 0, 0)
}

func openConfig() {
	url := fmt.Sprintf("http://localhost:%d", guiPort)
	exec.Command("cmd", "/c", "start", url).Start()
}

// Quit sends a quit message to the tray window.
func Quit() {
	if hwnd != 0 {
		pPostMessage.Call(uintptr(hwnd), wmDestroy, 0, 0)
	}
}

// Available returns true on Windows.
func Available() bool {
	return true
}

// OpenConfigURL opens the config GUI in a browser.
func OpenConfigURL(port int) {
	guiPort = port
	openConfig()
}

func init() {
	log.Println("System tray support: available (Windows)")
}
