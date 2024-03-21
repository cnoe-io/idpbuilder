#!/bin/bash
sudo sysctl fs.inotify.max_user_instances=1280
sudo sysctl fs.inotify.max_user_watches=655360

# more agressive settings
#sudo sysctl -w fs.inotify.max_user_watches=2099999999
#sudo sysctl -w fs.inotify.max_user_instances=2099999999
#sudo sysctl -w fs.inotify.max_queued_events=2099999999
