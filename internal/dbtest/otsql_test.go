package dbtest_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	exporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	jaegerCollectorEndpoint = "http://localhost:14268/api/traces"
	serviceName             = "otsql@bun"
	prometheusPort          = "2222"
)

func InitTracer() {
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerCollectorEndpoint)),
	)
	if err != nil {
		panic(err)
	}

	resource, err := resource.New(
		context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		panic(err)
	}

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(exporter),
		oteltrace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
}

func InitMeter() {
	config := exporter.Config{}
	ctl := controller.New(
		processor.New(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			export.CumulativeExportKindSelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(resource.Empty()),
	)
	exp, err := exporter.New(
		exporter.Config{Registry: prometheus.DefaultRegisterer.(*prometheus.Registry)},
		ctl,
	)
	if err != nil {
		panic(err)
	}
	global.SetMeterProvider(exp.MeterProvider())

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", prometheusPort), exp)
		if err != nil {
			panic(err)
		}
	}()
}
