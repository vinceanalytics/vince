# NAME

v8s - kubernetes controller for vince - The Cloud Native Web Analytics Platform.

# SYNOPSIS

v8s

```
[--default-image]=[value]
[--kubeconfig]=[value]
[--master-url]=[value]
[--port]=[value]
```

**Usage**:

```
v8s [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--default-image**="": Default image of vince to use (default: ghcr.io/vinceanalytics/vince:v0.0.0)

**--kubeconfig**="": Path to a kubeconfig. Only required if out-of-cluster.

**--master-url**="": The address of the Kubernetes API server. Overrides any value in kubeconfig.

**--port**="": controller api port (default: 9000)


# COMMANDS

## version

prints version information
