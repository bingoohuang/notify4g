#!/bin/bash

if [[ -d /etc/opt/notify4g ]]; then
    # Legacy configuration found
    if [[ ! -d /etc/notify4g ]]; then
        # New configuration does not exist, move legacy configuration to new location
        echo -e "Please note, Rigaga's configuration is now located at '/etc/notify4g' (previously '/etc/opt/notify4g')."
        mv -vn /etc/opt/notify4g /etc/notify4g

        if [[ -f /etc/notify4g/notify4g.conf ]]; then
            backup_name="notify4g.conf.$(date +%s).backup"
            echo "A backup of your current configuration can be found at: /etc/notify4g/${backup_name}"
            cp -a "/etc/notify4g/notify4g.conf" "/etc/notify4g/${backup_name}"
        fi
    fi
fi