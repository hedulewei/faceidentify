#!/bin/bash
filepath=$(cd "$(dirname "$0")"; pwd)
if [ $# != 1 ]
then
        echo -e "USAGE: $0 option [start | stop ]"
        exit 0;
fi

if [ $1 = "start" ]
then
        pidnum=`ps -ef|grep "$filepath/faceidentify"|grep -v grep|wc -l`
        if [ $pidnum -lt 1 ]
        then
                ($filepath/faceidentify > /dev/null &)
        fi
fi

if [ $1 = "stop" ]
then
        pidnum=`ps -ef|grep "$filepath/faceidentify"|grep -v grep|wc -l`
        if [ $pidnum -lt 1 ]
        then
                echo "no program killed."
        else
                for pid in `ps -ef|grep "$filepath/faceidentify"|grep -v grep|awk '{print $2}'`
                do
                        target_exe=`readlink /proc/$pid/exe | awk '{print $1}'`
                        #如果target_exe非空字符串
                        if [ -n "$target_exe" ]
                        then
                                #发信号10安全退出
                                        kill -10 $pid
                        fi
                done
                sleep 1
                echo "program stoped."
        fi
fi