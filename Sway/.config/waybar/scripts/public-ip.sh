#!/bin/bash

public_ip=$(curl -s ifconfig.me)
echo -n $public_ip
