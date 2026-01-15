package main

import (
	"context"

	"github.com/distr-sh/distr/api"
	hmr "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var metrics receiver.Metrics

const hostMetricsReceiverConfig = `
collection_interval: 30s
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
`

type defaultHost struct{}

func (nh *defaultHost) GetExtensions() map[component.ID]component.Component {
	return nil
}

func startMetrics(ctx context.Context) {
	if metrics != nil {
		return
	}
	logger.Info("starting metrics")

	factory := hmr.NewFactory()
	cfg := factory.CreateDefaultConfig().(*hmr.Config)

	if retrieved, err := confmap.NewRetrievedFromYAML([]byte(hostMetricsReceiverConfig)); err != nil {
		logger.Error("failed to create yaml metrics config", zap.Error(err))
		return
	} else if cnf, err := retrieved.AsConf(); err != nil {
		logger.Error("failed to parse metrics config", zap.Error(err))
		return
	} else if err := cfg.Unmarshal(cnf); err != nil {
		logger.Error("failed to apply metrics config", zap.Error(err))
		return
	}

	consmr, err := consumer.NewMetrics(func(ctx context.Context, md pmetric.Metrics) error {
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
									cpuUsed += dataPoint.DoubleValue()
								}
							}
						}

					case "system.memory.utilization":
						for _, dataPoint := range metric.Gauge().DataPoints().All() {
							if state, ok := dataPoint.Attributes().Get("state"); ok {
								if state.Str() == "used" {
									// TODO maybe make calculation more specific with the other possible states
									memoryUsed += dataPoint.DoubleValue()
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
		var usage float64
		if cores != 0 {
			usage = cpuUsed / float64(cores)
		}
		logger.Debug("cpu usage", zap.Any("usage", usage), zap.Any("cores", cores))
		logger.Debug("memory usage", zap.Any("usage", memoryUsed), zap.Any("total", memoryTotal))

		if err := client.ReportMetrics(ctx, api.AgentDeploymentTargetMetrics{
			CPUCoresMillis: cores * 1000,
			CPUUsage:       usage,
			MemoryBytes:    memoryTotal,
			MemoryUsage:    memoryUsed,
		}); err != nil {
			logger.Error("failed to report metrics", zap.Error(err))
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error("failed to create metrics consumer", zap.Error(err))
	}

	metrics, err = factory.CreateMetrics(ctx, receiver.Settings{
		ID: component.NewID(factory.Type()),
		TelemetrySettings: component.TelemetrySettings{
			MeterProvider:  otel.GetMeterProvider(),
			TracerProvider: otel.GetTracerProvider(),
			Logger:         logger,
		},
	}, cfg, consmr)
	if err != nil {
		logger.Error("failed to create metrics", zap.Error(err))
	}

	err = metrics.Start(ctx, &defaultHost{})
	if err != nil {
		logger.Error("failed to start metrics", zap.Error(err))
	}
}

func stopMetrics(ctx context.Context) {
	if metrics != nil {
		logger.Info("stopping metrics")
		if err := metrics.Shutdown(ctx); err != nil {
			logger.Error("failed to stop metrics", zap.Error(err))
		}
		metrics = nil
	}
}
