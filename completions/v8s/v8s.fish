# v8s fish shell completion

function __fish_v8s_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i version
            return 1
        end
    end
    return 0
end

complete -c v8s -n '__fish_v8s_no_subcommand' -f -l master-url -r -d 'The address of the Kubernetes API server. Overrides any value in kubeconfig.'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l kubeconfig -r -d 'Path to a kubeconfig. Only required if out-of-cluster.'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l default-image -r -d 'Default image of vince to use'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l port -r -d 'controller api port'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l namespace -r -d 'default namespace where resource managed by v8s will be deployed'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l watch-namespaces -r -d 'namespaces to watch for Vince and Site custom resources'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l ignore-namespaces -r -d 'namespaces to ignore for Vince and Site custom resources'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l help -s h -d 'show help'
complete -c v8s -n '__fish_v8s_no_subcommand' -f -l version -s v -d 'print the version'
complete -c v8s -n '__fish_seen_subcommand_from version' -f -l help -s h -d 'show help'
complete -r -c v8s -n '__fish_v8s_no_subcommand' -a 'version' -d 'prints version information'
