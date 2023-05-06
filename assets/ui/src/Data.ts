export interface Data {
    all?: Aggregate;
    Event?: EntryMap;
    page?: EntryMap;
    entryPage?: EntryMap;
    exitPage?: EntryMap;
    referrer?: EntryMap;
    utmMedium?: EntryMap;
    utmSource?: EntryMap;
    utmCampaign?: EntryMap;
    utmContent?: EntryMap;
    utmTerm?: EntryMap;
    utmDevice?: EntryMap;
    utmBrowser?: EntryMap;
    utmBrowserVersion?: EntryMap;
    os?: EntryMap;
    osVersion?: EntryMap;
    country?: EntryMap;
    region?: EntryMap;
    city?: EntryMap;
}

export interface EntryMap {
    entries?: Entries;
    sum?: Summary;
    percent?: Summary,
}

export interface Entries {
    [key: string]: AggregateEntry;
}

export interface Summary {
    visitors?: Item[];
    views?: Item[];
    events?: Item[];
    visits?: Item[];
    bounceRate?: Item[];
    visitDuration?: Item[];
    viewsPerVisit?: Item[];
}
export interface Item {
    key: string;
    value: string;
}

export interface AggregateEntry {
    aggregate: Aggregate,
}

export interface Aggregate {
    visitors?: Entry;
    views?: Entry;
    events?: Entry;
    visits?: Entry;
    bounceRate?: Entry;
    visitDuration?: Entry;
    viewsPerVisit?: Entry;
}

export interface Entry {
    sum: number;
    values: number[];
}