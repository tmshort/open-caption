# Troubleshooting

## The Troubleshoot button

The **Troubleshoot** button at the top of Setup, next to Settings, is the place to start when people cannot connect or something looks wrong. It lists what the program knows: whether this computer has a network address to give out, which port the web server really got, whether HTTPS and its certificate are in order, whether the `.local` name is announced, whether the settings file can be saved, whether every language matches the main show file, and how many devices are connected. Each check is marked **OK**, **Note**, **Check** or **Problem**, and the ones that can be put right have a button that goes to the right setting. Below the checks are the latest warnings and errors from the log, a link to the full log, and the support bundle.

When the program logs new warnings or errors, a number appears on the button until you open it.

## Nobody can open the address

- Check that the phone is on the venue's guest WiFi, not mobile data.
- Check that the address matches the one in the **Share** step. If it has a port number, include it.
- Ask the venue whether guest WiFi devices are allowed to reach this computer.
- Your computer's firewall may be blocking the program. On Windows, allow it on **private networks** when asked. On macOS, allow incoming connections for it.
- Open the address on this computer using `localhost` to check the program itself.

## The number of connected devices is lower than expected

Devices that fell asleep disconnect and reconnect by themselves when woken. Ask people to turn off auto-lock for the show if their screen dims.

## Pictures appear slowly

Large pictures take time on a busy WiFi. Check the Images step for files marked "large" and replace them with smaller ones.

For the full list of checks on the network side, see the [venue network guide](topic:admin-venue-network-guide).

## A language is missing for the audience

Open the Show file step and look at that language's tab: a ⚠ means its show file no longer has the same cues as the main one. Fix the error shown; the language returns as soon as you save. See [Languages](topic:admin-languages).

## Show file problems

See "What the warnings mean" in [Writing a show file](topic:admin-writing-a-show-file).

## Passwords

To change either password, use **Settings** at the top of the Setup page. The new password takes effect immediately and is saved in `config.yaml`. Anyone signed in with a password that changed is signed out; you stay signed in. The same window can show the current operator password so you can give it to the operator.

If you cannot sign in at all, both passwords are in `config.yaml`, in the program's data folder. Open the file in any text editor. Its location is printed when the program starts.

## Looking at the log

**Troubleshoot → Open the full log…** shows what the program has logged lately. Choose *Problems only* to see warnings and errors, such as refused sign-ins or a name already used on the network. The same window sets how much detail is logged (**Log detail**); *Debug* adds every request, which helps when something odd is happening. Passwords are never written to the log.

## Asking for help

Download the **support bundle** with the **Troubleshoot** button at the top of Setup. It contains the settings (without passwords), recent log lines and basic information about this computer, and is easy to send to whoever maintains the program.
