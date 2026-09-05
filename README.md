# RIFT — Windows Gaming on Apple Silicon

<p align="center">
  <img src="https://raw.githubusercontent.com/Luciferjimmy/RIFT/main/assets/rift_logo.png" alt="RIFT Logo" width="120" />
</p>

<p align="center">
  <strong>Windows gaming on Apple Silicon without the Reddit trauma.</strong><br>
  Consolidated gaming console powered by D3DMetal, DXVK, and automated Wine prefix virtualization for macOS.
</p>

<p align="center">
  <a href="https://github.com/Luciferjimmy/RIFT/releases/latest"><img src="https://img.shields.io/github/v/release/Luciferjimmy/RIFT?color=00f2fe&style=flat-square" alt="Latest Release"></a>
  <img src="https://img.shields.io/badge/platform-macOS%20Apple%20Silicon-blue?style=flat-square" alt="Apple Silicon">
  <img src="https://img.shields.io/badge/arch-ARM64-brightgreen?style=flat-square" alt="ARM64">
  <img src="https://img.shields.io/badge/license-Proprietary-orange?style=flat-square" alt="License">
</p>

---

## ⚡ Quick Install (1-Line Terminal Command)

Open **Terminal** on your Mac and run:

```bash
curl -fsSL https://raw.githubusercontent.com/Luciferjimmy/RIFT/main/install.sh | bash
```

> **Why use this?** Automatically detects Apple Silicon, installs Rosetta 2 if missing, downloads the latest RIFT DMG, installs it to `/Applications`, and removes macOS Gatekeeper quarantine flags so it opens instantly.

---

## 📦 Manual Download

1. Go to the [**Latest Release**](https://github.com/Luciferjimmy/RIFT/releases/latest).
2. Download **`RIFT-macOS-arm64.dmg`**.
3. Open the `.dmg` and drag **RIFT** into your **Applications** folder.
4. If macOS displays an *unidentified developer* warning:
   - Right-click **RIFT.app** in Applications and choose **Open**.
   - Or run `xattr -cr /Applications/RIFT.app` in Terminal.

---

## ✨ Features

- **🎮 Zero-Config Translation**: No manual WINEPREFIX management, no winetricks scripts, no terminal wrestling.
- **🚀 Dual Graphics Engines**: 
  - **Apple Game Porting Toolkit (D3DMetal)** for DirectX 11/12 AAA games.
  - **DXVK 2.5.3 (MoltenVK)** for DirectX 9/10/11 titles.
  - Automatically picks the fastest engine for every game using heuristic binary scanning.
- **🛡️ Isolated Game Capsules**: Every game runs in its own sandboxed Wine bottle. Installing or modifying one game never corrupts another.
- **🔌 Unified Store Integration**: Connect your Steam and Epic Games accounts directly to discover, download, and launch games seamlessly.
- **📊 Real-Time Hardware Telemetry**: Live FPS, Apple Silicon GPU core utilization, RAM/Swap monitors, and thermal status.

---

## 🖥️ System Requirements

| Specification | Requirement |
|---|---|
| **Architecture** | Apple Silicon (M1, M2, M3, M4 — all variants) |
| **Operating System** | macOS Sonoma (14.0+) or macOS Sequoia (15.0+) |
| **Memory** | 8 GB unified memory minimum (16 GB+ recommended) |
| **Disk Space** | 2 GB for engine runtimes + game storage |

---

## 🛠️ Issues & Feedback

Found a bug or have a game compatibility report?
Please [open an issue](https://github.com/Luciferjimmy/RIFT/issues) on GitHub!
