// @generated by protobuf-ts 2.9.1 with parameter generate_dependencies
// @generated from protobuf file "vince/api/v1/api.proto" (package "v1", syntax proto3)
// tslint:disable
import type { RpcTransport } from "@protobuf-ts/runtime-rpc";
import type { ServiceInfo } from "@protobuf-ts/runtime-rpc";
import { Vince } from "./api";
import type { Build } from "../../config/v1/config";
import type { Empty } from "../../../google/protobuf/empty";
import type { GetClusterResponse } from "./api";
import type { GetClusterRequest } from "./api";
import type { ApplyClusterResponse } from "./api";
import type { ApplyClusterRequest } from "./api";
import type { QueryResponse } from "./api";
import type { QueryRequest } from "./api";
import type { DeleteSiteResponse } from "./api";
import type { DeleteSiteRequest } from "./api";
import type { ListSitesResponse } from "./api";
import type { ListSitesRequest } from "./api";
import type { GetSiteResponse } from "./api";
import type { GetSiteRequest } from "./api";
import type { CreateSiteResponse } from "./api";
import type { CreateSiteRequest } from "./api";
import { stackIntercept } from "@protobuf-ts/runtime-rpc";
import type { CreateTokenResponse } from "./api";
import type { CreateTokenRequest } from "./api";
import type { UnaryCall } from "@protobuf-ts/runtime-rpc";
import type { RpcOptions } from "@protobuf-ts/runtime-rpc";
/**
 * @generated from protobuf service v1.Vince
 */
export interface IVinceClient {
    /**
     * @generated from protobuf rpc: CreateToken(v1.CreateTokenRequest) returns (v1.CreateTokenResponse);
     */
    createToken(input: CreateTokenRequest, options?: RpcOptions): UnaryCall<CreateTokenRequest, CreateTokenResponse>;
    /**
     * @generated from protobuf rpc: CreateSite(v1.CreateSiteRequest) returns (v1.CreateSiteResponse);
     */
    createSite(input: CreateSiteRequest, options?: RpcOptions): UnaryCall<CreateSiteRequest, CreateSiteResponse>;
    /**
     * @generated from protobuf rpc: GetSite(v1.GetSiteRequest) returns (v1.GetSiteResponse);
     */
    getSite(input: GetSiteRequest, options?: RpcOptions): UnaryCall<GetSiteRequest, GetSiteResponse>;
    /**
     * @generated from protobuf rpc: ListSites(v1.ListSitesRequest) returns (v1.ListSitesResponse);
     */
    listSites(input: ListSitesRequest, options?: RpcOptions): UnaryCall<ListSitesRequest, ListSitesResponse>;
    /**
     * @generated from protobuf rpc: DeleteSite(v1.DeleteSiteRequest) returns (v1.DeleteSiteResponse);
     */
    deleteSite(input: DeleteSiteRequest, options?: RpcOptions): UnaryCall<DeleteSiteRequest, DeleteSiteResponse>;
    /**
     * @generated from protobuf rpc: Query(v1.QueryRequest) returns (v1.QueryResponse);
     */
    query(input: QueryRequest, options?: RpcOptions): UnaryCall<QueryRequest, QueryResponse>;
    /**
     * @generated from protobuf rpc: ApplyCluster(v1.ApplyClusterRequest) returns (v1.ApplyClusterResponse);
     */
    applyCluster(input: ApplyClusterRequest, options?: RpcOptions): UnaryCall<ApplyClusterRequest, ApplyClusterResponse>;
    /**
     * @generated from protobuf rpc: GetCluster(v1.GetClusterRequest) returns (v1.GetClusterResponse);
     */
    getCluster(input: GetClusterRequest, options?: RpcOptions): UnaryCall<GetClusterRequest, GetClusterResponse>;
    /**
     * @generated from protobuf rpc: Version(google.protobuf.Empty) returns (v1.Build);
     */
    version(input: Empty, options?: RpcOptions): UnaryCall<Empty, Build>;
}
/**
 * @generated from protobuf service v1.Vince
 */
export class VinceClient implements IVinceClient, ServiceInfo {
    typeName = Vince.typeName;
    methods = Vince.methods;
    options = Vince.options;
    constructor(private readonly _transport: RpcTransport) {
    }
    /**
     * @generated from protobuf rpc: CreateToken(v1.CreateTokenRequest) returns (v1.CreateTokenResponse);
     */
    createToken(input: CreateTokenRequest, options?: RpcOptions): UnaryCall<CreateTokenRequest, CreateTokenResponse> {
        const method = this.methods[0], opt = this._transport.mergeOptions(options);
        return stackIntercept<CreateTokenRequest, CreateTokenResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: CreateSite(v1.CreateSiteRequest) returns (v1.CreateSiteResponse);
     */
    createSite(input: CreateSiteRequest, options?: RpcOptions): UnaryCall<CreateSiteRequest, CreateSiteResponse> {
        const method = this.methods[1], opt = this._transport.mergeOptions(options);
        return stackIntercept<CreateSiteRequest, CreateSiteResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: GetSite(v1.GetSiteRequest) returns (v1.GetSiteResponse);
     */
    getSite(input: GetSiteRequest, options?: RpcOptions): UnaryCall<GetSiteRequest, GetSiteResponse> {
        const method = this.methods[2], opt = this._transport.mergeOptions(options);
        return stackIntercept<GetSiteRequest, GetSiteResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: ListSites(v1.ListSitesRequest) returns (v1.ListSitesResponse);
     */
    listSites(input: ListSitesRequest, options?: RpcOptions): UnaryCall<ListSitesRequest, ListSitesResponse> {
        const method = this.methods[3], opt = this._transport.mergeOptions(options);
        return stackIntercept<ListSitesRequest, ListSitesResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: DeleteSite(v1.DeleteSiteRequest) returns (v1.DeleteSiteResponse);
     */
    deleteSite(input: DeleteSiteRequest, options?: RpcOptions): UnaryCall<DeleteSiteRequest, DeleteSiteResponse> {
        const method = this.methods[4], opt = this._transport.mergeOptions(options);
        return stackIntercept<DeleteSiteRequest, DeleteSiteResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: Query(v1.QueryRequest) returns (v1.QueryResponse);
     */
    query(input: QueryRequest, options?: RpcOptions): UnaryCall<QueryRequest, QueryResponse> {
        const method = this.methods[5], opt = this._transport.mergeOptions(options);
        return stackIntercept<QueryRequest, QueryResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: ApplyCluster(v1.ApplyClusterRequest) returns (v1.ApplyClusterResponse);
     */
    applyCluster(input: ApplyClusterRequest, options?: RpcOptions): UnaryCall<ApplyClusterRequest, ApplyClusterResponse> {
        const method = this.methods[6], opt = this._transport.mergeOptions(options);
        return stackIntercept<ApplyClusterRequest, ApplyClusterResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: GetCluster(v1.GetClusterRequest) returns (v1.GetClusterResponse);
     */
    getCluster(input: GetClusterRequest, options?: RpcOptions): UnaryCall<GetClusterRequest, GetClusterResponse> {
        const method = this.methods[7], opt = this._transport.mergeOptions(options);
        return stackIntercept<GetClusterRequest, GetClusterResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: Version(google.protobuf.Empty) returns (v1.Build);
     */
    version(input: Empty, options?: RpcOptions): UnaryCall<Empty, Build> {
        const method = this.methods[8], opt = this._transport.mergeOptions(options);
        return stackIntercept<Empty, Build>("unary", this._transport, method, opt, input);
    }
}
