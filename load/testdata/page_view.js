const site = "https://vinceanalytics.com"
for (let index = 0; index < limit; index++) {
    const session = createSession(site);
    session.Send("pageview", "/");
}