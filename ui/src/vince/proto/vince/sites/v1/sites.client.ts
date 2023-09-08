// @generated by protobuf-ts 2.9.1 with parameter generate_dependencies
// @generated from protobuf file "vince/sites/v1/sites.proto" (package "v1", syntax proto3)
// tslint:disable
import type { RpcTransport } from "@protobuf-ts/runtime-rpc";
import type { ServiceInfo } from "@protobuf-ts/runtime-rpc";
import { Sites } from "./sites";
import type { DeleteSiteResponse } from "./sites";
import type { DeleteSiteRequest } from "./sites";
import type { ListSitesResponse } from "./sites";
import type { ListSitesRequest } from "./sites";
import type { Site } from "./sites";
import type { GetSiteRequest } from "./sites";
import { stackIntercept } from "@protobuf-ts/runtime-rpc";
import type { CreateSiteResponse } from "./sites";
import type { CreateSiteRequest } from "./sites";
import type { UnaryCall } from "@protobuf-ts/runtime-rpc";
import type { RpcOptions } from "@protobuf-ts/runtime-rpc";
/**
 * @generated from protobuf service v1.Sites
 */
export interface ISitesClient {
    /**
     * @generated from protobuf rpc: CreateSite(v1.CreateSiteRequest) returns (v1.CreateSiteResponse);
     */
    createSite(input: CreateSiteRequest, options?: RpcOptions): UnaryCall<CreateSiteRequest, CreateSiteResponse>;
    /**
     * @generated from protobuf rpc: GetSite(v1.GetSiteRequest) returns (v1.Site);
     */
    getSite(input: GetSiteRequest, options?: RpcOptions): UnaryCall<GetSiteRequest, Site>;
    /**
     * @generated from protobuf rpc: ListSites(v1.ListSitesRequest) returns (v1.ListSitesResponse);
     */
    listSites(input: ListSitesRequest, options?: RpcOptions): UnaryCall<ListSitesRequest, ListSitesResponse>;
    /**
     * @generated from protobuf rpc: DeleteSite(v1.DeleteSiteRequest) returns (v1.DeleteSiteResponse);
     */
    deleteSite(input: DeleteSiteRequest, options?: RpcOptions): UnaryCall<DeleteSiteRequest, DeleteSiteResponse>;
}
/**
 * @generated from protobuf service v1.Sites
 */
export class SitesClient implements ISitesClient, ServiceInfo {
    typeName = Sites.typeName;
    methods = Sites.methods;
    options = Sites.options;
    constructor(private readonly _transport: RpcTransport) {
    }
    /**
     * @generated from protobuf rpc: CreateSite(v1.CreateSiteRequest) returns (v1.CreateSiteResponse);
     */
    createSite(input: CreateSiteRequest, options?: RpcOptions): UnaryCall<CreateSiteRequest, CreateSiteResponse> {
        const method = this.methods[0], opt = this._transport.mergeOptions(options);
        return stackIntercept<CreateSiteRequest, CreateSiteResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: GetSite(v1.GetSiteRequest) returns (v1.Site);
     */
    getSite(input: GetSiteRequest, options?: RpcOptions): UnaryCall<GetSiteRequest, Site> {
        const method = this.methods[1], opt = this._transport.mergeOptions(options);
        return stackIntercept<GetSiteRequest, Site>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: ListSites(v1.ListSitesRequest) returns (v1.ListSitesResponse);
     */
    listSites(input: ListSitesRequest, options?: RpcOptions): UnaryCall<ListSitesRequest, ListSitesResponse> {
        const method = this.methods[2], opt = this._transport.mergeOptions(options);
        return stackIntercept<ListSitesRequest, ListSitesResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: DeleteSite(v1.DeleteSiteRequest) returns (v1.DeleteSiteResponse);
     */
    deleteSite(input: DeleteSiteRequest, options?: RpcOptions): UnaryCall<DeleteSiteRequest, DeleteSiteResponse> {
        const method = this.methods[3], opt = this._transport.mergeOptions(options);
        return stackIntercept<DeleteSiteRequest, DeleteSiteResponse>("unary", this._transport, method, opt, input);
    }
}