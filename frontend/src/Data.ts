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