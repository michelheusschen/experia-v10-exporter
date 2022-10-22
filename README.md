# experia-v10-exporter
A [prometheus](https://prometheus.io) exporter for getting some metrics of an Experia Box v10 (H369A)

Disclaimer: this is the first thing I ever did in Go and I have no clue what I'm doing

## Installation
If you have a working Go installation, getting the binary should be as simple as

```
go get github.com/michelheusschen/experia-v10-exporter
```

To run the exporter as a Docker container, provide the `EXPERIA_V10_ROUTER_USERNAME` and `EXPERIA_V10_ROUTER_PASSWORD` environment variables and run

```
docker compose up -d
```

## Usage
```plain
$ experia-v10-exporter
```

The following environment variables are required:
```
EXPERIA_V10_LISTEN_ADDR=localhost:9684 
EXPERIA_V10_TIMEOUT=10 
EXPERIA_V10_ROUTER_IP=192.168.2.254
EXPERIA_V10_ROUTER_USERNAME=Admin 
EXPERIA_V10_ROUTER_PASSWORD="PASSWORD"
```

## Metrics
The following metrics are currently returned:
```
# HELP experia_v10_auth_errors_total Counts number of authentication errors encountered by the collector.
# TYPE experia_v10_auth_errors_total counter
experia_v10_auth_errors_total 15
# HELP experia_v10_dsl All dsl related metadata.
# TYPE experia_v10_dsl counter
experia_v10_dsl{value="Atuc_fec_errors"} 0
experia_v10_dsl{value="BitSwapEnable"} 1
experia_v10_dsl{value="DownCrc_errors"} 0
experia_v10_dsl{value="DownInterleaveDelay"} 0
experia_v10_dsl{value="DownInterleavedepth"} 4
experia_v10_dsl{value="DownstreamInp"} 790
experia_v10_dsl{value="Downstream_attenuation"} 241
experia_v10_dsl{value="Downstream_current_rate"} 32714
experia_v10_dsl{value="Downstream_max_rate"} 33652
experia_v10_dsl{value="Downstream_noise_margin"} 51
experia_v10_dsl{value="Downstream_power"} 137
experia_v10_dsl{value="Enable"} 1
experia_v10_dsl{value="Fec_errors"} 18846
experia_v10_dsl{value="Showtime_start"} 791165
experia_v10_dsl{value="UpCrc_errors"} 0
experia_v10_dsl{value="UpInterleaveDelay"} 0
experia_v10_dsl{value="UpInterleaveDepth"} 2
experia_v10_dsl{value="UpstreamInp"} 690
experia_v10_dsl{value="Upstream_attenuation"} 93
experia_v10_dsl{value="Upstream_current_rate"} 2858
experia_v10_dsl{value="Upstream_max_rate"} 2888
experia_v10_dsl{value="Upstream_noise_margin"} 57
experia_v10_dsl{value="Upstream_power"} 47

# HELP experia_v10_interface_received_bytes_total The total number of bytes received on the interface
# TYPE experia_v10_interface_received_bytes_total counter
experia_v10_interface_received_bytes_total{alias="LAN1",id="IGD.LD1.ETH1"} 6.15211195e+08
experia_v10_interface_received_bytes_total{alias="LAN2",id="IGD.LD1.ETH2"} 0
experia_v10_interface_received_bytes_total{alias="LAN3",id="IGD.LD1.ETH3"} 4.213383323e+09
experia_v10_interface_received_bytes_total{alias="LAN4",id="IGD.LD1.ETH4"} 0

# HELP experia_v10_interface_sent_bytes_total The total number of bytes transmitted out of the interface
# TYPE experia_v10_interface_sent_bytes_total counter
experia_v10_interface_sent_bytes_total{alias="LAN1",id="IGD.LD1.ETH1"} 1.827165148e+09
experia_v10_interface_sent_bytes_total{alias="LAN2",id="IGD.LD1.ETH2"} 0
experia_v10_interface_sent_bytes_total{alias="LAN3",id="IGD.LD1.ETH3"} 4.216932101e+09
experia_v10_interface_sent_bytes_total{alias="LAN4",id="IGD.LD1.ETH4"} 0

# HELP experia_v10_scrape_errors_total Counts the number of scrape errors by this collector.
# TYPE experia_v10_scrape_errors_total counter
experia_v10_scrape_errors_total 0

# HELP experia_v10_up Shows if the Experia Box V10 is deemed up by the collector.
# TYPE experia_v10_up gauge
experia_v10_up 1
```
