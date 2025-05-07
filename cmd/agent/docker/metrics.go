package main

import (
	"context"
	"fmt"
	"github.com/glasskube/distr/api"
	hmr "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

type noopHost struct {
	reportFunc func(event *componentstatus.Event)
}

func (nh *noopHost) GetExtensions() map[component.ID]component.Component {
	return nil
}

func (nh *noopHost) Report(event *componentstatus.Event) {
	nh.reportFunc(event)
}

func startMetrics(ctx context.Context) {
	factory := hmr.NewFactory()
	cfg := factory.CreateDefaultConfig().(*hmr.Config)

	if retrieved, err := confmap.NewRetrievedFromYAML([]byte(`
collection_interval: 10s
scrapers:
  cpu:
    metrics:
      system.cpu.time:
        enabled: false
      system.cpu.logical.count:
        enabled: true
      system.cpu.utilization:
        enabled: true
  memory:
    metrics:
      system.memory.utilization:
        enabled: true
      system.memory.limit:
        enabled: true
`)); err != nil {
		panic(err)
	} else if cnf, err := retrieved.AsConf(); err != nil {
		panic(err)
	} else if err := cfg.Unmarshal(cnf); err != nil {
		panic(err)
	}

	consmr, err := consumer.NewMetrics(func(ctx context.Context, md pmetric.Metrics) error {
		fmt.Fprintf(os.Stderr, "consuming ---------------------------\n")
		var cores int64
		var cpuUsed float64
		var memoryTotal int64
		var memoryUsed float64
		for _, resourceMetrics := range md.ResourceMetrics().All() {
			for _, scopeMetrics := range resourceMetrics.ScopeMetrics().All() {
				for _, metric := range scopeMetrics.Metrics().All() {
					switch metric.Name() {
					case "system.cpu.logical.count":
						dataPoint := metric.Sum().DataPoints().At(metric.Sum().DataPoints().Len() - 1)
						cores = dataPoint.IntValue()

					case "system.cpu.utilization":
						// each datapoint has attributes cpu:<cpu-name> and state:<one of 8 states>
						// the value describes the utilization of this exact cpu in this state
						// for now we add up system+user states for each cpu and divide by the number of cpus
						for _, dataPoint := range metric.Gauge().DataPoints().All() {
							if state, ok := dataPoint.Attributes().Get("state"); ok {
								if state.Str() == "user" || state.Str() == "system" {
									// TODO do other states make sense too?
									cpuUsed = cpuUsed + dataPoint.DoubleValue()
								}
							}
						}

					case "system.memory.utilization":
						for _, dataPoint := range metric.Gauge().DataPoints().All() {
							if state, ok := dataPoint.Attributes().Get("state"); ok {
								if state.Str() == "used" {
									// TODO maybe make calculation more specific with the other possible states
									memoryUsed = memoryUsed + dataPoint.DoubleValue()
								}
							}
						}

					case "system.memory.limit":
						dataPoint := metric.Sum().DataPoints().At(metric.Sum().DataPoints().Len() - 1)
						memoryTotal = dataPoint.IntValue()

					}
				}
			}
		}
		if cores != 0 {
			usage := cpuUsed / float64(cores)
			fmt.Fprintf(os.Stderr, "cpu usage: %v (%v cores)\n", usage, cores)
		}

		fmt.Fprintf(os.Stderr, "memory usage: %v (%v total)\n", memoryUsed, memoryTotal)

		if err := client.ReportMetrics(ctx, api.AgentSystemMetrics{
			MeasuredAt:  time.Now(), // TODO really needed?
			CPUCoresM:   cores * 1000,
			CPUUsage:    cpuUsed,
			MemoryBytes: memoryTotal,
			MemoryUsage: memoryUsed,
		}); err != nil {
			logger.Error("failed to report metrics", zap.Error(err))
			return err
		}

		return nil
	})

	metrics, err := factory.CreateMetrics(ctx, receiver.Settings{
		ID: component.NewID(factory.Type()),
		TelemetrySettings: component.TelemetrySettings{
			MeterProvider:  otel.GetMeterProvider(),
			TracerProvider: otel.GetTracerProvider(),
			Logger:         logger,
		},
	}, cfg, consmr)
	if err != nil {
		log.Fatal(err)
	}

	err = metrics.Start(ctx, &noopHost{})
	if err != nil {
		log.Fatal(err)
	}
}
