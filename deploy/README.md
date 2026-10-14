# Running open-caption as a service

Most theaters can simply start the program by hand before the show (see
[docs/usage-guide.md](../docs/usage-guide.md)). If the captioning computer should start the server by itself
after a reboot, use one of these:

| File | Platform |
|------|----------|
| `open-caption.service` | Linux (systemd) |
| `org.open-caption.server.plist` | macOS (launchd, starts at login) |
| `windows-install.ps1` | Windows (Task Scheduler, starts at boot) |

Each file starts with its own installation steps. All three pass `--no-browser` (a service has no screen) and an explicit
`--data` folder, so the passwords and shows live in a known place that you can back up.
