load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["retry.go"],
    importpath = "aqwari.net/retry",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["retry_test.go"],
    embed = [":go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    deps = [":go_default_library"],
)
