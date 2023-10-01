import { createContext, PropsWithChildren, useState, useContext, useEffect, useCallback } from "react"
import { QueryResponse } from "../../vince"
import { useEditor } from "../EditorProvider"
import { getQueryRequestFromEditor } from "../../scenes/Editor/Monaco/utils"
import { useVince } from "../VinceProvider"
import { RpcError } from "@protobuf-ts/runtime-rpc"


type ContextProps = {
    running: boolean
    request: string
    error: RpcError | undefined
    result: QueryResponse | undefined
    toggleRun: () => void
    stopRunning: () => void
}

const emptyResult = {
    elapsed: "", columns: [], rows: []
}

const defaultValues = {
    running: false,
    request: "",
    result: undefined,
    error: undefined,
    toggleRun: () => undefined,
    stopRunning: () => undefined,
}

const QueryContext = createContext<ContextProps>(defaultValues)


export const QueryProvider = ({ children }: PropsWithChildren<{}>) => {
    const [running, setRunning] = useState<boolean>(false)
    const [request, setRequest] = useState<string>("")
    const [result, setResult] = useState<QueryResponse | undefined>(undefined)
    const [error, setErr] = useState<RpcError | undefined>(undefined)
    const { editorRef } = useEditor()
    const { queryClient } = useVince()
    useEffect(() => {
        if (running && request !== "") {
            setErr(undefined)
            queryClient?.query({
                query: request,
                params: [],
            }).then((result) => {
                setResult(result.response)
            })
                .catch((e: RpcError) => {
                    setErr(e)
                    setRunning(false)
                })
        }
    }, [running, queryClient, request])

    useEffect(() => {
        if (running && result) {
            setRunning(false)
        }
    }, [running, result])

    const toggleRun = () => {
        setRunning(!running)
        if (!running) {
            const query = getQueryRequestFromEditor(editorRef?.current!)
            if (query) {
                setRequest(query.query)
            }
        }
    }

    const stopRunning = () => {
        setRunning(false)
    }

    return (
        <QueryContext.Provider
            value={{
                running, toggleRun,
                stopRunning, request, result, error
            }}
        >
            {children}
        </QueryContext.Provider>
    )
}

export const useQuery = () => {
    return useContext(QueryContext)
}