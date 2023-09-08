// @generated by protobuf-ts 2.9.1 with parameter generate_dependencies
// @generated from protobuf file "vince/api/v1/api.proto" (package "v1", syntax proto3)
// tslint:disable
import type { RpcTransport } from "@protobuf-ts/runtime-rpc";
import type { ServiceInfo } from "@protobuf-ts/runtime-rpc";
import { Vince } from "./api";
import type { Event } from "./api";
import type { Build } from "../../config/v1/config";
import type { Empty } from "../../../google/protobuf/empty";
import type { GetClusterResponse } from "./api";
import type { GetClusterRequest } from "./api";
import type { ApplyClusterResponse } from "./api";
import type { ApplyClusterRequest } from "./api";
import { stackIntercept } from "@protobuf-ts/runtime-rpc";
import type { LoginResponse } from "./api";
import type { LoginRequest } from "./api";
import type { UnaryCall } from "@protobuf-ts/runtime-rpc";
import type { RpcOptions } from "@protobuf-ts/runtime-rpc";
/**
 * @generated from protobuf service v1.Vince
 */
export interface IVinceClient {
    /**
     * @generated from protobuf rpc: Login(v1.LoginRequest) returns (v1.LoginResponse);
     */
    login(input: LoginRequest, options?: RpcOptions): UnaryCall<LoginRequest, LoginResponse>;
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
    /**
     * @generated from protobuf rpc: SendEvent(v1.Event) returns (google.protobuf.Empty);
     */
    sendEvent(input: Event, options?: RpcOptions): UnaryCall<Event, Empty>;
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
     * @generated from protobuf rpc: Login(v1.LoginRequest) returns (v1.LoginResponse);
     */
    login(input: LoginRequest, options?: RpcOptions): UnaryCall<LoginRequest, LoginResponse> {
        const method = this.methods[0], opt = this._transport.mergeOptions(options);
        return stackIntercept<LoginRequest, LoginResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: ApplyCluster(v1.ApplyClusterRequest) returns (v1.ApplyClusterResponse);
     */
    applyCluster(input: ApplyClusterRequest, options?: RpcOptions): UnaryCall<ApplyClusterRequest, ApplyClusterResponse> {
        const method = this.methods[1], opt = this._transport.mergeOptions(options);
        return stackIntercept<ApplyClusterRequest, ApplyClusterResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: GetCluster(v1.GetClusterRequest) returns (v1.GetClusterResponse);
     */
    getCluster(input: GetClusterRequest, options?: RpcOptions): UnaryCall<GetClusterRequest, GetClusterResponse> {
        const method = this.methods[2], opt = this._transport.mergeOptions(options);
        return stackIntercept<GetClusterRequest, GetClusterResponse>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: Version(google.protobuf.Empty) returns (v1.Build);
     */
    version(input: Empty, options?: RpcOptions): UnaryCall<Empty, Build> {
        const method = this.methods[3], opt = this._transport.mergeOptions(options);
        return stackIntercept<Empty, Build>("unary", this._transport, method, opt, input);
    }
    /**
     * @generated from protobuf rpc: SendEvent(v1.Event) returns (google.protobuf.Empty);
     */
    sendEvent(input: Event, options?: RpcOptions): UnaryCall<Event, Empty> {
        const method = this.methods[4], opt = this._transport.mergeOptions(options);
        return stackIntercept<Event, Empty>("unary", this._transport, method, opt, input);
    }
}