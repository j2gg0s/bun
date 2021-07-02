package dbtest_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	exporter "go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"

	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
)

var (
	jaegerCollectorEndpoint = "http://localhost:14268/api/traces"
	serviceName             = "otsql@bun"
	prometheusPort          = "2222"
)

func InitTracer() {
	exporter, err := jaeger.NewRawExporter(
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
	exporter, err := exporter.InstallNewPipeline(
		exporter.Config{
			Registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
		},
		controller.WithResource(resource.Empty()))
	if err != nil {
		panic(err)
	}

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", prometheusPort), exporter)
		if err != nil {
			panic(err)
		}
	}()
}
