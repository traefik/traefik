#!/bin/bash
# Copyright (c) 2017 Brian 'redbeard' Harrington <redbeard@dead-city.org>
#
# dumpcerts.sh - A simple utility to explode a Traefik acme.json file into a
#                directory of certificates and a private key
#
# Usage - dumpcerts.sh /etc/traefik/acme.json /etc/ssl/
#
# Dependencies - 
#   util-linux
#   openssl
#   jq
# The MIT License (MIT)
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

# Exit codes:
# 1 - A component is missing or could not be read
# 2 - There was a problem reading acme.json
# 4 - The destination certificate directory does not exist
# 8 - Missing private key

USAGE="dumpcerts.sh <path to acme> <destination cert directory>"

# Allow us to exit on a missing jq binary
exitJQ() {
	echo ""
	echo "You must have the binary 'jq' to use this."
	echo "jq is available at: https://stedolan.github.io/jq/download/"
	echo ""
	echo ${USAGE}
	exit 1
}


badACME() {
	echo ""
	echo "There was a problem parsing your acme.json file."
	echo ""
	echo ${USAGE}
	exit 2
}


acme="${1}"
certdir="${2}"

if [ ! -r "${acme}" ]; then
	echo ""
	echo "There was a problem reading from '${acme}'"
	echo "We need to read this file to explode the JSON bundle... exiting."
	echo ""
	echo ${USAGE}
	exit 1
fi	


if [ ! -d "${certdir}" ]; then
	echo ""
	echo "Path ${certdir} does not seem to be a directory"
	echo "We need a directory in which to explode the JSON bundle... exiting."
	echo ""
	echo ${USAGE}
	exit 4
fi	

jq=$(which jq) || exitJQ

priv=$(${jq} -r '.PrivateKey' ${acme}) || badACME

if [ ! -n "${priv}" ]; then
	echo ""
	echo "There didn't seem to be a private key in ${acme}."
	echo "Please ensure that there is a key in this file and try again."
	exit 8
fi

# If they do not exist, create the needed subdirectories for our assets
# and place each in a variable for later use, normalizing the path
mkdir -p "${certdir}"/{certs,private}

pdir=$(realpath "${certdir}/private/")
cdir=$(realpath "${certdir}/certs/")

# Save the existing umask, change the default mode to 600, then
# after writing the private key switch it back to the default
oldumask=$(umask)
umask 177
# For some reason traefik stores the private key in stripped base64 format, but
# the certificates bundled as a base64 object without stripping headers.  This
# normalizes the headers and formatting.
#
# In testing this out it was a balance between the following mechanisms:
# gawk:
#  echo ${priv} | awk 'BEGIN {print "-----BEGIN RSA PRIVATE KEY-----"}
#     {gsub(/.{64}/,"&\n")}1
#     END {print "-----END RSA PRIVATE KEY-----"}' > "${pdir}/letsencrypt.key"
#
# openssl:
# echo -e "-----BEGIN RSA PRIVATE KEY-----\n${priv}\n-----END RSA PRIVATE KEY-----" \
#   | openssl rsa -inform pem -out "${pdir}/letsencrypt.key" 
#
# and sed:
# echo "-----BEGIN RSA PRIVATE KEY-----" > "${pdir}/letsencrypt.key"
# echo ${priv} | sed 's/(.{64})/\1\n/g' >> "${pdir}/letsencrypt.key"
# echo "-----END RSA PRIVATE KEY-----" > "${pdir}/letsencrypt.key"
#
# In the end, openssl was chosen because most users will need this script
# *because* of openssl combined with the fact that it will refuse to write the
# key if it does not parse out correctly. The other mechanisms were left as
# comments so that the user can choose the mechanism most appropriate to them.
echo -e "-----BEGIN RSA PRIVATE KEY-----\n${priv}\n-----END RSA PRIVATE KEY-----" \
   | openssl rsa -inform pem -out "${pdir}/letsencrypt.key" 

umask ${oldumask}

# Process the certificates for each of the domains in acme.json
for domain in $(jq -r '.DomainsCertificate.Certs[].Certificate.Domain' acme.json); do
	# Traefik stores a cert bundle for each domain.  Within this cert 
	# bundle there is both proper the certificate and the Let's Encrypt CA
	echo "Extracting cert bundle for ${domain}"
	cert=$(jq -r --arg domain "$domain" '.DomainsCertificate.Certs[].Certificate |
         	select (.Domain == $domain )| .Certificate' ${acme}) || badACME
	echo ${cert} | base64 --decode > "${cdir}/${domain}.pem"
done
