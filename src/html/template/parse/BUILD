licenses(["notice"])  # BSD-style

go_library(
    name = "parse",
    srcs = [
        "context.go",
        "delim_string.go",
        "error.go",
        "jsctx_string.go",
        "state_string.go",
        "urlpart_string.go",
    ],
    build_for_appengine = 1,
    visibility = [
        "//third_party/golang/safehtml/template:__pkg__",
        "//third_party/golang/template/html/template:__pkg__",
    ],
)
