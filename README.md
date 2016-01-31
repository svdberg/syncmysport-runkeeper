[![Build Status](https://travis-ci.org/svdberg/syncmysport-runkeeper.svg?branch=master)](https://travis-ci.org/svdberg/syncmysport-runkeeper)

From Strava to Runkeeper
========================

Reads activities from Strava and copies them to Runkeeper in a hopefully smart way.


TODO
----

- Proper resilence, error handling
- Swimming activities apprently use a different duration, and therefore fail to be recognized as the same Activity
- Refactor sync worker code (and a shitload of other parts)
- Bi-directional sync
~~- OAUth web interface + cookie storing of uid if already exists~~
~~- Bootstap to make it pretty~~
~~- Patch the freakin TZ troubles in the c9s runkeeper lib~~

Website
-------

This application is available at [www.syncmysport.com](http://www.syncmysport.com/)
