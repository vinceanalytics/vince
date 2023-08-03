export class Client {
    private readonly _host: string
    private _controllers: AbortController[] = []

    constructor(host?: string) {
        if (!host) {
            this._host = window.location.origin
        } else {
            this._host = host
        }
    }
}  