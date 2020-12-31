#!/bin/bash

set -eux

curl -LOs https://github.com/stevenwilkin/treasury/releases/download/latest/treasuryd
chmod a+x treasuryd
sudo mv treasuryd /usr/local/bin
sudo systemctl restart treasuryd
