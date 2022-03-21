# Treasury

My personal digital asset management system


## Data storage path

The following directory must exist and be writeable:

	/var/lib/treasuryd


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
