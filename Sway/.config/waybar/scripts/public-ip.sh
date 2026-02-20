#!/bin/bash

public_ip=$(curl -s ifconfig.me)
echo -n $public_ip  # [ ] TODO: Double quote to prevent globbing and word splitting.; Double quote to prevent globbing and word splitting.
