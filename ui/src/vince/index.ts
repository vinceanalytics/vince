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

type ColumnDefinition = Readonly<{ name: string; type: string }>

type Value = string | number | boolean
type RawData = Record<string, Value>

export enum Type {
    DDL = "ddl",
    DQL = "dql",
    ERROR = "error",
}

export type Timings = {
    compiler: number
    count: number
    execute: number
    fetch: number
}

export type Explain = { jitCompiled: boolean }

type DatasetType = Array<boolean | string | number>

type RawDqlResult = {
    columns: ColumnDefinition[]
    count: number
    dataset: DatasetType[]
    ddl: undefined
    error: undefined
    query: string
    timings: Timings
    explain?: Explain
}

type RawDdlResult = {
    ddl: "OK"
}

type RawErrorResult = {
    ddl: undefined
    error: "<error message>"
    position: number
    query: string
}

type DdlResult = {
    query: string
    type: Type.DDL
}

type RawResult = RawDqlResult | RawDdlResult | RawErrorResult

export type ErrorResult = RawErrorResult & {
    type: Type.ERROR
}

export type QueryRawResult =
    | (Omit<RawDqlResult, "ddl"> & { type: Type.DQL })
    | DdlResult
    | ErrorResult

export type QueryResult<T extends Record<string, any>> =
    | {
        columns: ColumnDefinition[]
        count: number
        data: T[]
        timings: Timings
        type: Type.DQL
        explain?: Explain
    }
    | ErrorResult
    | DdlResult

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


export type Column = {
    column: string
    indexed: boolean
    designated: boolean
    indexBlockCapacity: number
    symbolCached: boolean
    symbolCapacity: number
    type: string
}

export type Options = {
    limit?: string
    explain?: boolean
}




export type ErrorShape = Readonly<{
    error: string
}>


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

    static encodeParams = (
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

    async queryRaw(query: string, options?: Options): Promise<QueryRawResult> {
        const controller = new AbortController()
        const payload = {
            ...options,
            count: true,
            src: "con",
            query,
            timings: true,
        }

        this._controllers.push(controller)
        let response: Response
        const start = new Date()

        try {
            response = await fetch(
                `${this._host}/exec?${Client.encodeParams(payload)}`,
                { signal: controller.signal },
            )
        } catch (error) {
            const err = {
                position: -1,
                query,
                type: Type.ERROR,
            }

            const genericErrorPayload = {
                ...err,
                error: "An error occured, please try again",
            }

            if (error instanceof DOMException) {
                // eslint-disable-next-line prefer-promise-reject-errors
                return await Promise.reject({
                    ...err,
                    error:
                        error.code === 20
                            ? "Cancelled by user"
                            : JSON.stringify(error).toString(),
                })
            }

            // eslint-disable-next-line prefer-promise-reject-errors
            return await Promise.reject(genericErrorPayload)
        } finally {
            const index = this._controllers.indexOf(controller)

            if (index >= 0) {
                this._controllers.splice(index, 1)
            }
        }

        if (response.ok || response.status === 400) {
            // 400 is only for SQL errors
            const fetchTime = (new Date().getTime() - start.getTime()) * 1e6
            const data = (await response.json()) as RawResult


            if (data.ddl) {
                return {
                    query,
                    type: Type.DDL,
                }
            }

            if (data.error) {
                // eslint-disable-next-line prefer-promise-reject-errors
                return await Promise.reject({
                    ...data,
                    type: Type.ERROR,
                })
            }

            return {
                ...data,
                timings: {
                    ...data.timings,
                    fetch: fetchTime,
                },
                type: Type.DQL,
            }
        }

        const errorPayload = {
            error: `QuestDB is not reachable [${response.status}]`,
            position: -1,
            query,
            type: Type.ERROR,
        }

        // eslint-disable-next-line prefer-promise-reject-errors
        return await Promise.reject(errorPayload)
    }
}

