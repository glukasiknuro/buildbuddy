load("@npm//@bazel/typescript:index.bzl", "ts_library")

package(default_visibility = ["//visibility:public"])

exports_files(["sidebar.css"])

ts_library(
    name = "sidebar",
    srcs = glob(["*.tsx"]),
    deps = [
        "//app/auth",
        "//app/capabilities",
        "//app/router",
        "//app/service",
        "@npm//@types/react",
        "@npm//lucide-react",
        "@npm//react",
    ],
)
