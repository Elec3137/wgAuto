For linux, tested on Arch/endeavorOS

uses `wg-tools`, `awk`, relies on `systemd` and `networkmanager`

intended to be used primarily as a systemd service; replace the variables in the .service file with your wireguard server information, and copy it into `/etc/systemd/system/`

Usage:  `e_wgAuto start [netname] [privkey] [pubkey] [endpoint] [port] [addre] [nm]`
