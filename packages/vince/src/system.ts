

// Schedules function call for execution every interval. interval is a duration
// string.
//
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
export function schedule(interval: string, call: () => void) {
    //@ts-ignore
    __schedule__(interval, call);
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
    | "bounceRates"
    | "visitDurations"


export type SelectKind =
    | "exact"
    | "re"
    | "glob"

export type QueryError =
    | "domain not found"


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

export interface Email {
    from: Address;
    to: Address;
    subject: string;
    contentType: string;
    msg: string;
}

export interface Address {
    name: string;
    address: string;
}

export type EmailError =
    | "Mailer not configured"
    | "Email creation failed"
    | "Email sending failed"

function build(query: Query) {
    //@ts-ignore
    const o = new __Query__();
    //@ts-ignore
    if (query.offset) o.offset = new __Duration__(query.offset);
    //@ts-ignore
    if (query.window) o.window = new __Duration__(query.window);
    //@ts-ignore
    o.props = new __Props__();
    Object.entries(query.props).forEach(([key, value]) => {
        o.props[key] = buildMetric(value)
    });
    return o;
}

function buildMetric(metrics: Metrics) {
    //@ts-ignore
    const o = new __Metrics__();
    Object.entries(metrics).forEach(([key, value]) => {
        o[key] = buildSelect(value)
    });
    return o
}

function buildSelect(select: Select) {
    for (let key in select) {
        switch (key) {
            case "exact":
                //@ts-ignore
                return new __SelectExact__(select[key]);
            case "glob":
                //@ts-ignore
                return new __SelectGlob__(select[key]);
            case "re":
                //@ts-ignore
                return new __SelectRe__(select[key]);
            default:
                break;
        }
    }
}

export function query(domain: string, request: Query): QueryResult | QueryError {
    let o: QueryResult;
    try {
        //@ts-ignore
        o = __query__(domain, build(request));
    } catch (error) {
        //@ts-ignore
        return error.message as QueryError;
    }
    return o as QueryResult;
}


export function sendMail(mail: Email): number | EmailError {
    let o: number;
    try {
        //@ts-ignore
        o = __sendMail__(buildMai(mail));
    } catch (error) {
        //@ts-ignore
        return error.message as EmailError;
    }
    return o;
}

function buildMail(e: Email) {
    //@ts-ignore
    let m = new __Email__();
    //@ts-ignore
    m.from = new __Address__(e.from.name, e.from.address);
    //@ts-ignore
    m.to = new __Address__(e.to.name, e.to.address);
    m.subject = e.subject;
    m.contentType = e.contentType;
    m.msg = e.msg;
    return m;
}
