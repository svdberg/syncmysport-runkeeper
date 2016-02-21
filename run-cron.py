#!/usr/bin/env python

# run-cron.py
# sets environment variable crontab fragments and runs cron

import os
from subprocess import call
import fileinput

# read docker environment variables and set them in the appropriate crontab fragment
environment_variable = os.environ["DATABASE_URL"]

for line in fileinput.input("/etc/cron.d/syncmysport-cron",inplace=1):
    print line.replace("XXXXXXX", environment_variable)

environment_variable = os.environ["CLEARDB_DATABASE_URL"]

for line in fileinput.input("/etc/cron.d/syncmysport-cron",inplace=1):
    print line.replace("YYYYYYY", environment_variable)


args = ["cron","-f", "-L 15"]
call(args)
