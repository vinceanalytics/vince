/*******************************************************************************
 *     ___                  _   ____  ____
 *    / _ \ _   _  ___  ___| |_|  _ \| __ )
 *   | | | | | | |/ _ \/ __| __| | | |  _ \
 *   | |_| | |_| |  __/\__ \ |_| |_| | |_) |
 *    \__\_\\__,_|\___||___/\__|____/|____/
 *
 *  Copyright (c) 2014-2019 Appsicle
 *  Copyright (c) 2019-2022 QuestDB
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 ******************************************************************************/


export type QueryRequestOptions = {
    query: string;
    params?: Param[];
}

export type Timings = {
    compiler: number
    count: number
    execute: number
    fetch: number
}

export type QueryResult = {
    elapsed: string;
    columns: Column[];
    rows: Row[];
}

export type Column = {
    name: string;
    dataType: DataType;
}

export type Row = {
    values: Value[];
}

export type DataType = "UNKNOWN" | "NUMBER" | "DOUBLE" | "STRING" | "BOOL" | "TIMESTAMP";

export type Value = {
    number?: number;
    double?: number;
    string?: string;
    bool?: boolean;
    timestamp?: string;
}

export type Param = {
    name: string;
    value: Value;
}

export type Site = {
    domain: string
}

export type SiteList = {
    list: Site[];
}

export type TokenRequest = {
    name: string;
    password: string;
    generate: true;
}

export type TokenResult = {
    token: string;
}


export type ErrorResult = ErrorShape & {
    position: number
    query: string
}

export type ErrorShape = {
    error: string
}


export type Version = {
    version: string
}



export class Client {
    private readonly _host: string
    private readonly _token: string
    private _controllers: AbortController[] = []

    constructor(token?: string) {
        this._host = window.location.origin
        if (token) {
            this._token = token
        } else {
            this._token = ""
        }
    }

    authenticated = (): boolean => {
        return this._token !== ""
    }

    abort = () => {
        this._controllers.forEach((controller) => {
            controller.abort()
        })
        this._controllers = []
    }


    async login(request: TokenRequest): Promise<TokenResult | ErrorShape> {
        return this.do<TokenResult>("/tokens", {
            method: "POST",
            body: JSON.stringify(request),
        })
    }

    async sites(): Promise<SiteList | ErrorShape> {
        return this.do<SiteList>("/sites")
    }

    async version(): Promise<Version | ErrorShape> {
        return this.do<Version>("/version")
    }

    async create(domain: string): Promise<Site | ErrorShape> {
        return this.do<Site>("/sites", {
            method: "POST",
            body: JSON.stringify({ domain })
        })
    }

    async query(request: QueryRequestOptions): Promise<QueryResult | ErrorShape> {
        return this.do<QueryResult>("/query", {
            method: "POST",
            body: JSON.stringify(request)
        })
    }


    async do<T extends Record<string, any>>(
        uri: string | URL, init?: RequestInit | undefined,
    ): Promise<T | ErrorShape> {
        const controller = new AbortController()
        this._controllers.push(controller)
        const path = `${this._host}${uri}`
        let response: Response
        try {
            response = await fetch(path, {
                ...init,
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                    'authorization': 'Bearer ' + this._token
                },
            })
        } catch (error) {
            return await Promise.reject({
                error: JSON.stringify(error).toString()
            })
        } finally {
            const index = this._controllers.indexOf(controller)
            if (index >= 0) {
                this._controllers.splice(index, 1)
            }
        }
        if (response.ok) {
            if (
                !response.headers.get("content-type")?.startsWith("application/json")
            ) {
                return await Promise.reject({ error: "unexpected content type" })
            }
            const data = (await response.json()) as T | ErrorShape
            if (data.error) {
                return await Promise.reject(data)
            }
            return data as T
        }
        return await Promise.reject({
            error: response.statusText,
        })
    }
}

