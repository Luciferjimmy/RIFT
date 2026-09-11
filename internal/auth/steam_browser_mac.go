package auth

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const steamBrowserSwiftTemplate = `
import Cocoa
import WebKit

// Allowed domains for the locked-down Steam browser
let allowedDomains = ["steampowered.com", "steamcommunity.com", "steamstore-a.akamaihd.net", "steamcdn-a.akamaihd.net", "steamstatic.com"]

class SteamBrowserDelegate: NSObject, NSApplicationDelegate, WKNavigationDelegate {
    var window: NSWindow!
    var webView: WKWebView!

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        let appName = "RIFT — Steam Browser"
        let windowSize = NSSize(width: 1200, height: 800)
        let rect = NSRect(x: 0, y: 0, width: windowSize.width, height: windowSize.height)

        let styleMask: NSWindow.StyleMask = [.titled, .closable, .miniaturizable, .resizable]
        window = NSWindow(contentRect: rect, styleMask: styleMask, backing: .buffered, defer: false)
        window.title = appName
        window.center()

        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)

        // Use persistent data store so the user stays logged in
        let config = WKWebViewConfiguration()
        if #available(macOS 13.3, *) {
            // Use default persistent store
        }
        config.websiteDataStore = WKWebsiteDataStore.default()

        webView = WKWebView(frame: rect, configuration: config)
        webView.navigationDelegate = self
        window.contentView = webView

        // Load the URL passed as command-line argument, or default to the Steam store
        let args = CommandLine.arguments
        let urlString = args.count > 1 ? args[1] : "https://store.steampowered.com"
        if let url = URL(string: urlString) {
            webView.load(URLRequest(url: url))
        }

        window.makeKeyAndOrderFront(nil)
    }

    // Domain whitelist: only allow navigation to allowed domains
    func webView(_ webView: WKWebView, decidePolicyFor navigationAction: WKNavigationAction, decisionHandler: @escaping (WKNavigationActionPolicy) -> Void) {
        if let url = navigationAction.request.url, let host = url.host?.lowercased() {
            for allowed in allowedDomains {
                if host == allowed || host.hasSuffix("." + allowed) {
                    decisionHandler(.allow)
                    return
                }
            }
            // Also allow steam:// links (for launching games etc.)
            if url.scheme == "steam" {
                if let appURL = URL(string: "steam://" + url.path) {
                    NSWorkspace.shared.open(appURL)
                }
                decisionHandler(.cancel)
                return
            }
            // Block all other domains
            decisionHandler(.cancel)
        } else {
            decisionHandler(.allow)
        }
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        print("STEAM_BROWSER_CLOSED")
        fflush(stdout)
        return true
    }
}

let app = NSApplication.shared
let delegate = SteamBrowserDelegate()
app.delegate = delegate
app.run()
`

// SpawnSteamBrowser opens a locked-down WKWebView window pointed at the given Steam URL.
// Navigation is restricted to steampowered.com, steamcommunity.com, and their CDNs.
// The browser uses a persistent data store so the user stays logged in between sessions.
func SpawnSteamBrowser(url string) error {
	home, _ := os.UserHomeDir()
	binDir := filepath.Join(home, ".rift", "engines")
	os.MkdirAll(binDir, 0755)
	binPath := filepath.Join(binDir, "steam_browser")

	// Compile the Swift binary if it doesn't exist
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		tmpScript := filepath.Join(os.TempDir(), "rift_steam_browser.swift")
		err := os.WriteFile(tmpScript, []byte(steamBrowserSwiftTemplate), 0644)
		if err != nil {
			return fmt.Errorf("failed to create swift script: %w", err)
		}
		defer os.Remove(tmpScript)

		fmt.Println("[Swift] Compiling Steam Browser (this may take a minute on first run)...")
		compileCmd := exec.Command("swiftc", tmpScript, "-o", binPath)
		if out, err := compileCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to compile steam browser: %w\nOutput: %s", err, string(out))
		}
		fmt.Println("[Swift] Steam Browser compiled successfully.")
	}

	cmd := exec.Command(binPath, url)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start steam browser: %w", err)
	}

	// Read stdout in background to log any messages
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "STEAM_BROWSER_CLOSED" {
				fmt.Println("[Steam Browser] Window closed by user.")
			} else {
				fmt.Printf("[Steam Browser] %s\n", line)
			}
		}
		cmd.Wait()
	}()

	return nil
}
