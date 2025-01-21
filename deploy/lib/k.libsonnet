// This is required so that 'k' is available as an import by dependencies
//
// See https://tanka.dev/known-issues#evaluating-jsonnet-runtime-error-couldnt-open-import-klibsonnet-no-match-locally-or-in-the-jsonnet-library-paths
import 'github.com/jsonnet-libs/k8s-libsonnet/1.29/main.libsonnet'
