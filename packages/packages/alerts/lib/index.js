"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.sendMail = exports.query = exports.schedule = void 0;
function schedule(call) {
    //@ts-ignore
    __schedule__(call);
}
exports.schedule = schedule;
function build(query) {
    //@ts-ignore
    const o = new __Query__();
    //@ts-ignore
    if (query.offset)
        o.offset = new __Duration__(query.offset);
    //@ts-ignore
    if (query.window)
        o.window = new __Duration__(query.window);
    //@ts-ignore
    o.props = new __Props__();
    Object.keys(query.props).forEach((key) => {
        o.props[key] = buildMetric(query.props[key]);
    });
    return o;
}
function buildMetric(metrics) {
    //@ts-ignore
    const o = new __Metrics__();
    Object.keys(metrics).forEach((key) => {
        o[key] = buildSelect(metrics[key]);
    });
    return o;
}
function buildSelect(select) {
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
function query(domain, request) {
    let o;
    try {
        //@ts-ignore
        o = __query__(domain, build(request));
    }
    catch (error) {
        //@ts-ignore
        return error.message;
    }
    return o;
}
exports.query = query;
function sendMail(mail) {
    let o;
    try {
        //@ts-ignore
        o = __sendMail__(buildMail(mail));
    }
    catch (error) {
        //@ts-ignore
        return error.message;
    }
    return o;
}
exports.sendMail = sendMail;
function buildMail(e) {
    //@ts-ignore
    let m = new __Email__();
    m.to.name = e.to.name;
    m.to.address = e.to.address;
    m.subject = e.subject;
    m.contentType = e.contentType;
    m.msg = e.msg;
    return m;
}
