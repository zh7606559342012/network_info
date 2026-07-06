#!/bin/bash

CUR_PATH=$(dirname "$(readlink -f "$0")")
GO_ROOT="/opt/golang"
AGENT_ROOT="/opt/monitor_agent"
AGENT_BIN="${AGENT_ROOT}/bin"
AGENT_LIB="${AGENT_ROOT}/lib"
LOG_PATH="/var/log/monitor_agent"
BIND_IP="0.0.0.0"

die() {
    echo $1
    exit 1
}

usage() {
  script_name=`basename "$0"`
  echo "./${script_name} -i 127.0.0.1 #bind local ip addr"
  exit 0
}

disable_selinux() {
    if [ x${OS_TYPE} == x"centos" ];then
        CHECK=$(grep SELINUX= /etc/selinux/config | grep -v "#")
        case $CHECK in
        "SELINUX=enforcing")
            sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config
            setenforce 0
        ;;
        "SELINUX=permissive")
            sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config
            setenforce 0
        ;;
        esac
    fi
}


env_check() {
    /opt/monitor_agent/bin/stop_all.sh >/dev/null 2>&1
    rm -rf ${AGENT_ROOT} >/dev/null 2>&1

    mkdir -p ${AGENT_ROOT}
    mkdir -p ${AGENT_BIN}
    mkdir -p ${LOG_PATH}
    mkdir -p ${AGENT_ROOT}/scripts
}

install_scripts() {
    cp -f ${CUR_PATH}/scripts/*.sh ${AGENT_BIN}
    chmod +x ${AGENT_BIN}/*.sh
}

redis_check() {
  if [ -f /usr/local/redis/src/redis-server ]; then
      echo "redis has installed"
  else
      echo "redis is installing"
      mkdir -p ${AGENT_LIB}
      install_depend_redis
  fi
}

install_depend_redis() {
    id -u redis &>/dev/null || useradd -r redis -s /sbin/nologin
    tar -xzf ${CUR_PATH}/3rd/redis.tar.gz -C ${AGENT_LIB} --overwrite > /dev/null

    [ -d ${AGENT_LIB}/redisdata ] || mkdir -p ${AGENT_LIB}/redisdata
    \cp -f ${CUR_PATH}/bin/conf/redis.conf ${AGENT_LIB}/redisdata
    chmod 600 "${AGENT_LIB}/redisdata/redis.conf"
    systemctl stop redis

cat << EOF > /usr/lib/systemd/system/redis.service
[Unit]
Description=Redis
After=syslog.target

[Service]
User=redis
Group=redis
ExecStart=${AGENT_LIB}/redis-server ${AGENT_LIB}/redisdata/redis.conf
RestartSec=5s
Restart=on-success

[Install]
WantedBy=multi-user.target
EOF
    chown redis:redis -R ${AGENT_LIB}
    systemctl daemon-reload
    systemctl restart redis
    systemctl enable redis

}

_gen_agent_service() {
cat << EOF > /usr/lib/systemd/system/monitor_agent.service
[Unit]
Description=Monitor Agent service
After=network.target

[Service]
User=root
Group=root
Type=simple
ExecStart=${AGENT_BIN}/agent

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable monitor_agent.service
    #定时器启动
}

_gen_log() {

cat << EOF > /etc/logrotate.d/monitor_agent
/var/log/monitor_agent/agent.log {
    daily
    rotate 10
    size 20M
    missingok
    notifempty
    sharedscripts
    delaycompress
    copytruncate
}
EOF

}

make_autostart() {
    _gen_agent_service
    _gen_log
}

install_agent_withtar() {
    cd ${CUR_PATH}
    cp -f ${CUR_PATH}/bin/agent ${AGENT_BIN}
    mkdir -p ${AGENT_BIN}/conf
    cp -f ${CUR_PATH}/bin/conf/* ${AGENT_BIN}/conf/
    cp -rf ${CUR_PATH}/scripts/* ${AGENT_ROOT}/scripts
    chmod +x ${AGENT_ROOT}/scripts/*
    chmod +x ${AGENT_BIN}/agent
}

install_watchdog()
{
cat << EOF > /usr/lib/systemd/system/monitor_agentwatching.service
[Unit]
Description=Watching Monitor Agent process

[Service]
Type=simple
ExecStart=${AGENT_BIN}/agent_watching.sh
EOF

cat << EOF > /usr/lib/systemd/system/monitor_agentwatching.timer
[Unit]
Description=Watching process every 5 seconds

[Timer]
OnBootSec=120
OnUnitActiveSec=5
AccuracySec=1ms
Unit=monitor_agentwatching.service

[Install]
WantedBy=timers.target
EOF

systemctl daemon-reload
systemctl enable monitor_agentwatching.timer

# disable nagios nologin
sed -i '/nagios/s/bin\/bash/sbin\/nologin/' /etc/passwd
}


config_firewall() {
    firewall-cmd --zone=public --add-port=20000/tcp --permanent > /dev/null 2>&1
    firewall-cmd --reload > /dev/null 2>&1
}


arg_index=1
while [[ ${arg_index} -le $# ]]
do
  key=$(eval echo \$${arg_index})
  case $key in
    -i|--ip)
      ((arg_index++))
      BIND_IP="$(eval echo \$${arg_index})"
      ((arg_index++))
      ;;
    *)
      echo "invalid arguments"
      ((arg_index++))
      ;;
  esac
done

if [ $(cat /etc/os-release | grep -i centos >/dev/null 2>&1;echo $?) -eq 0 ];then
    OS_TYPE="centos"
elif [ $(cat /etc/os-release | grep -i ubuntu > /dev/null 2>&1;echo $?) -eq 0 ];then
    OS_TYPE="ubuntu"
else
    echo "unsupport os,currently only support ubuntu&centos."
    exit 1
fi

if [[ ${BIND_IP} =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "oms bind to ${BIND_IP}"
else
    die "invalid ip addrress: ${BIND_IP}"
fi

disable_selinux
env_check
install_agent_withtar
install_scripts
redis_check
make_autostart
install_watchdog
#config_firewall

#start
systemctl restart monitor_agentwatching.timer