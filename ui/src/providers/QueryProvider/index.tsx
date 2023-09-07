import { createContext, PropsWithChildren, useState, useContext, useEffect, useCallback } from "react"
import { QueryResponse } from "../../vince"
import { useEditor } from "../EditorProvider"
import { getQueryRequestFromEditor } from "../../scenes/Editor/Monaco/utils"
import { useVince } from "../VinceProvider"


type ContextProps = {
    running: boolean
    request: string
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
    toggleRun: () => undefined,
    stopRunning: () => undefined,
}

const QueryContext = createContext<ContextProps>(defaultValues)


export const QueryProvider = ({ children }: PropsWithChildren<{}>) => {
    const [running, setRunning] = useState<boolean>(false)
    const [request, setRequest] = useState<string>("")
    const [result, setResult] = useState<QueryResponse | undefined>(undefined)
    const { editorRef } = useEditor()
    const { queryClient } = useVince()
    useEffect(() => {
        if (running && request !== "") {
            queryClient?.query({
                query: request,
                params: [],
            }).then((result) => {
                setResult(result.response)
            })
                .catch((e) => {
                    console.log(e)
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
            value={{ running, toggleRun, stopRunning, request, result }}
        >
            {children}
        </QueryContext.Provider>
    )
}

export const useQuery = () => {
    return useContext(QueryContext)
}