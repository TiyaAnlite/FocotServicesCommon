package tracex

import (
	"context"
	"fmt"
	"github.com/TiyaAnlite/FocotServicesCommon/envx"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"
)

func MakeTraceOptions(opts ...uptrace.Option) []uptrace.Option {
	return opts
}

func CheckTraceEnabled() bool {
	e := struct {
		TraceEnabled bool `json:"trace_enabled" yaml:"traceEnabled" env:"TRACE_ENABLED,required" envDefault:"false" validate:"required"`
	}{}
	envx.MustLoadEnv(&e)
	if e.TraceEnabled {
		return true
	} else {
		return false
	}
}

type ServiceTraceHelper struct {
	Scheme         string `json:"scheme" yaml:"scheme" env:"TRACE_SCHEME,required" envDefault:"http" validate:"required"`
	Address        string `json:"address" yaml:"address" env:"TRACE_ADDRESS,required" validate:"required"`
	Port           int    `json:"port" yaml:"port" env:"TRACE_PORT,required" envDefault:"14317" validate:"required"`
	Key            string `json:"key" yaml:"key" env:"TRACE_KEY,required" validate:"required"`
	ProjectId      int    `json:"project_id" yaml:"projectId" env:"TRACE_PROJECT_ID,required" validate:"required"`
	ServiceName    string `json:"service_name" yaml:"serviceName" env:"TRACE_SERVICE_NAME,required" validate:"required"`
	ServiceVersion string `json:"service_version" yaml:"serviceVersion" env:"TRACE_SERVICE_VERSION,required" validate:"required"`
	Environment    string `json:"environment" yaml:"environment" env:"TRACE_ENVIRONMENT,required" validate:"required"`
	PkgName        string `json:"pkg_name" yaml:"pkgName" env:"TRACE_PKG_NAME,required" validate:"required"`
	HostName       string `json:"host_name" yaml:"hostName" env:"TRACE_HOST_NAME"`
}

func (helper *ServiceTraceHelper) SetupTrace() {
	if !CheckTraceEnabled() {
		return
	}
	envx.MustLoadEnv(helper)
	opts := MakeTraceOptions(
		uptrace.WithDSN(fmt.Sprintf("%s://%s@%s:%d/%d", helper.Scheme, helper.Key, helper.Address, helper.Port, helper.ProjectId)),
		uptrace.WithServiceName(helper.ServiceName),
		uptrace.WithServiceVersion(helper.ServiceVersion),
		uptrace.WithDeploymentEnvironment(helper.Environment),
	)
	if helper.HostName != "" {
		opts = append(opts, uptrace.WithResourceAttributes(attribute.String("host.name", helper.HostName)))
	}
	klog.Infof("Uptrace setup: [%s]%s@%s - %s", helper.HostName, helper.ServiceName, helper.ServiceVersion, helper.Environment)
	uptrace.ConfigureOpentelemetry(opts...)
}

func (helper *ServiceTraceHelper) NewTracer() trace.Tracer {
	return otel.Tracer(helper.PkgName)
}

func (*ServiceTraceHelper) Shutdown(ctx context.Context) {
	if !CheckTraceEnabled() {
		return
	}
	klog.Info("Uptrace shutdown")
	_ = uptrace.Shutdown(ctx)
}
