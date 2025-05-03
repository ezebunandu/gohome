# Hue Color Looper

This project uses the hue bridge API to set phillips hue lights to the colorloop mode when a post is received.

Requests to `/colorloop/{light_name}` will put the specified light into colorloop mode, if `light_name` is a valid name, else, it returns a 400 response.

Requests to `/colorloop/all` will put all the configured lights in colorloop mode, returning a 200 response.

## To-Dos

- make the lights that can be controlled configurable with a yaml config file
