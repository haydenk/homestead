# Icon Reference

The `icon` field on both sections and items accepts an emoji character or an image URL.

---

## Inserting emoji in a terminal editor

### macOS (vim, nano, etc.)

Press `Ctrl+Cmd+Space` to open the system emoji picker. Click an emoji to insert it at the cursor position. This works in any terminal application.

### Linux

The method depends on your desktop environment and terminal:

- **GTK terminals (GNOME Terminal, Tilix, etc.):** Press `Ctrl+Shift+U`, type the Unicode code point (e.g. `1f3e0` for 🏠), then press `Enter` or `Space`.
- **KDE / Plasma:** Use the emoji picker at `Win+.` if your version supports it, or install `ibus` and use its emoji input (`Ctrl+Shift+E`).
- **Copy/paste:** The most reliable option on any distro — open the emoji table below in a browser and copy the character directly.

### Windows (WSL, PowerShell, cmd)

Press `Win+.` (Windows key + period) or `Win+;` to open the built-in emoji picker. Search by name and click to insert. Works in Windows Terminal and most modern terminal emulators running on Windows.

---

## Using icons in config

```toml
# Emoji — paste the character directly
icon = "🏠"

# Image URL (for service-specific logos)
icon = "https://example.local/logo.png"
```

---

## General / Dashboard

| Emoji | Name |
|---|---|
| 🏠 | House |
| 🏡 | House with garden |
| 🌐 | Globe |
| 🔍 | Magnifying glass |
| ⚙️ | Gear |
| 📋 | Clipboard |
| 📌 | Pushpin |
| ⭐ | Star |
| 🔖 | Bookmark |
| 🗂️ | Card index dividers |

---

## Networking

| Emoji | Name |
|---|---|
| 🌐 | Globe |
| 📡 | Satellite antenna |
| 🔀 | Shuffle / Router |
| 🔗 | Link |
| 🕳️ | Hole |
| 🛡️ | Shield |
| 🔒 | Lock |
| 🔓 | Unlocked lock |
| 🌍 | Earth globe |
| 📶 | Signal bars |
| 🔧 | Wrench |
| 🚦 | Traffic light |

---

## Media

| Emoji | Name |
|---|---|
| 🎬 | Clapper board |
| 📺 | Television |
| 🎞️ | Film frames |
| 🎵 | Musical note |
| 🎶 | Musical notes |
| 🎧 | Headphone |
| 🎙️ | Studio microphone |
| 🎥 | Movie camera |
| 📷 | Camera |
| 📸 | Camera with flash |
| 🖼️ | Framed picture |
| 📻 | Radio |
| 🎮 | Video game |
| 🕹️ | Joystick |
| 📖 | Open book |
| 📚 | Books |

---

## Infrastructure

| Emoji | Name |
|---|---|
| 🖥️ | Desktop computer |
| 💻 | Laptop |
| 🖨️ | Printer |
| 💾 | Floppy disk |
| 💿 | Optical disk |
| 🗄️ | File cabinet |
| 🐳 | Whale (Docker) |
| ☁️ | Cloud |
| 🔵 | Blue circle |
| ⚡ | Lightning bolt |
| 🔋 | Battery |
| 🌡️ | Thermometer |
| 📦 | Package |
| 🏗️ | Building construction |

---

## Monitoring

| Emoji | Name |
|---|---|
| 📊 | Bar chart |
| 📈 | Chart increasing |
| 📉 | Chart decreasing |
| 📡 | Satellite antenna |
| 💚 | Green heart |
| ❤️ | Red heart |
| 🔔 | Bell |
| 🔕 | Bell with slash |
| ⏱️ | Stopwatch |
| 🕐 | Clock |
| 🔴 | Red circle |
| 🟢 | Green circle |
| 🟡 | Yellow circle |

---

## Development

| Emoji | Name |
|---|---|
| 💻 | Laptop |
| 🐱 | Cat face |
| 🐙 | Octopus |
| 🔧 | Wrench |
| 🔨 | Hammer |
| 🛠️ | Hammer and wrench |
| 🚀 | Rocket |
| 🐛 | Bug |
| ✅ | Check mark |
| ❌ | Cross mark |
| 🔁 | Repeat |
| 🚁 | Helicopter |
| 🆚 | VS button |
| 📝 | Memo |
| 🗃️ | Card file box |
| 🐘 | Elephant |
| 🐬 | Dolphin |
| 🍃 | Leaf |

---

## Security & Access

| Emoji | Name |
|---|---|
| 🔑 | Key |
| 🗝️ | Old key |
| 🔐 | Locked with key |
| 🛡️ | Shield |
| 👤 | User silhouette |
| 👥 | Users |
| 🔏 | Locked with pen |
| 🕵️ | Detective |

---

## Automation & Home

| Emoji | Name |
|---|---|
| 🤖 | Robot |
| 💡 | Light bulb |
| 🏠 | House |
| 🌡️ | Thermometer |
| 🔌 | Electric plug |
| 📱 | Mobile phone |
| 🖨️ | Printer |
| ☀️ | Sun |
| 🌙 | Crescent moon |
| 🌤️ | Sun behind cloud |

---

## Notes

Any single Unicode emoji character is valid in the `icon` field. The tables above are a curated starting point; any emoji from [unicode.org](https://unicode.org/emoji/charts/full-emoji-list.html) can be used.
