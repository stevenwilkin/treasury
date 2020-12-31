#!/bin/bash

set -eux

curl -LOs https://github.com/stevenwilkin/treasury/releases/download/current/treasuryd.tar
sudo tar -xf treasuryd.tar -C /usr/local/bin
sudo systemctl restart treasuryd
