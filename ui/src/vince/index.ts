export interface Version {
    version: string
}


export class Vince {
    base: string

    constructor() {
        this.base = window.location.origin
        if (process.env.NODE_ENV === 'development') {
            this.base = process.env.REACT_APP_API_URL!
        }
    }

    async version(): Promise<Version> {
        const response = await fetch(this.base + "/api/v1/version")
        const data = await response.json()
        return data as Version
    }
}