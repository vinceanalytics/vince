// @generated by protobuf-ts 2.9.1 with parameter generate_dependencies
// @generated from protobuf file "vince/snippets/v1/snippets.proto" (package "v1", syntax proto3)
// tslint:disable
import { Empty } from "../../../google/protobuf/empty";
import { ServiceType } from "@protobuf-ts/runtime-rpc";
import type { BinaryWriteOptions } from "@protobuf-ts/runtime";
import type { IBinaryWriter } from "@protobuf-ts/runtime";
import { WireType } from "@protobuf-ts/runtime";
import type { BinaryReadOptions } from "@protobuf-ts/runtime";
import type { IBinaryReader } from "@protobuf-ts/runtime";
import { UnknownFieldHandler } from "@protobuf-ts/runtime";
import type { PartialMessage } from "@protobuf-ts/runtime";
import { reflectionMergePartial } from "@protobuf-ts/runtime";
import { MESSAGE_TYPE } from "@protobuf-ts/runtime";
import { MessageType } from "@protobuf-ts/runtime";
import { Timestamp } from "../../../google/protobuf/timestamp";
import { QueryParam } from "../../query/v1/query";
/**
 * @generated from protobuf message v1.Snippet
 */
export interface Snippet {
    /**
     * @generated from protobuf field: string id = 1;
     */
    id: string;
    /**
     * @generated from protobuf field: string name = 2;
     */
    name: string;
    /**
     * @generated from protobuf field: string query = 3;
     */
    query: string;
    /**
     * @generated from protobuf field: repeated v1.QueryParam params = 4;
     */
    params: QueryParam[];
    /**
     * @generated from protobuf field: google.protobuf.Timestamp created_at = 5;
     */
    createdAt?: Timestamp;
    /**
     * @generated from protobuf field: google.protobuf.Timestamp updated_at = 6;
     */
    updatedAt?: Timestamp;
}
/**
 * @generated from protobuf message v1.CreateSnippetRequest
 */
export interface CreateSnippetRequest {
    /**
     * @generated from protobuf field: string name = 1;
     */
    name: string;
    /**
     * @generated from protobuf field: string query = 2;
     */
    query: string;
    /**
     * @generated from protobuf field: repeated v1.QueryParam params = 3;
     */
    params: QueryParam[];
}
/**
 * @generated from protobuf message v1.UpdateSnippetRequest
 */
export interface UpdateSnippetRequest {
    /**
     * @generated from protobuf field: string id = 1;
     */
    id: string;
    /**
     * @generated from protobuf field: string name = 2;
     */
    name: string;
    /**
     * @generated from protobuf field: string query = 3;
     */
    query: string;
    /**
     * @generated from protobuf field: repeated v1.QueryParam params = 4;
     */
    params: QueryParam[];
}
/**
 * @generated from protobuf message v1.DeleteSnippetRequest
 */
export interface DeleteSnippetRequest {
    /**
     * @generated from protobuf field: string id = 1;
     */
    id: string;
}
/**
 * @generated from protobuf message v1.ListSnippetsRequest
 */
export interface ListSnippetsRequest {
}
/**
 * @generated from protobuf message v1.ListSnippetsResponse
 */
