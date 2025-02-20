# Light Scheduler

This project is a Go-based microservice for turning Phillips Hue lights on and off on a schedule. The service takes a list of lights and a night start and night time. At the night start time, it powers all the lights off and then waits until the night end time to power them back on.

A `/turnOn` endpoint also listens to turn the lights on when a request is received. The `/turnOff` endpoint will likewise power the lights off when called.
