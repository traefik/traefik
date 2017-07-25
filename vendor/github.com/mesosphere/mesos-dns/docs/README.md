# Mesos-DNS Docs Website

You can view the online version of these documents at [https://mesosphere.github.io/mesos-dns](https://mesosphere.github.io/mesos-dns).

## Run it locally

Ensure you have installed everything listed in the dependencies section before
following the instructions.

### Dependencies

* [Bundler](http://bundler.io/)
* [Node.js](http://nodejs.org/) (for compiling assets)
* Python
* Ruby
* [RubyGems](https://rubygems.org/)

### Instructions

1. Install packages needed to generate the site

    * On Linux:

            apt-get install ruby-dev make autoconf nodejs nodejs-legacy python-dev
    * On Mac OS X:
    
            brew install node

2. Clone the Mesos-DNS repository

3. Change into the "docs" directory where docs live

        cd docs

4. Install Bundler

        sudo gem install bundler

5. Install the bundle's dependencies

        bundle install

6. Start the web server

        bundle exec jekyll serve --watch

7. Visit the site at
   [http://localhost:4000/mesos-dns/](http://localhost:4000/mesos-dns/)

## Deploying the site

1. Clone a separate copy of the Mesos-DNS repo as a sibling of your normal
   Mesos-DNS project directory and name it "mesos-dns-gh-pages".

        git clone git@github.com:mesosphere/mesos-dns.git mesos-dns-gh-pages

2. Check out the "gh-pages" branch.

        cd /path/to/mesos-dns-gh-pages
        git checkout -b gh-pages

3. Copy the contents of the "docs" directory in master to the root of your
   mesos-dns-gh-pages directory.

        cd /path/to/mesos-dns
        cp -r docs/** ../mesos-dns-gh-pages

4. Change to the mesos-dns-gh-pages directory, commit, and push the changes

        cd /path/to/mesos-dns-gh-pages
        git commit . -m "Syncing docs with master branch"
        git push
