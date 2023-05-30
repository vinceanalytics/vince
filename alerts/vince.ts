

// Register alert function exec for domain. exec will be called after each
// interval.
//
// Note that exec can't exceed one minute, it will be automatically cancelled. Alert
// functions are only for logic actual processing happens async on wrapper around
// VINCE object.
export function register(domain: string, interval: string, exec: () => void): void {
    // @ts-ignore
    VINCE.Register(domain, interval, exec);
}


export type Property =
    | "base"
    | "event"
    | "page"
    | "entryPage"
    | "exitPage"
    | "referrer"
    | "UtmMedium"
    | "UtmSource"
    | "UtmCampaign"
    | "UtmContent"
    | "UtmTerm"
    | "UtmDevice"
    | "UtmBrowser"
    | "browserVersion"
    | "os"
    | "osVersion"
    | "country"
    | "region"
    | "city";

export type Metric =
    | "visitors"
    | "views"
    | "events"
    | "visits"
    | "bounceRate"
    | "visitDuration"
    | "viewsPerVisit";

export interface Query {
    range: Range;
    metrics: Metric[];
    prop: Property;
    match: Match;
}

export interface Range {
    from: Date;
    to: Date;
}

export interface Match {
    text: string;
    isRe: boolean;
}

export interface QueryResult {
    elapsed: string;
    result: Result[];
}

export interface Result {
    metric: Metric;
    values: Values;
}

export interface Values {
    [key: string]: Value[]
}

export interface Value {
    timestamp: number;
    value: number;
}
