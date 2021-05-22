#!/bin/bash

# go build

if [ $1 == 'client' ]
then
	./tuntap client 4 192 168 56 1 35 &
	ifconfig wg2 172.16.0.2/30
else
	./tuntap server 4 192 168 56 1 35 &
	ifconfig wg2 172.16.0.1/30
fi
ifconfig wg2 mtu 1460 txqueuelen 2000

