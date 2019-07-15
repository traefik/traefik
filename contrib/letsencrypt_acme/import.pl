#!/usr/bin/perl -w
use strict;

use File::Slurp qw/read_file write_file/;
use JSON::XS;
use MIME::Base64 qw/decode_base64 encode_base64/;
use Getopt::Long qw/GetOptions/;
use Pod::Usage qw/pod2usage/;

my $coder = JSON::XS->new->ascii->pretty;

my $acme    = "acme_part.json";
my $chainIn = "chain.pem";
my $certIn  = "cert.pem";
my $keyIn   = "privkey.pem";
my $help = 0;
GetOptions(
    "file:s"  => \$acme,
    "chain:s" => \$chainIn,
    "cert:s"  => \$certIn,
    "key:s"   => \$keyIn,
    "help|?"  => \$help
    );
pod2usage(1) if $help;
if( $acme    =~ m|/$| || -d $acme    ) { $acme    = "$acme/acme.json"; }
if( $chainIn =~ m|/$| || -d $chainIn ) { $chainIn = "$chainIn/chain.pem"; }
if( $certIn  =~ m|/$| || -d $certIn  ) { $certIn  = "$certIn/cert.pem"; }
if( $keyIn   =~ m|/$| || -d $keyIn   ) { $keyIn   = "$keyIn/privkey.pem"; }

import( $acme, $chainIn, $certIn, $keyIn );

sub import {
    my ( $acme, $chainFile, $certFile, $keyFile ) = @_;
    if( ! -e $chainFile ) {
        die "Chain file \"$chainFile\" not found";
    }
    if( ! -e $certFile ) {
        die "Chain file \"$certFile\" not found";
    }
    if( ! -e $keyFile ) {
        die "Chain file \"$keyFile\" not found";
    }

    my $certRaw  = read_file( $certFile );
    if( $certRaw !~ m/^-----BEGIN CERTIFICATE-----\n.+-----END CERTIFICATE-----\n?$/s ) {
        die "Cert \"$certFile\" is not armored properly";
    }
    my $chainRaw = read_file( $chainFile );
    if( $chainRaw !~ m/^-----BEGIN CERTIFICATE-----\n.+-----END CERTIFICATE-----\n?$/s ) {
        die "Cert \"$chainFile\" is not armored properly";
    }
    my $keyRaw   = read_file( $keyFile );
    
    my $subject = `openssl x509 -subject -noout -in $certFile`;
    if( $subject =~ m/^(subject=)?CN = ([a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+)$/ ) {
        my $domain = $2;    
    
        my $keyEncoded  = encode_base64( $keyRaw );
        my $certEncoded = encode_base64( "$certRaw\n\n$chainRaw" );
        $keyEncoded  =~ s/\n//g;
        $certEncoded =~ s/\n//g;
        my $json = {
            Certificates => [
                {
                    Domain => {
                        Main => $domain,
                        SANs => undef
                    },
                    Certificate => $certEncoded,
                    Key => $keyEncoded
                }
            ]
        };
        print "Portion of acme.json written to \"$acme\"\n";
        write_file( $acme, $coder->encode( $json ) );
    }
    else {
        die "Subject of cert \"$certFile\" does not look like a domain: \"$subject\"\n";
    }
}

__END__

=head1 NAME

LetsEncypt acme.json certificate importer

=head1 SYNOPSIS

import.pl [options]

  Options:
    --file   Path to save acme_part.json
    --chain  Path to chain.pem
    --cert   Path to cert.pem
    --key    Path to privkey.pem

=cut