export interface ListSnippetsResponse {
    /**
     * @generated from protobuf field: repeated v1.Snippet snippets = 1;
     */
    snippets: Snippet[];
}
// @generated message type with reflection information, may provide speed optimized methods
class Snippet$Type extends MessageType<Snippet> {
    constructor() {
        super("v1.Snippet", [
            { no: 1, name: "id", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 2, name: "name", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 3, name: "query", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 4, name: "params", kind: "message", repeat: 1 /*RepeatType.PACKED*/, T: () => QueryParam },
            { no: 5, name: "created_at", kind: "message", T: () => Timestamp },
            { no: 6, name: "updated_at", kind: "message", T: () => Timestamp }
        ]);
    }
    create(value?: PartialMessage<Snippet>): Snippet {
        const message = { id: "", name: "", query: "", params: [] };
        globalThis.Object.defineProperty(message, MESSAGE_TYPE, { enumerable: false, value: this });
        if (value !== undefined)
            reflectionMergePartial<Snippet>(this, message, value);
        return message;
    }
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: Snippet): Snippet {
        let message = target ?? this.create(), end = reader.pos + length;
        while (reader.pos < end) {
            let [fieldNo, wireType] = reader.tag();
            switch (fieldNo) {
                case /* string id */ 1:
                    message.id = reader.string();
                    break;
                case /* string name */ 2:
                    message.name = reader.string();
                    break;
                case /* string query */ 3:
                    message.query = reader.string();
                    break;
                case /* repeated v1.QueryParam params */ 4:
                    message.params.push(QueryParam.internalBinaryRead(reader, reader.uint32(), options));
                    break;
                case /* google.protobuf.Timestamp created_at */ 5:
                    message.createdAt = Timestamp.internalBinaryRead(reader, reader.uint32(), options, message.createdAt);
                    break;
                case /* google.protobuf.Timestamp updated_at */ 6:
                    message.updatedAt = Timestamp.internalBinaryRead(reader, reader.uint32(), options, message.updatedAt);
                    break;
                default:
                    let u = options.readUnknownField;
                    if (u === "throw")
                        throw new globalThis.Error(`Unknown field ${fieldNo} (wire type ${wireType}) for ${this.typeName}`);
                    let d = reader.skip(wireType);
                    if (u !== false)
                        (u === true ? UnknownFieldHandler.onRead : u)(this.typeName, message, fieldNo, wireType, d);
            }
        }
        return message;
    }
    internalBinaryWrite(message: Snippet, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter {
        /* string id = 1; */
        if (message.id !== "")
            writer.tag(1, WireType.LengthDelimited).string(message.id);
        /* string name = 2; */
        if (message.name !== "")
            writer.tag(2, WireType.LengthDelimited).string(message.name);
        /* string query = 3; */
        if (message.query !== "")
            writer.tag(3, WireType.LengthDelimited).string(message.query);
        /* repeated v1.QueryParam params = 4; */
        for (let i = 0; i < message.params.length; i++)
            QueryParam.internalBinaryWrite(message.params[i], writer.tag(4, WireType.LengthDelimited).fork(), options).join();
        /* google.protobuf.Timestamp created_at = 5; */
        if (message.createdAt)
            Timestamp.internalBinaryWrite(message.createdAt, writer.tag(5, WireType.LengthDelimited).fork(), options).join();
        /* google.protobuf.Timestamp updated_at = 6; */
        if (message.updatedAt)
            Timestamp.internalBinaryWrite(message.updatedAt, writer.tag(6, WireType.LengthDelimited).fork(), options).join();
        let u = options.writeUnknownFields;
        if (u !== false)
            (u == true ? UnknownFieldHandler.onWrite : u)(this.typeName, message, writer);
        return writer;
    }
}
/**
 * @generated MessageType for protobuf message v1.Snippet
 */
export const Snippet = new Snippet$Type();
// @generated message type with reflection information, may provide speed optimized methods
class CreateSnippetRequest$Type extends MessageType<CreateSnippetRequest> {
    constructor() {
        super("v1.CreateSnippetRequest", [
            { no: 1, name: "name", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 2, name: "query", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 3, name: "params", kind: "message", repeat: 1 /*RepeatType.PACKED*/, T: () => QueryParam }
        ]);
    }
    create(value?: PartialMessage<CreateSnippetRequest>): CreateSnippetRequest {
        const message = { name: "", query: "", params: [] };
        globalThis.Object.defineProperty(message, MESSAGE_TYPE, { enumerable: false, value: this });
        if (value !== undefined)
            reflectionMergePartial<CreateSnippetRequest>(this, message, value);
        return message;
    }
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: CreateSnippetRequest): CreateSnippetRequest {
        let message = target ?? this.create(), end = reader.pos + length;
        while (reader.pos < end) {
            let [fieldNo, wireType] = reader.tag();
            switch (fieldNo) {
                case /* string name */ 1:
                    message.name = reader.string();
                    break;
                case /* string query */ 2:
                    message.query = reader.string();
                    break;
                case /* repeated v1.QueryParam params */ 3:
                    message.params.push(QueryParam.internalBinaryRead(reader, reader.uint32(), options));
                    break;
                default:
                    let u = options.readUnknownField;
                    if (u === "throw")
                        throw new globalThis.Error(`Unknown field ${fieldNo} (wire type ${wireType}) for ${this.typeName}`);
                    let d = reader.skip(wireType);
                    if (u !== false)
                        (u === true ? UnknownFieldHandler.onRead : u)(this.typeName, message, fieldNo, wireType, d);
            }
        }
        return message;
    }
    internalBinaryWrite(message: CreateSnippetRequest, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter {
        /* string name = 1; */
        if (message.name !== "")
            writer.tag(1, WireType.LengthDelimited).string(message.name);
        /* string query = 2; */
        if (message.query !== "")
            writer.tag(2, WireType.LengthDelimited).string(message.query);
        /* repeated v1.QueryParam params = 3; */
        for (let i = 0; i < message.params.length; i++)
            QueryParam.internalBinaryWrite(message.params[i], writer.tag(3, WireType.LengthDelimited).fork(), options).join();
        let u = options.writeUnknownFields;
        if (u !== false)
            (u == true ? UnknownFieldHandler.onWrite : u)(this.typeName, message, writer);
        return writer;
    }
}
/**
 * @generated MessageType for protobuf message v1.CreateSnippetRequest
 */
