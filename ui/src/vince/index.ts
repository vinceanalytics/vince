
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport';
import {
    VinceClient, SitesClient, QueryClient, SnippetsClient, ClusterClient, EventsClient,
    ImportClient
} from "./proto";

export * from "./proto"

export type Timings = {
    compiler: number
    count: number
    execute: number
    fetch: number
}


export type ErrorResult = {
    error: string
    position: number
    query: string
}


export const createVinceClient = (source: TokenSource) => {
    return new VinceClient(createTransport(source))
}

export const createSitesClient = (source: TokenSource) => {
    return new SitesClient(createTransport(source))
}

export const createQueryClient = (source: TokenSource) => {
    return new QueryClient(createTransport(source))
}

export const createSnippetsClient = (source: TokenSource) => {
    return new SnippetsClient(createTransport(source))
}

export const createClusterClient = (source: TokenSource) => {
    return new ClusterClient(createTransport(source))
}

export const createEventsClient = (source: TokenSource) => {
    return new EventsClient(createTransport(source))
}

export const createImportClient = (source: TokenSource) => {
    return new ImportClient(createTransport(source))
}

const createTransport = (source: TokenSource) => {
    return new GrpcWebFetchTransport({
        baseUrl: window.location.origin,
        interceptors: [
            {
                interceptUnary(next, method, input, options) {
                    if (!options.meta) {
                        options.meta = {};
                    }
                    options.meta['Authorization'] = source.header();
                    return next(method, input, options);
                },
            },
        ],
    })
}


type Oauth2Config = {
    client_id: string
    client_secret: string
    token_url: string
}

export type Endpoint = {
    auth_url: string
    token_url: string
}

type OauthParams = Record<string, string | number | boolean | undefined>

export type OauthToken = {
    access_token: string
    refresh_token: string
    expires_in: number
    token_type: string
}

export type Token = {
    access_token: string
    refresh_token: string
    expires_at: string
    token_type: string
}

export const parseOauthToken = ({ access_token, refresh_token, expires_in, token_type }: OauthToken): Token => {
    const exp = new Date(Date.now() + (expires_in * 1000))
    return {
        access_token, refresh_token, token_type,
        expires_at: JSON.stringify(exp),
    }
}

export class TokenSource {
    readonly token: Token

    constructor(payload: string) {
        this.token = JSON.parse(payload)
    }

    header() {
        return `${this.token.token_type} ${this.token.access_token}`
    }

}

export const passwordCredentialsToken = (username: string, password: string) => {
    return retrieveToken({
        client_id: username,
        client_secret: password,
        token_url: `${window.location.origin}/token`
    }, {
        "grant_type": "password",
        username, password
    })
}

const retrieveToken = ({ client_id, client_secret, token_url }: Oauth2Config, params: OauthParams) => {
    return newTokenRequest(token_url, client_id, client_secret, params)
}

const newTokenRequest = (
    token_uri: string,
    client_id: string,
    client_secret: string,
    params: OauthParams,
) => {
    params["client_id"] = client_id
    params["client_secret"] = client_secret
    return fetch(token_uri, {
        method: "POST",
        body: encodeParams(params),
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        }
    })
}

const encodeParams = (
    params: Record<string, string | number | boolean | undefined>,
) =>
    Object.keys(params)
        .filter((k) => typeof params[k] !== "undefined")
        .map(
            (k) =>
                `${encodeURIComponent(k)}=${encodeURIComponent(
                    params[k] as string | number | boolean,
                )}`,
        )
        .join("&")
