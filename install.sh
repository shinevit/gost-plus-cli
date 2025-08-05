sudo cp ./gost-tunnel /usr/local/bin/
sudo cp ./gost.plus.service /etc/systemd/system/gost.plus.service
sudo systemctl daemon-reload
sudo systemctl enable gost.plus
sudo systemctl start gost.plus
sudo systemctl status gost.plus