export const CreateSnippetRequest = new CreateSnippetRequest$Type();
// @generated message type with reflection information, may provide speed optimized methods
class UpdateSnippetRequest$Type extends MessageType<UpdateSnippetRequest> {
    constructor() {
        super("v1.UpdateSnippetRequest", [
            { no: 1, name: "id", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 2, name: "name", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 3, name: "query", kind: "scalar", T: 9 /*ScalarType.STRING*/ },
            { no: 4, name: "params", kind: "message", repeat: 1 /*RepeatType.PACKED*/, T: () => QueryParam }
        ]);
    }
    create(value?: PartialMessage<UpdateSnippetRequest>): UpdateSnippetRequest {
        const message = { id: "", name: "", query: "", params: [] };
        globalThis.Object.defineProperty(message, MESSAGE_TYPE, { enumerable: false, value: this });
        if (value !== undefined)
            reflectionMergePartial<UpdateSnippetRequest>(this, message, value);
        return message;
    }
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: UpdateSnippetRequest): UpdateSnippetRequest {
        let message = target ?? this.create(), end = reader.pos + length;
        while (reader.pos < end) {
            let [fieldNo, wireType] = reader.tag();
            switch (fieldNo) {
                case /* string id */ 1:
                    message.id = reader.string();
                    break;
                case /* string name */ 2:
                    message.name = reader.string();
                    break;
                case /* string query */ 3:
                    message.query = reader.string();
                    break;
                case /* repeated v1.QueryParam params */ 4:
                    message.params.push(QueryParam.internalBinaryRead(reader, reader.uint32(), options));
                    break;
                default:
                    let u = options.readUnknownField;
                    if (u === "throw")
                        throw new globalThis.Error(`Unknown field ${fieldNo} (wire type ${wireType}) for ${this.typeName}`);
                    let d = reader.skip(wireType);
                    if (u !== false)
                        (u === true ? UnknownFieldHandler.onRead : u)(this.typeName, message, fieldNo, wireType, d);
            }
        }
        return message;
    }
    internalBinaryWrite(message: UpdateSnippetRequest, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter {
        /* string id = 1; */
        if (message.id !== "")
            writer.tag(1, WireType.LengthDelimited).string(message.id);
        /* string name = 2; */
        if (message.name !== "")
            writer.tag(2, WireType.LengthDelimited).string(message.name);
        /* string query = 3; */
        if (message.query !== "")
            writer.tag(3, WireType.LengthDelimited).string(message.query);
        /* repeated v1.QueryParam params = 4; */
        for (let i = 0; i < message.params.length; i++)
            QueryParam.internalBinaryWrite(message.params[i], writer.tag(4, WireType.LengthDelimited).fork(), options).join();
        let u = options.writeUnknownFields;
        if (u !== false)
            (u == true ? UnknownFieldHandler.onWrite : u)(this.typeName, message, writer);
        return writer;
    }
}
/**
 * @generated MessageType for protobuf message v1.UpdateSnippetRequest
 */
export const UpdateSnippetRequest = new UpdateSnippetRequest$Type();
// @generated message type with reflection information, may provide speed optimized methods
class DeleteSnippetRequest$Type extends MessageType<DeleteSnippetRequest> {
    constructor() {
        super("v1.DeleteSnippetRequest", [
            { no: 1, name: "id", kind: "scalar", T: 9 /*ScalarType.STRING*/ }
        ]);
    }
    create(value?: PartialMessage<DeleteSnippetRequest>): DeleteSnippetRequest {
        const message = { id: "" };
        globalThis.Object.defineProperty(message, MESSAGE_TYPE, { enumerable: false, value: this });
        if (value !== undefined)
            reflectionMergePartial<DeleteSnippetRequest>(this, message, value);
        return message;
    }
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: DeleteSnippetRequest): DeleteSnippetRequest {
        let message = target ?? this.create(), end = reader.pos + length;
        while (reader.pos < end) {
            let [fieldNo, wireType] = reader.tag();
            switch (fieldNo) {
                case /* string id */ 1:
                    message.id = reader.string();
                    break;
                default:
                    let u = options.readUnknownField;
                    if (u === "throw")
                        throw new globalThis.Error(`Unknown field ${fieldNo} (wire type ${wireType}) for ${this.typeName}`);
                    let d = reader.skip(wireType);
                    if (u !== false)
                        (u === true ? UnknownFieldHandler.onRead : u)(this.typeName, message, fieldNo, wireType, d);
            }
        }
        return message;
    }
    internalBinaryWrite(message: DeleteSnippetRequest, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter {
        /* string id = 1; */
        if (message.id !== "")
            writer.tag(1, WireType.LengthDelimited).string(message.id);
        let u = options.writeUnknownFields;
        if (u !== false)
            (u == true ? UnknownFieldHandler.onWrite : u)(this.typeName, message, writer);
        return writer;
    }
}
/**
 * @generated MessageType for protobuf message v1.DeleteSnippetRequest
 */
