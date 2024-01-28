#!/bin/bash
#set -x
#set -u

# The EXACT network name that you want to autoconnect wireguard with
netname_to_connect=""

# Default mode
wgmode="oracle"

# Wireguard peer info
privkey1=
pubkey1=
endpoint1port=
endpoint1=

addre1=
addre1nm=

privkey2=
pubkey2=
endpoint2port=
endpoint2=

addre2=
addre2nm=

# Outputs logs to variable: line
journalctl -f | while read line;

do

# For manual imput into the logs
# ie logger wgmode_home to set the wgmode variable to "home"
if [[ $line =~ wgmode_home ]]; then
    wgmode="home"
fi

if [[ $line =~ wgmode_oracle ]]; then
    wgmode="oracle"
fi

if [[ $line =~ wgmode_stop ]]; then
    exit
fi

# Detects a line in the logs where the network name is listed and sets a variable to hold it
if [[ $line =~ "device (wlan0): Activation: starting connection" ]]; then
    netname=$(echo "$line" | cut -d " " -f 14)
    echo $netname
fi


if [[ $netname =~ $netname_to_connect ]] && [[ $line =~ CONNECTED_GLOBAL ]] && [[ $line =~ NetworkManager ]] && [[ $wgmode == "oracle" ]]; then

    # Creates a tempfile to serve as the wireguard configuration file
    # Fill in all the variables with your Peer information
cat >/tmp/wg_oracle.conf <<EOF
    [Interface]
    PrivateKey = ${privkey1}

    [Peer]
    PublicKey = ${pubkey1}
    Endpoint = ${endpoint1port}
    AllowedIPs = 0.0.0.0/0
    PersistentKeepalive = 10
EOF

    # Manual wireguard setup
    ip link add dev wg_oracle type wireguard
    ip address add dev wg_oracle $addre1nm
    ip link set up dev wg_oracle
    wg setconf wg_oracle /tmp/wg_oracle.conf

    # Cuts the current default gateway to use
    devgw=$(ip route show default | cut -d " " -f 3)
    ip route add $endpoint1 via $devgw
    ip route del default
    ip route add default via $addre1
    sed  -i 's/nameserver .*/nameserver 8.8.8.8/g' /etc/resolv.conf

    # If you want to use a local nameserver
    #sed  -i 's/nameserver .*/nameserver 127.0.0.1/g' /etc/resolv.conf
    #systemctl start named
fi

# Second "home" wireguard configuration; uncomment to use
: <<'END_COMMENT'
if [[ $netname =~ $netname_to_connect ]] && [[ $line =~ CONNECTED_GLOBAL ]] && [[ $line =~ NetworkManager ]] && [[ $wgmode == "home" ]]; then

    # Creates a tempfile to serve as the wireguard configuration file
    # Fill in all the variables with your Peer information
cat >/tmp/wg_home.conf <<EOF
    [Interface]
    PrivateKey = ${privkey2}

    [Peer]
    PublicKey = ${pubkey2}
    Endpoint = ${endpoint2port}
    AllowedIPs = 0.0.0.0/0
    PersistentKeepalive = 10
EOF

    # Manual wireguard setup
    ip link add dev wg_home type wireguard
    ip address add dev wg_home $addre2nm
    ip link set up dev wg_home
    wg setconf wg_home /tmp/wg_home.conf

    # Cuts the current default gateway to use
    devgw=$(ip route show default | cut -d " " -f 3)
    ip route add $endpoint2 via $devgw
    ip route del default
    ip route add default via $addre2
    sed  -i 's/nameserver .*/nameserver 8.8.8.8/g' /etc/resolv.conf

    # If you want to use a local nameserver
    #sed  -i 's/nameserver .*/nameserver 127.0.0.1/g' /etc/resolv.conf
    #systemctl start named
fi
END_COMMENT

# Disables wireguard when the network(wifi) deactivates/changes.
# Nessesary because the manual activation here relies on the default gateway being set
if [[ $line =~ deactivating ]] && [[ $line =~ disconnected ]] && [[ $line =~ NetworkManager ]] || [[ $line =~ "failed for connection" ]] && [[ $line =~ NetworkManager ]]; then
    ifconfig wg_home down
    ifconfig wg_oracle down
fi


done
