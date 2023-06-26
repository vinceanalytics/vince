# vince fish shell completion

function __fish_vince_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i config version
            return 1
        end
    end
    return 0
end

complete -c vince -n '__fish_vince_no_subcommand' -f -l listen -r -d 'http address to listen to'
complete -c vince -n '__fish_vince_no_subcommand' -f -l log-level -r -d 'log level, values are (trace,debug,info,warn,error,fatal,panic)'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-tls -d 'Enables serving https traffic.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l tls-address -r -d 'https address to listen to. You must provide tls-key and tls-cert or configure auto-tls'
complete -c vince -n '__fish_vince_no_subcommand' -f -l tls-key -r -d 'Path to key file used for https'
complete -c vince -n '__fish_vince_no_subcommand' -f -l tls-cert -r -d 'Path to certificate file used for https'
complete -c vince -n '__fish_vince_no_subcommand' -f -l data -r -d 'path to data directory'
complete -c vince -n '__fish_vince_no_subcommand' -f -l url -r -d 'url for the server on which vince is hosted(it shows up on emails)'
complete -c vince -n '__fish_vince_no_subcommand' -f -l uploads-dir -r -d 'Path to store uploaded assets'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-backup -d 'Allows backing up and restoring'
complete -c vince -n '__fish_vince_no_subcommand' -f -l backup-dir -r -d 'directory where backups are stored'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-email -d 'allows sending emails'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-address -r -d 'email address used for the sender of outgoing emails '
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-address-name -r -d 'email address name  used for the sender of outgoing emails '
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-address -r -d 'host:port address of the smtp server used for outgoing emails'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-enable-mailhog -d 'port address of the smtp server used for outgoing emails'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-anonymous-enable -d 'enables anonymous authenticating smtp client'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-anonymous-trace -r -d 'trace value for anonymous smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-enabled -d 'enables PLAIN authentication of smtp client'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-identity -r -d 'identity value for plain smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-username -r -d 'username value for plain smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-plain-password -r -d 'password value for plain smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-username -d 'allows oauth authentication on smtp client'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-token -r -d 'token value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-host -r -d 'host value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l mailer-smtp-oauth-port -r -d 'port value for oauth bearer smtp auth'
complete -c vince -n '__fish_vince_no_subcommand' -f -l cache-refresh-interval -r -d 'window for refreshing sites cache'
complete -c vince -n '__fish_vince_no_subcommand' -f -l ts-buffer-sync-interval -r -d 'window for buffering timeseries in memory before savin them'
complete -c vince -n '__fish_vince_no_subcommand' -f -l gc-interval -r -d 'How often to perform value log garbage collection'
complete -c vince -n '__fish_vince_no_subcommand' -f -l secret -r -d 'path to a file with  ed25519 private key'
complete -c vince -n '__fish_vince_no_subcommand' -f -l secret-age -r -d 'path to file with age.X25519Identity'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-auto-tls -d 'Enables using acme for automatic https.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-domain -r -d 'Domain to use with letsencrypt.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-certs-path -r -d 'Patch where issued certs will be stored'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-ca -r -d 'The endpoint of the directory for the ACME  CA'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-test-ca -r -d 'The endpoint of the directory for the ACME  CA to use to test domain validation'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-email -r -d 'The email address to use when creating or selecting an existing ACME server account'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-account-key-pem -r -d 'The PEM-encoded private key of the ACME account to use'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-agreed -d 'Agree to CA\'s subscriber agreement'
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-external-account-key-id -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-external-account-mac-key -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-disable-http-challenge
complete -c vince -n '__fish_vince_no_subcommand' -f -l acme-issuer-disable-tls-alpn-challenge
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-bootstrap -d 'allows creating a user and api key on startup.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l bootstrap-name -r -d 'User name of the user to bootstrap.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l bootstrap-full-name -r -d 'Full name of the user to bootstrap.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l bootstrap-email -r -d 'Email address of the user to bootstrap.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l bootstrap-password -r -d 'Password of the user to bootstrap.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l bootstrap-key -r -d 'API Key of the user to bootstrap.'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-profile -d 'Expose /debug/pprof endpoint'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-alerts -d 'allows loading and executing alerts'
complete -c vince -n '__fish_vince_no_subcommand' -f -l alerts-source -r -d 'comma separated list of alert files of the form file[name,interval] eg foo.ts[spike,15m]'
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-origin -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-credentials
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-max-age -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-headers -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-expose -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-methods -r
complete -c vince -n '__fish_vince_no_subcommand' -f -l cors-send-preflight-response
complete -c vince -n '__fish_vince_no_subcommand' -f -l super-users -r -d 'a list of user ID with super privilege'
complete -c vince -n '__fish_vince_no_subcommand' -f -l enable-firewall -d 'allow blocking ip address'
complete -c vince -n '__fish_vince_no_subcommand' -f -l firewall-block-list -r -d 'block  ip address from this list'
complete -c vince -n '__fish_vince_no_subcommand' -f -l firewall-allow-list -r -d 'allow  ip address from this list'
complete -c vince -n '__fish_vince_no_subcommand' -f -l help -s h -d 'show help'
complete -c vince -n '__fish_vince_no_subcommand' -f -l version -s v -d 'print the version'
complete -c vince -n '__fish_seen_subcommand_from config' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'config' -d 'generates configurations for vince'
complete -c vince -n '__fish_seen_subcommand_from version' -f -l help -s h -d 'show help'
complete -r -c vince -n '__fish_vince_no_subcommand' -a 'version' -d 'prints version information'
