# Getting started


## Installation


### Installation script
```bash
curl -fsSL https://vinceanalytics.com/install.sh | bash
```

### Homebrew
```bash
brew install vinceanalytics/tap/vince
```
### Container image
```bash
docker pull ghcr.io/vinceanalytics/vince
```


### Starting server

```bash
vince --data=vince-data --domains=example.com
```

This will start vince server listening on port `8080`

### Check if your server is up and running

```bash
$ curl http://localhost:8080/api/v1/version
{
  "version": "v0.0.62"
}
```

## AddScript to your website

To integrate your website with Vince Analytics, you need to be able to update the HTML code of the website you want to track. Paste your Vince Analytics tracking script code into the Header (`<head>`) section of your site. Place the tracking script within the `<head> … </head>` tags.

Your Vince Analytics tracking script code will look something like this.

```html
<script defer data-domain="example.com" src="http://localhost:8080/js/vince.js"></script>
```


After adding the script no further configuration on the website is needed. When a user visits the website events will be sent to your Vince Analytics instance.