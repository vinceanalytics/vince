
<p align="center">
    <img src="./assets/ui/logo.svg" alt="Vince Logo" />
</p>

> :warning: **This is still under early development its not in a usable state yet**. 

# vince

The Cloud Native Web Analytics Platform. Built on Apache Arrow and Apache Parquet.

# Features

- [x] SQL for querying stats (All MySQL compatible clients are supported)
- [x] Time on site tracking
- [ ] Conversion tracking 
- [x] Multiple site management
- [ ] User Interaction Management 
- [ ] Campaign Management 
- [ ] Report Generation
- [ ] Goal Tracking 
- [x] Event Tracking 
- [ ] Cloud Native (seamless k8s integration)
- [ ] API for sites management
- [x] No runtime dependency (Static binary with everything you need)

## Usage

<details markdown="1">
<summary>Initialize a project</summary>

```bash
export VINCE_ROOT_USER=root
export VINCE_ROOT_PASSWORD=xxxxx
vince init example
```
</details>

## Contributing

[Contributing](https://vinceanalytics.github.io/contibuting)
