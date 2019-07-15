#!/usr/bin/perl -w
use strict;

use File::Slurp qw/read_file write_file/;
use JSON::XS;
use MIME::Base64 qw/decode_base64 encode_base64/;
use File::Copy qw/mv/;
use Getopt::Long qw/GetOptions/;
use Pod::Usage qw/pod2usage/;

my $coder = JSON::XS->new->ascii->pretty;

my $acme     = "acme.json";
my $chainOut = "chain.pem";
my $certOut  = "cert.pem";
my $keyOut   = "privkey.pem";
my $certNum  = -1;
my $help = 0;
GetOptions(
    "file:s"    => \$acme,
    "chain:s"   => \$chainOut,
    "cert:s"    => \$certOut,
    "key:s"     => \$keyOut,
    "certnum:i" => \$certNum,
    "help|?"    => \$help
    );
pod2usage(1) if $help;
if( $acme     =~ m|/$| || -d $acme     ) { $acme     = "$acme/acme.json"; }
if( $chainOut =~ m|/$| || -d $chainOut ) { $chainOut = "$chainOut/chain.pem"; }
if( $certOut  =~ m|/$| || -d $certOut  ) { $certOut  = "$certOut/cert.pem"; }
if( $keyOut   =~ m|/$| || -d $keyOut   ) { $keyOut   = "$keyOut/privkey.pem"; }

if( $certNum == -1 ) {
    listDomains( $acme );
}
else {
    export( $acme, $chainOut, $certOut, $keyOut, $certNum );
}

sub listDomains {
    my $acme = shift;
    my $raw_json = read_file( $acme );
    my $json = $coder->decode( $raw_json );
    my $certs = $json->{Certificates};
    
    print "Sites in acme.json:\n";
    my $n = 0;
    for my $cert ( @$certs ) {
        my $domain = $cert->{Domain};
        my $main = $domain->{Main};
        print "$n. $main\n";
        $n++;
    }
}

sub export {
    my ( $acme, $chainFile, $certFile, $keyFile, $certNum ) = @_;
    if( ! -e $acme ) {
        die "Acme file \"$acme\" not found";
    }
    my $raw_json = read_file( $acme );
    
    my $json = $coder->decode( $raw_json );
    
    my $certs = $json->{Certificates};
    
    #for my $cert ( @$certs ) {
        my $certN = $certs->[$certNum];
        my $key = $certN->{Key};
        my $cert = $certN->{Certificate};
        $key = decode_base64( to_lines( $key ) );
        $cert = decode_base64( to_lines( $cert ) );
        
        my $n = 1;
        while( $cert =~ m/(-----BEGIN CERTIFICATE-----\n.+?\n-----END CERTIFICATE-----)/sg ) {
            write_file("part$n.pem", $1 );
            my $subject = `openssl x509 -subject -noout -in part$n.pem`;
            # This matches the intermediate cert of Let's Encrypt production
            if( $subject =~ m/O = Let's Encrypt/ ) {
                print "Writing Let's Encrypt intermediate cert to $chainFile\n";
                mv "part$n.pem", $chainFile;
            }
            # This matches the intermediate cert of Let's Encrypt staging
            elsif( $subject =~ m/Intermediate/ ) {
                print "Writing intermediate cert to $chainFile\n";
                mv "part$n.pem", $chainFile;
            }
            elsif( $subject =~ m/^(subject=)?CN = ([a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+)$/ ) {
                my $domain = $2;
                print "Writing \"$domain\" cert to $certFile\n";
                mv "part$n.pem", $certFile;
            }
            else {
                print "Unrecognized CN present: \"$subject\"\n";
            }
            $n++;
        }
        
        print "Writing private key to $keyFile\n";
        write_file( $keyFile, $key );
        #print `openssl rsa -text -noout -in key.pem`;
    #}
}

sub to_lines {
    $_ = shift;
    s/(.{64})/$1\n/g;
    return $_;
}

__END__

=head1 NAME

LetsEncypt acme.json certificate exporter

=head1 SYNOPSIS

export.pl [options]

  Options:
    --file    Path to acme.json
    --chain   Path to save chain.pem as
    --cert    Path to save cert.pem as
    --key     Path to save privkey.pem as
    --certnum Number of domain within acme.json

Example:

  ./export.pl --file /path/to/acme.json
    List the domains within acme.json
    
  ./export.pl --file /path/to/acme.json --certnum 0
    Save out the pem files for site number 0

=cut