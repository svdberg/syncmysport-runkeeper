FROM golang:onbuild

RUN apt-get update
RUN apt-get install -y cron

# Add crontab file in the cron directory
ADD crontab /etc/cron.d/syncmysport-cron

# Give execution rights on the cron job
RUN chmod 0644 /etc/cron.d/syncmysport-cron

# Create the log file to be able to run tail
RUN touch /var/log/cron.log
