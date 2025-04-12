update-center.json signer service for Jenkins
===

## Introduction
update-center.json file in the root of a Jenkins mirror contains information needed to keep up with the latest updates 
both of Jenkins core (through WAR file) and all the plugins (which are in the official plugins 'marketplace').

This file is cryptographically signed with SHA512WithRSA SHA1WithRSA algorithms. Any modification of the file (e.g. 
replacement of download URLs) invalidates the signature and thus such a method can't be directly used to update 
Jenkinses in a corporation via the in-house mirror, like Artifactory.

## What is the service intended for?
The sole function of the service is patch update-center.json to override the Jenkins WAR and its plugins download 
locations and to sign the patched file with a private key that is either issued by a corporate CA or with a 
self-signed one.


## Jenkins configuration

* create `${JENKINS_HOME}/update-center-rootCAs` directory (if not exists)
* place signing certificate to `${JENKINS_HOME}/update-center-rootCAs` directory ()
* restart Jenkins server
