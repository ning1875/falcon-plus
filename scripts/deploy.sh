#!/usr/bin/env bash

#1、此脚本实现部署和检测更新功能
#2、更新配置和更新二进制文件

if grep -q ebian /etc/issue;then
	echo "..."
else
	exit 100
fi


IPADDR=`ifconfig eth0|awk -F: '/inet addr/{split($2,a," ");print a[1];exit}'|awk -F "." '{print $2}'`;
if [ $IPADDR > 99 ];then
	BASE_URL="http://10.100.6.187:8888/agent/"
else
	BASE_URL="http://10.6.193.159:8888/agent/"
fi
BASE_DIR="/opt/open-falcon/agent/"

#First 判断是更新还是创建
# 查看是否有user 2000 为falcon,有则部署过了，只做更新。

get_md5(){
	wget -q -O ${BASE_DIR}md5.txt ${BASE_URL}md5.txt
	cat  ${BASE_DIR}md5.txt
}


reconfig(){
		wget  -O ${BASE_DIR}cfg.json ${BASE_URL}cfg.json
		chown  falcon:falcon ${BASE_DIR}cfg.json
		#Restart the service (jessie or wheezy)
		if  grep -q  "8" /etc/issue;then
			sudo -u falcon -g falcon XDG_RUNTIME_DIR=/run/user/2000 systemctl --user restart falcon-agent.service
		else
            ps auxf |grep falcon-agent|grep -v grep|awk '{print $2}'|xargs kill
			cd $BASE_DIR;sudo -u falcon ./control restart
		fi
}

update_agent(){
	if  grep -q  "8" /etc/issue;then
		systemctl stop user@2000.service
		wget -O ${BASE_DIR}falcon-agent ${BASE_URL}falcon-agent
		systemctl start user@2000.service
	else
 	    svc -d /etc/service/falcon_agent_run
		wget -O ${BASE_DIR}falcon-agent ${BASE_URL}falcon-agent
 	    svc -u /etc/service/falcon_agent_run

	fi
}

install_agent(){
	#add user falcon
	groupadd -g 2000 falcon
	useradd -g 2000 -u 2000 -s /bin/bash -m falcon
	#install open-falcon agent
	mkdir -p ${BASE_DIR}var
	wget -O /opt/open-falcon/falcon-agent.tar.gz   http://10.100.6.187:8888/falcon-agent.tar.gz
	cd /opt/open-falcon/; tar -zxvf falcon-agent.tar.gz;
	cd /opt/;chown -R falcon:falcon open-falcon/
	#jessie or wheezy
	if  grep -q  "8" /etc/issue;then
		#add systemd longin tools
		apt-get update;apt-get install -y libpam-systemd
		#add falcon systemd
		[ -d /etc/systemd/system/user@2000.service.d ] || mkdir /etc/systemd/system/user@2000.service.d
		echo "[Service]
		Restart=always" > /etc/systemd/system/user@2000.service.d/always.conf
		echo "[Service]
		LimitNOFILE=1000000
		LimitMEMLOCK=infinity" > /etc/systemd/system/user@2000.service.d/limits.conf
		loginctl enable-linger falcon
		systemctl daemon-reload
		#setup service
		echo "[Unit]
		Description=ops meta service
		[Service]
		ExecStart=/opt/open-falcon/agent/falcon-agent -c /opt/open-falcon/agent/cfg.json
		Restart=always
		[Install]
		WantedBy=default.target" > /opt/open-falcon/agent/falcon-agent.service

		systemctl restart user@2000.service
		sudo -u falcon -g falcon XDG_RUNTIME_DIR=/run/user/2000 systemctl --user enable /opt/open-falcon/agent/falcon-agent.service
		ps auxf |grep falcon-agent|grep -v grep|awk '{print $2}'|xargs kill
		sudo -u falcon -g falcon XDG_RUNTIME_DIR=/run/user/2000 systemctl --user restart falcon-agent.service
	else
        ps auxf |grep falcon-agent|grep -v grep|awk '{print $2}'|xargs kill
		if [ -e /etc/service/falcon_agent_run ] ;then
			svc -h /etc/service/falcon_agent_run
		else
			ln -s /opt/open-falcon/agent/falcon_agent_run /etc/service/falcon_agent_run
			svc -h /etc/service/falcon_agent_run
		fi
	fi
}

if grep  2000 /etc/passwd|grep -q falcon ;then
    echo "start reconf"
    #查看是配置文件还是二进制文件
    agent_md5=`md5sum /opt/open-falcon/agent/falcon-agent|cut  -d " " -f1`
    FALCON_AGENT_MD5=`get_md5`
	if [ $FALCON_AGENT_MD5 == $agent_md5 ];then
		reconfig
	else
		update_agent
		reconfig
	fi
else
	echo "start install"
	install_agent
fi