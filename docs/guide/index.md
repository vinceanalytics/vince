
# Vince Analytics

## Introduction

`Vince Analytics` is an  open source cloud native web analytics platform. It allows
you to collect website stats while respecting your visitors privacy.


## Why Vince ?

I was looking for a self hosted web analytics solution that offered these features

- Alert notification
- Time on site tracking
- Conversion tracking 
- Multiple site management
- User Interaction Management 
- Campaign Management 
- Report Generation
- Goal Tracking 
- Event Tracking 
- Cloud Native ( Works well with k8s)
- Backup and Restore

I failed to find one which was easy to use/manage and that was specifically designed
for self hosting. Most solution out there are geared towards cloud offering and 
self hosting comes as a bonus.

Vince was designed from grounds up to be a cloud citizen and a painless self hosted
solution for web analytics ,catering for teams of all sizes.

With ease of use as a self hosted solution, vince comes as a single binary with
everything you need for production deployments. There is zero runtime dependency.

## Core tech we use

### sqlite

Battle tested sql database is used for operational data. This scales very well 
with organization of any size.

### Badger 

Aggregate stats are stored in [Badger](https://github.com/dgraph-io/badger) a key
value store that is used in production. This ensures vince to offer faster permanent
storage of your site stats.

### Ristretto

To ensure fast operations we use [ristretto](https://github.com/dgraph-io/ristretto)
a fast embedded cache. This allows vince to be very fast for events ingestion and other
endpoints. This library is used throughout vince for many performance critical 
components.


These are core components we ship with vince. All embedded in th single binary for yoru
delight.

