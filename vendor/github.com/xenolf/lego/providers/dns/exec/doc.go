/*
Package exec implements a manual DNS provider which runs a program for adding/removing the DNS record.

The file name of the external program is specified in the environment variable `EXEC_PATH`.
When it is run by lego, three command-line parameters are passed to it:
The action ("present" or "cleanup"), the fully-qualified domain name, the value for the record and the TTL.

For example, requesting a certificate for the domain 'foo.example.com' can be achieved by calling lego as follows:

	EXEC_PATH=./update-dns.sh \
		lego --dns exec \
		--domains foo.example.com \
		--email invalid@example.com run

It will then call the program './update-dns.sh' with like this:

	./update-dns.sh "present" "_acme-challenge.foo.example.com." "MsijOYZxqyjGnFGwhjrhfg-Xgbl5r68WPda0J9EgqqI" "120"

The program then needs to make sure the record is inserted.
When it returns an error via a non-zero exit code, lego aborts.

When the record is to be removed again,
the program is called with the first command-line parameter set to "cleanup" instead of "present".

If you want to use the raw domain, token, and keyAuth values with your program, you can set `EXEC_MODE=RAW`:

	EXEC_MODE=RAW \
	EXEC_PATH=./update-dns.sh \
		lego --dns exec \
		--domains foo.example.com \
		--email invalid@example.com run

It will then call the program './update-dns.sh' like this:

	./update-dns.sh "present" "foo.example.com." "--" "some-token" "KxAy-J3NwUmg9ZQuM-gP_Mq1nStaYSaP9tYQs5_-YsE.ksT-qywTd8058G-SHHWA3RAN72Pr0yWtPYmmY5UBpQ8"

NOTE:
The `--` is because the token MAY start with a `-`, and the called program may try and interpret a - as indicating a flag.
In the case of urfave, which is commonly used,
you can use the `--` delimiter to specify the start of positional arguments, and handle such a string safely.
*/
package exec
