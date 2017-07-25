all:
	find * -type d -exec bash -exc "cd {}; go build . || kill $${PPID}" \;
