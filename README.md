# C-Quartier_Household
This is the codebase for installing the component on the Raspberry Pi of the Households.

## Requirements
### Running on a Raspberry PI
  - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - [GO](https://go.dev/dl/) (or [GoLand](https://www.jetbrains.com/de-de/go/))
  - [Grafana](https://grafana.com/docs/grafana/latest/setup-grafana/installation/)
  - (only for Consumer) plug in the IR-Reading head and set up the vzlogger tool according to the documentation of ([R.Frank](https://github.com/robertf26/c-power))
  - (only for Consumers) if no IR-Reading head is available install [Python](https://www.python.org/)
  - (only for Producers) [OpenEMS](https://openems.io/) set up a gitpod instance according to [R.Frank](https://github.com/robertf26/c-power)
### Running on a Computer (simulating a Raspberry PI)
  - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - [GoLand](https://www.jetbrains.com/de-de/go/) (or at least [GO](https://go.dev/dl/))
  - [Grafana](https://grafana.com/docs/grafana/latest/setup-grafana/installation/)
  - (only for Consumer)[Python](https://www.python.org/)
## How to start
1. Clone the project
2. Start Docker Desktop
3. Execute docker-compose up in a terminal inside the git project
4. Start the [Grafana Agent](https://grafana.com/docs/agent/latest/flow/setup/start-agent/) and call [http://localhost:3000](http://localhost:3000/) in a browser and set up Grafana as follows
## Project Structure
- controller.go is the MAIN of C-Quarter_Households
- config.json keeps the configuartions
- api.go holds the calls for VzLogger / OpenEMS
- docker-compose.yml is the configuration to set up the database
- init.sql keeps the 
- structs.go holds the needed Structs
- relais.go 
