# vince fish shell completion

function __fish_vince_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i login serve k8s init query
            return 1
        end
    end
    return 0
end

complete -c vince -n '__fish_vince_no_subcommand' -f -l help -s h -d 'show help'
complete -c vince -n '__fish_vince_no_subcommand' -f -l version -s v -d 'print the version'
complete -c vince -n '__fish_seen_subcommand_from login' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'login' -d 'Authenticate into vince instance'
complete -c vince -n '__fish_seen_subcommand_from login' -f -l i -s no-i -d 'Shows interactive prompt for username and password'
complete -c vince -n '__fish_seen_subcommand_from login' -f -l username -r -d 'Name of the root user'
complete -c vince -n '__fish_seen_subcommand_from login' -f -l password -r -d 'password of the root user'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'serve' -d 'Serves web ui console and expose /api/events that collects web analytics'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l listen -r -d 'http address to listen to'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l listen-mysql -r -d 'serve mysql clients on this address'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l log-level -r -d 'log level, values are (trace,debug,info,warn,error,fatal,panic)'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l meta-path -r -d 'path to meta data directory'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l blocks-path -r -d 'Path to store block files'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l sync-interval -r -d 'window for buffering timeseries in memory before saving them'
complete -c vince -n '__fish_seen_subcommand_from serve' -f -l enable-profile -d 'Expose /debug/pprof endpoint'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'k8s' -d 'kubernetes controller for vince - The Cloud Native Web Analytics Platform.'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l master-url -r -d 'The address of the Kubernetes API server. Overrides any value in kubeconfig.'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l kubeconfig -r -d 'Path to a kubeconfig. Only required if out-of-cluster.'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l default-image -r -d 'Default image of vince to use'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l port -r -d 'controller api port'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l namespace -r -d 'default namespace where resource managed by v8s will be deployed'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l watch-namespaces -r -d 'namespaces to watch for Vince and Site custom resources'
complete -c vince -n '__fish_seen_subcommand_from k8s' -f -l ignore-namespaces -r -d 'namespaces to ignore for Vince and Site custom resources'
complete -c vince -n '__fish_seen_subcommand_from init' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'init' -d 'Initializes a vince project'
complete -c vince -n '__fish_seen_subcommand_from init' -f -l i -s no-i -d 'Shows interactive prompt for username and password'
complete -c vince -n '__fish_seen_subcommand_from init' -f -l username -r -d 'Name of the root user'
complete -c vince -n '__fish_seen_subcommand_from init' -f -l password -r -d 'password of the root user'
complete -c vince -n '__fish_seen_subcommand_from query' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'query' -d 'connect to vince and execute sql query'
