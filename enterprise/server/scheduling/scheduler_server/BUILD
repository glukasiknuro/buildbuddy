load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "scheduler_server",
    srcs = ["scheduler_server.go"],
    importpath = "github.com/buildbuddy-io/buildbuddy/enterprise/server/scheduling/scheduler_server",
    visibility = ["//visibility:public"],
    deps = [
        "//enterprise/server/tasksize",
        "//proto:api_key_go_proto",
        "//proto:remote_execution_go_proto",
        "//proto:scheduler_go_proto",
        "//proto:trace_go_proto",
        "//server/environment",
        "//server/interfaces",
        "//server/resources",
        "//server/util/background",
        "//server/util/grpc_client",
        "//server/util/log",
        "//server/util/perms",
        "//server/util/status",
        "//server/util/tracing",
        "@com_github_go_redis_redis_v8//:redis",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_prometheus_client_golang//prometheus",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//peer",
    ],
)
