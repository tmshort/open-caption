# Troubleshooting

## Nobody can open the address

- Check that the phone is on the venue's guest WiFi, not mobile data.
- Check that the address matches the one in the **Go Live** step. If it has a port number, include it.
- Ask the venue whether guest WiFi devices are allowed to reach this computer.
- Your computer's firewall may be blocking the program. On Windows, allow it on **private networks** when asked. On macOS, allow incoming connections for it.
- Open the address on this computer using `localhost` to check the program itself.

## The number of connected devices is lower than expected

Devices that fell asleep disconnect and reconnect by themselves when woken. Ask people to turn off auto-lock for the show if their screen dims.

## Pictures appear slowly

Large pictures take time on a busy WiFi. Check the Images step for files marked "large" and replace them with smaller ones.

For the full list of checks on the network side, see the [venue network guide](topic:admin-venue-network-guide).

## A language is missing for the audience

Open the Script step and look at that language's tab: a ⚠ means its script no longer has the same cues as the main one. Fix the error shown; the language returns as soon as you save. See [Languages](topic:admin-languages).

## Script problems

See "What the warnings mean" in [Writing a script](topic:admin-writing-a-script).

## Passwords

To change either password, use **Settings** at the top of the Setup page. The new password takes effect immediately and is saved in `config.yaml`. Anyone signed in with a password that changed is signed out; you stay signed in. The same window can show the current operator password so you can give it to the operator.

If you cannot sign in at all, both passwords are in `config.yaml`, in the program's data folder. Open the file in any text editor. Its location is printed when the program starts.

## Asking for help

Download the **support bundle** from the bottom of the Go Live step. It contains the settings (without passwords), recent log lines and basic information about this computer, and is easy to send to whoever maintains the program.
