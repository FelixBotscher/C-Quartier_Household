# C-Quartier_Household
This is the codebase for installing the component on the Raspberry Pi of the Households.

# Requirements
## Running on a Raspberry PI
  - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - [GO](https://go.dev/dl/) (or [GoLand](https://www.jetbrains.com/de-de/go/))
  - [Grafana](https://grafana.com/docs/grafana/latest/setup-grafana/installation/)
  - (only for Consumers) plug in the IR-Reading head and set up the vzlogger tool according to the documentation of ([R.Frank](https://github.com/robertf26/c-power))
  - (only for Consumers) if no IR-Reading head is available install [Python](https://www.python.org/)
  - (only for Producers) [OpenEMS](https://openems.io/) set up a gitpod instance according to [R.Frank](https://github.com/robertf26/c-power)
  - (only for Producers) connect a relay to port 27 (+3.3 V) and port 30 (Ground) with an appropriate relay
## Running on a Computer (simulating a Raspberry PI)
  - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - [GoLand](https://www.jetbrains.com/de-de/go/) (or at least [GO](https://go.dev/dl/))
  - [Grafana](https://grafana.com/docs/grafana/latest/setup-grafana/installation/)
  - (only for Consumers)[Python](https://www.python.org/)
# How to start
1. Clone the project
2. Start Docker Desktop
3. Execute docker-compose up in a terminal inside the git project
4. Start the [Grafana Agent](https://grafana.com/docs/agent/latest/flow/setup/start-agent/) and call [http://localhost:3000](http://localhost:3000/) in a browser
5. Add the data source MySQL Data with the Panels (see chapter #Grafana Configuration)
6. Start VzLogger for turning on the IR-Reading head
7. (Simulation: Start requestDummy.py with executing `uvicorn requestDummy:app --reload`, if fastapi is not installed run: `pip install fastapi`)
8. Start in GoLand the configuration with the name "go build botscher.eu/cquartierhousehold"
# Project Structure
- controller.go is the **MAIN** of C-Quarter_Households which is beeing executed and runs the inizialization bevor executing the **sheduled jobs**.
- config.json keeps the configuartions which have to be modified.
- api.go holds the calls for **VzLogger** / **OpenEMS**
- docker-compose.yml is the configuration to set up the database.
- init.sql is used in the docker-compose.yml to initizalize the database. 
- structs.go holds the needed structs.
- relais.go is used to turn on and off certain **GPIO** ports of the RaspberryPI.
- mqtt.go is the file for sending out the recordered data to the main server.
- chain.go has the functions to create/access/use the C-Chain System. 
# Grafana Configuration
This describes how to set up the Grafana Agent to display live data:
Add the Data Source of Type MySQL in the Configuration with the following parameters:
First Header  | Second Header
------------- | -------------
Host  | localhost:3306
Database  |  server 
User    |  server
Password | server
Create the first new Panel of Type TimeSeries and choose MySQL Data as Data Source and paste following JSON into the JSON Panel:
```
{
  "id": 123125,
  "gridPos": {
    "h": 21,
    "w": 12,
    "x": 0,
    "y": 0
  },
  "type": "timeseries",
  "title": "Registrierter Verbrauch (consumption)",
  "datasource": {
    "uid": "JnXAGBbVz",
    "type": "mysql"
  },
  "fieldConfig": {
    "defaults": {
      "custom": {
        "drawStyle": "line",
        "lineInterpolation": "linear",
        "barAlignment": 0,
        "lineWidth": 1,
        "fillOpacity": 0,
        "gradientMode": "none",
        "spanNulls": false,
        "showPoints": "auto",
        "pointSize": 5,
        "stacking": {
          "mode": "none",
          "group": "A"
        },
        "axisPlacement": "auto",
        "axisLabel": "",
        "axisColorMode": "text",
        "scaleDistribution": {
          "type": "linear"
        },
        "axisCenteredZero": false,
        "hideFrom": {
          "tooltip": false,
          "viz": false,
          "legend": false
        },
        "thresholdsStyle": {
          "mode": "off"
        }
      },
      "color": {
        "mode": "palette-classic"
      },
      "mappings": [],
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "red",
            "value": 80
          }
        ]
      }
    },
    "overrides": []
  },
  "options": {
    "tooltip": {
      "mode": "single",
      "sort": "none"
    },
    "legend": {
      "showLegend": true,
      "displayMode": "list",
      "placement": "bottom",
      "calcs": []
    }
  },
  "targets": [
    {
      "datasource": {
        "type": "mysql",
        "uid": "JnXAGBbVz"
      },
      "dataset": "server",
      "editorMode": "builder",
      "format": "table",
      "rawSql": "SELECT wamount AS \"Verbrauch in W\", time AS \"Zeit\" FROM server.consumption LIMIT 50 ",
      "refId": "A",
      "sql": {
        "columns": [
          {
            "alias": "\"Verbrauch in W\"",
            "parameters": [
              {
                "name": "wamount",
                "type": "functionParameter"
              }
            ],
            "type": "function"
          },
          {
            "alias": "\"Zeit\"",
            "parameters": [
              {
                "name": "time",
                "type": "functionParameter"
              }
            ],
            "type": "function"
          }
        ],
        "groupBy": [
          {
            "property": {
              "type": "string"
            },
            "type": "groupBy"
          }
        ],
        "limit": 50
      },
      "table": "consumption",
      "hide": false
    }
  ]
}
```

Create the second new Panel of Type TimeSeries and choose MySQL Data as Data Source and paste following JSON into the JSON Panel:
```
{
  "id": 123127,
  "gridPos": {
    "h": 21,
    "w": 12,
    "x": 12,
    "y": 0
  },
  "type": "timeseries",
  "title": "Registrierte Erzeugung (feeding)",
  "datasource": {
    "type": "mysql",
    "uid": "JnXAGBbVz"
  },
  "fieldConfig": {
    "defaults": {
      "custom": {
        "drawStyle": "line",
        "lineInterpolation": "linear",
        "barAlignment": 0,
        "lineWidth": 1,
        "fillOpacity": 0,
        "gradientMode": "none",
        "spanNulls": false,
        "showPoints": "auto",
        "pointSize": 5,
        "stacking": {
          "mode": "none",
          "group": "A"
        },
        "axisPlacement": "auto",
        "axisLabel": "",
        "axisColorMode": "text",
        "scaleDistribution": {
          "type": "linear"
        },
        "axisCenteredZero": false,
        "hideFrom": {
          "tooltip": false,
          "viz": false,
          "legend": false
        },
        "thresholdsStyle": {
          "mode": "off"
        }
      },
      "color": {
        "mode": "palette-classic"
      },
      "mappings": [],
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "red",
            "value": 80
          }
        ]
      }
    },
    "overrides": []
  },
  "options": {
    "tooltip": {
      "mode": "single",
      "sort": "none"
    },
    "legend": {
      "showLegend": true,
      "displayMode": "list",
      "placement": "bottom",
      "calcs": []
    }
  },
  "targets": [
    {
      "dataset": "server",
      "datasource": {
        "type": "mysql",
        "uid": "JnXAGBbVz"
      },
      "editorMode": "builder",
      "format": "table",
      "rawSql": "SELECT wamount AS \"Erzeugung in W\", time AS \"Zeit\" FROM server.feeding LIMIT 50 ",
      "refId": "A",
      "sql": {
        "columns": [
          {
            "alias": "\"Erzeugung in W\"",
            "parameters": [
              {
                "name": "wamount",
                "type": "functionParameter"
              }
            ],
            "type": "function"
          },
          {
            "alias": "\"Zeit\"",
            "parameters": [
              {
                "name": "time",
                "type": "functionParameter"
              }
            ],
            "type": "function"
          }
        ],
        "groupBy": [
          {
            "property": {
              "type": "string"
            },
            "type": "groupBy"
          }
        ],
        "limit": 50
      },
      "table": "feeding"
    }
  ]
}
```
