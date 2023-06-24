


export function schedule(call: () => void) {
    //@ts-ignore
    __schedule__(call);
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
    Object.keys(query.props).forEach((key) => {
        o.props[key] = buildMetric(query.props[key as Property])
    })
    return o;
}

function buildMetric(metrics: Metrics) {
    //@ts-ignore
    const o = new __Metrics__();
    Object.keys(metrics).forEach((key) => {
        o[key] = buildSelect(metrics[key as Metric])

    })
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
        o = __sendMail__(buildMail(mail));
    } catch (error) {
        //@ts-ignore
        return error.message as EmailError;
    }
    return o;
}

function buildMail(e: Email) {
    //@ts-ignore
    let m = new __Email__();
    m.to.name = e.to.name;
    m.to.address = e.to.address;
    m.subject = e.subject;
    m.contentType = e.contentType;
    m.msg = e.msg;
    return m;
}
