package auth

import (
	_ "embed"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed epic_auth_webview
var embeddedEpicAuthBinary []byte

const swiftScriptTemplate = `
import Cocoa
import WebKit

class AppDelegate: NSObject, NSApplicationDelegate, WKNavigationDelegate {
    var window: NSWindow!
    var webView: WKWebView!

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        let appName = "RIFT Secure Auth"
        let windowSize = NSSize(width: 500, height: 700)
        let rect = NSRect(x: 0, y: 0, width: windowSize.width, height: windowSize.height)
        
        let styleMask: NSWindow.StyleMask = [.titled, .closable, .miniaturizable]
        window = NSWindow(contentRect: rect, styleMask: styleMask, backing: .buffered, defer: false)
        window.title = appName
        window.center()
        
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
        
        let config = WKWebViewConfiguration()
        config.websiteDataStore = WKWebsiteDataStore.nonPersistent()
        
        webView = WKWebView(frame: rect, configuration: config)
        webView.navigationDelegate = self
        window.contentView = webView
        
        // Use the official Epic Games redirect URL that outputs the authorizationCode as JSON
        let url = URL(string: "https://www.epicgames.com/id/login?redirectUrl=https%3A%2F%2Fwww.epicgames.com%2Fid%2Fapi%2Fredirect%3FclientId%3D34a02cf8f4414e29b15921876da36f9a%26responseType%3Dcode")!
        webView.load(URLRequest(url: url))
        
        window.makeKeyAndOrderFront(nil)
    }
    
    // Hide the webview as soon as it starts navigating to the redirect page
    func webView(_ webView: WKWebView, decidePolicyFor navigationAction: WKNavigationAction, decisionHandler: @escaping (WKNavigationActionPolicy) -> Void) {
        if let urlString = navigationAction.request.url?.absoluteString {
            if urlString.contains("https://www.epicgames.com/id/api/redirect") {
                webView.isHidden = true
            }
        }
        decisionHandler(.allow)
    }
    
    // Once the redirect page loads, extract the JSON
    func webView(_ webView: WKWebView, didFinish navigation: WKNavigation!) {
        if let urlString = webView.url?.absoluteString {
            if urlString.contains("https://www.epicgames.com/id/api/redirect") {
                webView.evaluateJavaScript("document.body.innerText") { result, error in
                    guard let text = result as? String else { return }
                    
                    // Locate JSON bounds just in case of HTML wrappers (e.g. didFinish pre tags)
                    var jsonStr = text.trimmingCharacters(in: .whitespacesAndNewlines)
                    if let start = jsonStr.firstIndex(of: "{"), let end = jsonStr.lastIndex(of: "}") {
                        jsonStr = String(jsonStr[start...end])
                    }
                    
                    if let data = jsonStr.data(using: .utf8),
                       let json = try? JSONSerialization.jsonObject(with: data, options: []) as? [String: Any],
                       let code = json["authorizationCode"] as? String,
                       !code.isEmpty, code != "null" {
                        print("AUTH_CODE:\(code)")
                        fflush(stdout)
                        NSApp.terminate(nil)
                    } else {
                        // Log for debugging if the parsing failed or authorizationCode is null
                        print("PARSING_FAILED: Raw text was: \(text)")
                        fflush(stdout)
                    }
                }
            }
        }
    }
    
    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        print("WINDOW_CLOSED")
        fflush(stdout)
        return true
    }
}

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
`

// SpawnIsolatedEpicAuthPopup creates a temporary Swift script, runs it via 'swift',
// and captures the authentication token (authorizationCode) from its stdout.
func SpawnIsolatedEpicAuthPopup() (string, error) {
	home, _ := os.UserHomeDir()
	binDir := filepath.Join(home, ".rift", "engines")
	os.MkdirAll(binDir, 0755)
	binPath := filepath.Join(binDir, "epic_auth_webview")

	// Ensure the pre-compiled binary exists on disk from embedded bytes (instant, no swiftc needed)
	if len(embeddedEpicAuthBinary) > 0 {
		info, err := os.Stat(binPath)
		if err != nil || info.Size() != int64(len(embeddedEpicAuthBinary)) {
			_ = os.WriteFile(binPath, embeddedEpicAuthBinary, 0755)
		}
	} else if _, err := os.Stat(binPath); os.IsNotExist(err) {
		// Fallback if embedded binary is somehow empty
		tmpScript := filepath.Join(os.TempDir(), "rift_auth.swift")
		if err := os.WriteFile(tmpScript, []byte(swiftScriptTemplate), 0644); err == nil {
			defer os.Remove(tmpScript)
			compileCmd := exec.Command("swiftc", tmpScript, "-o", binPath)
			_ = compileCmd.Run()
		}
	}

	cmd := exec.Command(binPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start popup: %v", err)
	}

	code := ""
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "AUTH_CODE:") {
			code = strings.TrimPrefix(line, "AUTH_CODE:")
		} else if strings.HasPrefix(line, "PARSING_FAILED:") {
			fmt.Printf("[Swift Webview Debug] %s\n", line)
		} else if line == "WINDOW_CLOSED" && code == "" {
			return "", fmt.Errorf("user closed the login window")
		} else {
			fmt.Printf("[Swift Webview Stdout] %s\n", line)
		}
	}

	cmd.Wait()

	if code != "" {
		return code, nil
	}

	return "", fmt.Errorf("login process ended without capturing an authorization code")
}
