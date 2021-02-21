# Retry Go

Retry is a go lib  which helps to implement retrying execution mechanics with persistence.


In many situations not everything goes as planned. Retry is designed to run a feature and retry as many times as necessary if it fails.
To retry is also to try each time. Retry is designed to retry indefinitely and implement scheduled job (cron) mechanisms, such as hourly automatic data purging etc.


You can easily implement HTTP POST / GET sends, execute a command, create an email sending mechanic and never miss an event!

As a program must be updated, a persistence in Filesystems, badgerDB, gorm is available to retry exactly where it stopped.

## How to use

## How to implement my own retry

## How to implement a scheduled job