# Networking a Temperature Monitor

The first project is a networked temperature monitor. This will be implemented using a Raspberry Pi Pico W, which has an onboard temperature sensor. The sensor will report the ambient temperature around the sensor. A webserver running on the Pico W will be polled by a Prometheus server running on a Raspberry Pi.

The formatted JSON data from the Pico W is consumed by the Prometheus server and visualized by a Grafana instance also running on the Raspberry Pi.

## Project Requirements

1. Raspberry Pi Pico W: microcontroller with onboard Wi-Fi and a temperature sensor to report ambient temperature
2. A Raspberry Pi (or any server) running a Prometheus exporter to scrape data from the Pico W
