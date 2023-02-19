package plug

func Browser() Pipeline {
	return Pipeline{
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

func API() Pipeline {
	return Pipeline{
		FetchSession,
		Auth,
	}
}

func InternalStatsAPI() Pipeline {
	return Pipeline{
		FetchSession,
		AuthorizedSiteAccess(),
	}
}

func PublicAPI() Pipeline {
	return Pipeline{}
}
