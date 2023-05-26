# vince

The Cloud Native Web Analytics Platform.


# Features

- Alert notification (Write Alerting Scripts in Typescript)
- Time on site tracking
- Conversion tracking 
- Multiple site management
- User Interaction Management 
- Campaign Management 
- Report Generation
- Goal Tracking 
- Event Tracking 
- Cloud Native (seamless k8s integration)
- Automatic TLS
- API for stats and sites management
- No runtime dependency (Static binary with everything you need)

## Usage

[DOcumentation]()

# Origins

This started as a go port of [Plausible](https://github.com/plausible/analytics), with 
the intention to remove clickhouse and postgresql dependency . I wanted a simpler
deployment model and focusing on a single organization managing its own sites.

# The name vince 

Vince is named after [Vince Staples](https://en.wikipedia.org/wiki/Vince_Staples) , 
my favorite hip hop artist.

It was late night, 1 year after I quit my job and took a sabbatical, I was listening
to one of his album [Big Fish Theory](https://en.wikipedia.org/wiki/Big_Fish_Theory)
. The song Big Fish inspired me to program again.

The lyrics that got me going.
```
I was up late night ballin'
Countin' up hundreds by the thousand
```

So, enjoy the outcome of my late nights, hoping to be counting hundreds by the thousand
soon.

# FAQ

## WHy is the binary so big ?

Assets are bundled with the binary. Also we ship esbuild so you can write alerts
in typescript. Painless deployment and convenient features don't come cheap.

## Why PR , Issues and Discussions are disabled

This is requiem for my defunct web analytics startup dream. If you are reading this,
[tigerbeetledb/tigerbeetle](https://github.com/tigerbeetledb/tigerbeetle) team, I am
more than happy to support your `vince` deployment if you ever need to use it.

I lack the technical chops to be part of your team, but I will always root for your success.
Your code was a therapy to me, with all my heart, Thank you.
