

import { Query, Metrics, Select, QueryResult, Metric, Property } from '@vinceanalytics/types'

export function schedule(call: () => void) {
    //@ts-ignore
    __schedule__(call);
}




export type QueryError =
    | "domain not found"



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
