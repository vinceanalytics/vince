# vince fish shell completion

function __fish_vince_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i config version
            return 1
        end
    end
    return 0
end

complete -c vince -n '__fish_vince_no_subcommand' -f -l config -r -d 'configuration file in json format'
complete -c vince -n '__fish_vince_no_subcommand' -f -l listen-address -r -d 'bind the server to this port'
complete -c vince -n '__fish_vince_no_subcommand' -f -l data -r -d 'path to data directory'
complete -c vince -n '__fish_vince_no_subcommand' -f -l env -r -d 'environment on which vince is run (dev,staging,production)'
complete -c vince -n '__fish_vince_no_subcommand' -f -l url -r -d 'url for the server on which vince is hosted(it shows up on emails)'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-email-verification -d 'send emails for account verification'
complete -c vince -n '__fish_vince_no_subcommand' -f -l self-host -d 'self hosted version'
complete -c vince -n '__fish_vince_no_subcommand' -f -l log -r -d 'level of logging'
complete -c vince -n '__fish_vince_no_subcommand' -f -l backup-dir -r -d 'directory where backups are stored'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-address -r -d 'email address used for the sender of outgoing emails '
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-address-name -r -d 'email address name  used for the sender of outgoing emails '
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-host -r -d 'host address of the smtp server used for outgoing emails'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-port -r -d 'port address of the smtp server used for outgoing emails'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-anonymous -r -d 'trace value for anonymous smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-identity -r -d 'identity value for plain smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-username -r -d 'username value for plain smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-password -r -d 'password value for plain smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-username -r -d 'username value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-token -r -d 'token value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-host -r -d 'host value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-port -r -d 'port value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l cache-refresh -r -d 'window for refreshing sites cache'
complete -c vince -n '__fish_vince_no_subcommand' -f -l rotation-check -r -d 'window for checking log rotation'
complete -c vince -n '__fish_vince_no_subcommand' -f -l ts-buffer -r -d 'window for buffering timeseries in memory before savin them'
complete -c vince -n '__fish_vince_no_subcommand' -f -l scrape-interval -r -d 'system wide metrics collection interval'
complete -c vince -n '__fish_vince_no_subcommand' -f -l secret-ed-priv -r -d 'path to a file with  ed25519 private key'
complete -c vince -n '__fish_vince_no_subcommand' -f -l secret-ed-pub -r -d 'path to a file with  ed25519 public key'
complete -c vince -n '__fish_vince_no_subcommand' -f -l secret-age-pub -r -d 'path to a file with  age public key'
complete -c vince -n '__fish_vince_no_subcommand' -f -l secret-age-priv -r -d 'path to a file with  age private key'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-system-stats -d 'Collect and visualize system stats'
complete -c vince -n '__fish_vince_no_subcommand' -f -l help -s h -d 'show help'
complete -c vince -n '__fish_vince_no_subcommand' -f -l version -s v -d 'print the version'
complete -c vince -n '__fish_seen_subcommand_from config' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'config' -d 'generates configurations for vince'
complete -c vince -n '__fish_seen_subcommand_from config' -f -l path -r -d 'directory to save configurations (including secrets)'
complete -c vince -n '__fish_seen_subcommand_from version' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'version' -d 'prints version information'