export const DeleteSnippetRequest = new DeleteSnippetRequest$Type();
// @generated message type with reflection information, may provide speed optimized methods
class ListSnippetsRequest$Type extends MessageType<ListSnippetsRequest> {
    constructor() {
        super("v1.ListSnippetsRequest", []);
    }
    create(value?: PartialMessage<ListSnippetsRequest>): ListSnippetsRequest {
        const message = {};
        globalThis.Object.defineProperty(message, MESSAGE_TYPE, { enumerable: false, value: this });
        if (value !== undefined)
            reflectionMergePartial<ListSnippetsRequest>(this, message, value);
        return message;
    }
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: ListSnippetsRequest): ListSnippetsRequest {
        return target ?? this.create();
    }
    internalBinaryWrite(message: ListSnippetsRequest, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter {
        let u = options.writeUnknownFields;
        if (u !== false)
            (u == true ? UnknownFieldHandler.onWrite : u)(this.typeName, message, writer);
        return writer;
    }
}
/**
 * @generated MessageType for protobuf message v1.ListSnippetsRequest
 */
export const ListSnippetsRequest = new ListSnippetsRequest$Type();
// @generated message type with reflection information, may provide speed optimized methods
class ListSnippetsResponse$Type extends MessageType<ListSnippetsResponse> {
    constructor() {
        super("v1.ListSnippetsResponse", [
            { no: 1, name: "snippets", kind: "message", repeat: 1 /*RepeatType.PACKED*/, T: () => Snippet }
        ]);
    }
    create(value?: PartialMessage<ListSnippetsResponse>): ListSnippetsResponse {
        const message = { snippets: [] };
        globalThis.Object.defineProperty(message, MESSAGE_TYPE, { enumerable: false, value: this });
        if (value !== undefined)
            reflectionMergePartial<ListSnippetsResponse>(this, message, value);
        return message;
    }
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: ListSnippetsResponse): ListSnippetsResponse {
        let message = target ?? this.create(), end = reader.pos + length;
        while (reader.pos < end) {
            let [fieldNo, wireType] = reader.tag();
            switch (fieldNo) {
                case /* repeated v1.Snippet snippets */ 1:
                    message.snippets.push(Snippet.internalBinaryRead(reader, reader.uint32(), options));
                    break;
                default:
                    let u = options.readUnknownField;
                    if (u === "throw")
                        throw new globalThis.Error(`Unknown field ${fieldNo} (wire type ${wireType}) for ${this.typeName}`);
                    let d = reader.skip(wireType);
                    if (u !== false)
                        (u === true ? UnknownFieldHandler.onRead : u)(this.typeName, message, fieldNo, wireType, d);
            }
        }
        return message;
    }
    internalBinaryWrite(message: ListSnippetsResponse, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter {
        /* repeated v1.Snippet snippets = 1; */
        for (let i = 0; i < message.snippets.length; i++)
            Snippet.internalBinaryWrite(message.snippets[i], writer.tag(1, WireType.LengthDelimited).fork(), options).join();
        let u = options.writeUnknownFields;
        if (u !== false)
            (u == true ? UnknownFieldHandler.onWrite : u)(this.typeName, message, writer);
        return writer;
    }
}
/**
 * @generated MessageType for protobuf message v1.ListSnippetsResponse
 */
export const ListSnippetsResponse = new ListSnippetsResponse$Type();
/**
 * @generated ServiceType for protobuf service v1.Snippets
 */
export const Snippets = new ServiceType("v1.Snippets", [
    { name: "CreateSnippet", options: { "google.api.http": { post: "/v1/snippets" } }, I: CreateSnippetRequest, O: Snippet },
    { name: "UpdateSnippet", options: { "google.api.http": { put: "/v1/snippets" } }, I: UpdateSnippetRequest, O: Snippet },
    { name: "ListSnippets", options: { "google.api.http": { get: "/v1/snippets" } }, I: ListSnippetsRequest, O: ListSnippetsResponse },
    { name: "DeteteSnippet", options: { "google.api.http": { delete: "/v1/snippets" } }, I: DeleteSnippetRequest, O: Empty }
]);
