#
# You can replace traefik:2.0 with a dev image
# in a quick and dirty way with the following command:
#
# $ make binary && docker build -f traefik-next.Dockerfile  --tag traefik:next .  
#
# Then replace the "2.0" tag with "next"
#
FROM traefik:2.0

ADD dist/traefik /usr/local/bin/traefik
