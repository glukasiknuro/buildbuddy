load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "buildbuddy_enterprise",
    testonly = 1,
    srcs = ["buildbuddy_enterprise.go"],
    data = [
        "//enterprise:config_files",
        "//enterprise/server/cmd/server:buildbuddy",
    ],
    importpath = "github.com/buildbuddy-io/buildbuddy/enterprise/server/testutil/buildbuddy_enterprise",
    visibility = ["//visibility:public"],
    deps = [
        "//enterprise/server/testutil/testredis",
        "//proto:context_go_proto",
        "//proto:user_go_proto",
        "//server/testutil/app",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_stretchr_testify//require",
    ],
)
