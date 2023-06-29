import { Query, QueryResult } from '@vinceanalytics/types';
export declare function schedule(call: () => void): void;
export type QueryError = "domain not found";
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
export type EmailError = "Mailer not configured" | "Email creation failed" | "Email sending failed";
export declare function query(domain: string, request: Query): QueryResult | QueryError;
export declare function sendMail(mail: Email): number | EmailError;
