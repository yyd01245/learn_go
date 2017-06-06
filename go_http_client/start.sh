#!/bin/bash - 
#===============================================================================
#
#          FILE:  start.sh
# 
#         USAGE:  ./start.sh 
# 
#   DESCRIPTION:  
# 
#       OPTIONS:  ---
#  REQUIREMENTS:  ---
#          BUGS:  ---
#         NOTES:  ---
#        AUTHOR: Yandong Yan (A program fan), yandong.yan@upai.com or (yydgame@163.com)
#       COMPANY: Upyun
#       CREATED: 2017/04/17 18时06分49秒 CST
#      REVISION:  ---
#===============================================================================

set -o nounset                              # Treat unset variables as an error

if [ $# != 5 ]
then
    echo "param number $#"
    echo "********************************************"
    echo "useage ./start.sh 2000 5 2 1 http://127.0.0.1:8088/uprtc"
    echo "param explain"
    echo "2000 is base clientid "
    echo "5 is loop number"
    echo "2 is type, 1 for publisher 2 for listener"
    echo "1 is step, when loop > 1 ,the next clientid = lastclientid + step "
    echo "http://127.0.0.1:8088/uprtc is uprtc server ip"
    echo "********************************************"
    exit
fi
Base_ID=$1
Loop_Num=$2
clientType=$3
step=$4
serverAddr=$5

echo "client: $Base_ID loop number: $Loop_Num type:$clientType step:$step serverip:$serverAddr"



for a in `seq $Loop_Num`
do
	((val=(($a-1))*$step))
    echo $val
    ((clientid=$Base_ID+$val))
    #aclientid=`expr $Base_ID+$val`
    echo $clientid
	./go_http_client_janus -p $clientType -u $serverAddr  -c $clientid -s $step  -q 512 >> /dev/null &
	# >> /dev/null
    sleep 0.3
#	echo "loop time $a"
done
