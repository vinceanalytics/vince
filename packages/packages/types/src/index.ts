



export type Property =
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
    | "bounceRates"
    | "visitDurations"


export type SelectKind =
    | "exact"
    | "re"
    | "glob"


export type Props = {
    [key in Property]: Metrics;
};

export type Metrics = {
    [key in Metric]: Select;
};

export type Select = {
    [key in SelectKind]: string;
};


export interface Query {
    offset?: string;
    window?: string;
    props: Props;
}

export type PropsResult = {
    [key in Property]: MetricsResult;
};


export type MetricsResult = {
    [key in Metric]: Result;
};

export interface Result {
    [key: string]: number[]
}

export interface QueryResult {
    timestamps: number[];
    props: PropsResult;
}

export interface CreateSite {
    domain: string;
    public?: boolean;
    description?: string;
}

export interface SiteObject {
    domain: string;
    public: boolean;
    description: string;
    owner: string;
    createdAt: string;
    updatedAt: string;
}

export interface UpdateSite {
    public?: boolean;
    description?: string;
}

export type Site = One<SiteObject>

export type SiteList = List<SiteObject>

export type Global = One<MetricsObject>

export type MetricResult = One<number>

export interface MetricsObject {
    visitors: number;
    views: number;
    events: number;
    visits: number;
}

export interface One<T> {
    item: T;
    elapsed: string;
}

export interface List<T> {
    items: T;
    elapsed: string;
}

