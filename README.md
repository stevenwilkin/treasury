# Treasury

My personal digital asset management system


## Systemd service

Copy the service unit file to the configuration directory:

	cp treasuryd.service /etc/systemd/system

Enable and start the service:

	systemctl enable treasuryd
	systemctl start treasuryd


## Rsyslog forwarding

	cat << EOF > /etc/rsyslog.d/10-treasuryd.conf
	if $programname == 'treasuryd' then @HOST.papertrailapp.com:PORT
	& ~
	EOF
	systemctl restart rsyslog
