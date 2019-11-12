#!/bin/bash

function disable_systemd {
    systemctl disable notify4g
    rm -f $1
}

function disable_update_rcd {
    update-rc.d -f notify4g remove
    rm -f /etc/init.d/notify4g
}

function disable_chkconfig {
    chkconfig --del notify4g
    rm -f /etc/init.d/notify4g
}

if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    # RHEL-variant logic
    if [[ "$1" = "0" ]]; then
        # InfluxDB is no longer installed, remove from init system
        rm -f /etc/default/notify4g

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /usr/lib/systemd/system/notify4g.service
        else
            # Assuming sysv
            disable_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    if [ "$1" == "remove" -o "$1" == "purge" ]; then
        # Remove/purge
        rm -f /etc/default/notify4g

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /lib/systemd/system/notify4g.service
        else
            # Assuming sysv
            # Run update-rc.d or fallback to chkconfig if not available
            if which update-rc.d &>/dev/null; then
                disable_update_rcd
            else
                disable_chkconfig
            fi
        fi
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ $ID = "amzn" ]]; then
        # Amazon Linux logic
        if [[ "$1" = "0" ]]; then
            # InfluxDB is no longer installed, remove from init system
            rm -f /etc/default/notify4g
            disable_chkconfig
        fi
    fi
fi
