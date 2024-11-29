# Lighting Weather

This runs a webserver that polls the openweatherAPI for the temperature in a given location and calls the phillips hue bridge to change a particular light to a different color depending on the outdoor temperature.

The idea here is to use the color of the lightbulb to indicate whether it's warm enough to go out during this colder months without my heaviest winter jacket (LoL, because it typically wouldn't here in Calgary.)

## Configuration

See `config.yml` for configuration options that are required for the OWM API Key, the Phillips Hue Bridge ID and the Hue Bridge IP address. The `OWM_API_KEY` and `HUE_ID` can be passed as environment variables to avoid committing secrets to version control.
