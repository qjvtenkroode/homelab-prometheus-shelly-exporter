# Prometheus Shelly Exporter
Getting scraping basic information from Shelly devices using a probe endpoint to target specific Shelly devices.

The [blackbox_exporter](https://github.com/prometheus/blackbox_exporter) and [snmp_exporter](https://github.com/prometheus/snmp_exporter) have been a source of examples and inspiration since I wanted an exporter without any configuration. Everything should be configured from the scraping Prometheus instance using label rewriting.

## Implemented metrics
Currently implemented metrics are:
 - Power
 - Total Power
 - Temperature

## TODO
- Cleanup code
- Do proper error handling, bail early and don't crash the exporter...
- Implement more metrics:
    - Uptime
    - Switch state
    - ...

## Configuration
Sample configuration that uses relabling to create a <machine>.qkroode.nl:8080/probe?target=shelly1pm-b1e281.qkroode.nl target:

    scrape_configs:
      - job_name: shelly-exporter
        static_configs:
          - targets: ['machine.qkroode.nl:8080']
      - job_name: shellies
        static_configs:
          - targets:
            - shelly1pm-b1e281.qkroode.nl
            - shelly1pm-6090fc.qkroode.nl
        metrics_path: /probe
        relabel_configs:
          - source_labels: [__address__]
            target_label: __param_target
          - source_labels: [__param_target]
            target_label: instance
          - target_label: __address__
            replacement: machine.qkroode.nl:8080

