package plug

import "context"

func Browser(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
		FetchSession,
		PutSecureBrowserHeaders,
		SessionTimeout,
		Auth,
		LastSeen,
	}
}

func SharedLink() Pipeline {
	return Pipeline{
		PutSecureBrowserHeaders,
	}
}

func Protect() Pipeline {
	return Pipeline{
		CSRF,
		Captcha,
	}
}

func API(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
		FetchSession,
		Auth,
	}
}

func InternalStatsAPI(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
		FetchSession,
		AuthorizedSiteAccess(),
	}
}

func PublicAPI(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
	}
}
