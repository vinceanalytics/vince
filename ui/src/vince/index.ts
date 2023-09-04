
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport';
import {
    VinceClient
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

export type CreateOptions = {
    token?: string
    basic?: Basic
}

export type Basic = {
    username: string
    password: string
}

export const createClient = ({ token, basic }: CreateOptions) => {
    let auth = ''
    if (token) {
        auth = `Bearer ${token}`
    }
    if (basic) {
        const base = btoa(basic.username + ":" + basic.password)
        auth = `Basic ${base}`
    }
    return new VinceClient(new GrpcWebFetchTransport({
        baseUrl: window.location.origin,
        interceptors: [
            {
                interceptUnary(next, method, input, options) {
                    if (!options.meta) {
                        options.meta = {};
                    }
                    options.meta['Authorization'] = auth;
                    return next(method, input, options);
                },
            },
        ],
    }))
}

export const login = (username: string, password: string) => {
    const client = createClient({
        basic: {
            username, password,
        }
    })
    return client.login({
        token: '',
        publicKey: new Uint8Array(32),
        generate: true,
    })
}



