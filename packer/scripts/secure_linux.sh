#!/usr/bin/env bash

## update all packages
#yum upgrade -y

# Restricting access to kernel logs
echo "kernel.dmesg_restrict = 1" > /etc/sysctl.d/50-dmesg-restrict.conf

# Restricting access to kernel pointers
echo "kernel.kptr_restrict = 1" > /etc/sysctl.d/50-kptr-restrict.conf

# Randomise memory space
echo "kernel.randomize_va_space=2" > /etc/sysctl.d/50-rand-va-space.conf

# Ensure syslog service is enabled and running
systemctl enable --now rsyslog

# Set auto logout inactive users
echo "readonly TMOUT=900" >> /etc/profile.d/idle-users.sh
echo "readonly HISTFILE" >> /etc/profile.d/idle-users.sh
chmod +x /etc/profile.d/idle-users.sh

# Enable hard/soft link protection
echo "fs.protected_hardlinks = 1" > /etc/sysctl.d/50-fs-hardening.conf
echo "fs.protected_symlinks = 1" >> /etc/sysctl.d/50-fs-hardening.conf

# Disable uncommon filesystems.
echo "install cramfs /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install freevxfs /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install jffs2 /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install hfs /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install hfsplus /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install squashfs /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install udf /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install fat /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install vfat /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install nfs /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install nfsv3 /bin/false" > /etc/modprobe.d/uncommon-fs.conf
echo "install gfs2 /bin/false" > /etc/modprobe.d/uncommon-fs.conf

# Enable TCP SYN Cookie protection.
echo "net.ipv4.tcp_syncookies = 1" > /etc/sysctl.d/50-net-stack.conf

# install ufw
yum install ufw -y

# back to clean state
yes | ufw reset

# block all traffic
ufw default deny

# allow ssh
ufw allow ssh

# turn on logging and enable firewall
ufw logging on
yes | ufw enable