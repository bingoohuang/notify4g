#!/bin/bash

BIN_DIR=/usr/bin
LOG_DIR=/var/log/notify4g
SCRIPT_DIR=/usr/lib/notify4g/scripts
LOGROTATE_DIR=/etc/logrotate.d
USER=rigaga
GROUP=rigaga

function install_init {
    cp -f $SCRIPT_DIR/init.sh /etc/init.d/notify4g
    chmod +x /etc/init.d/notify4g
}

function install_systemd {
    cp -f $SCRIPT_DIR/notify4g.service $1
    systemctl enable notify4g || true
    systemctl daemon-reload || true
}

function install_update_rcd {
    update-rc.d notify4g defaults
}

function install_chkconfig {
    chkconfig --add notify4g
}

if ! grep "^rigaga:" /etc/group &>/dev/null; then
    groupadd -r rigaga
fi

if ! id rigaga &>/dev/null; then
    useradd -r -M rigaga -s /bin/false -d /etc/rigaga -g rigaga
fi

test -d $LOG_DIR || mkdir -p $LOG_DIR
chown -R -L $USER:$GROUP $LOG_DIR
chmod 755 $LOG_DIR

##
# Remove legacy symlink, if it exists
if [[ -L /etc/init.d/notify4g ]]; then
    rm -f /etc/init.d/notify4g
fi
# Remove legacy symlink, if it exists
if [[ -L /etc/systemd/system/notify4g.service ]]; then
    rm -f /etc/systemd/system/notify4g.service
fi

# Add defaults file, if it doesn't exist
if [[ ! -f /etc/default/notify4g ]]; then
    touch /etc/default/notify4g
fi

if [[ ! -d /etc/notify4g ]]; then
  mkdir -p /etc/notify4g
fi

if [[ -d /etc/notify4g ]]; then
    chown -R -L $USER:$GROUP /etc/notify4g
fi

if [[ -f $BIN_DIR/notify4g ]]; then
    chmod +x $BIN_DIR/notify4g
fi

# Add .d configuration directory
if [[ ! -d /etc/notify4g/notify4g.d ]]; then
    mkdir -p /etc/notify4g/notify4g.d
    chown -R -L $USER:$GROUP /etc/notify4g/notify4g.d
fi

# Add snapshots configuration directory
if [[ ! -d /etc/notify4g/snapshots ]]; then
    mkdir -p /etc/notify4g/snapshots
    chown -R -L $USER:$GROUP /etc/notify4g/snapshots
fi

# Distribution-specific logic
if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    # RHEL-variant logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /usr/lib/systemd/system/notify4g.service
    else
        # Assuming SysVinit
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /lib/systemd/system/notify4g.service
        deb-systemd-invoke restart notify4g.service || echo "WARNING: systemd not running."
    else
        # Assuming SysVinit
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
        invoke-rc.d notify4g restart
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ $ID = "amzn" ]]; then
        # Amazon Linux logic
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
fi