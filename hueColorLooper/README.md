# Hue Color Looper

This project uses the hue bridge API to set phillips hue lights to the colorloop mode when a post request is received to a microservice endpoint, with a list of light names passed in the body of the request.

The service can be configured with a yaml file to define the actual names of lights to control, as well as details of the hue bridge device and the api token.

