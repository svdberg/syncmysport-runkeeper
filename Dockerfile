FROM golang:onbuild

RUN apt-get update && apt-get install -y wget
RUN wget https://github.com/jwilder/dockerize/releases/download/v0.1.0/dockerize-linux-amd64-v0.1.0.tar.gz
RUN tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.1.0.tar.gz

# Add crontab file in the cron directory
ADD crontab /etc/cron.d/syncmysport-cron

# Give execution rights on the cron job
RUN chmod 0644 /etc/cron.d/syncmysport-cron

# Create the log file to be able to run tail
RUN touch /var/log/cron.log

# Run the command on container startup
CMD cron && tail -f /var/log/cron.log
